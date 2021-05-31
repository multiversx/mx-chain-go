package state

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type pruningOperation byte

const (
	cancelPrune pruningOperation = 0
	prune       pruningOperation = 1
)

var numCheckpointsKey = []byte("state checkpoint")

type loadingMeasurements struct {
	sync.Mutex
	numCalls   uint64
	size       uint64
	duration   time.Duration
	identifier string
}

func (lm *loadingMeasurements) addMeasurement(size int, duration time.Duration) {
	lm.Lock()
	lm.numCalls++
	lm.size += uint64(size)
	lm.duration += duration
	lm.Unlock()
}

func (lm *loadingMeasurements) resetAndPrint() {
	lm.Lock()
	numCalls := lm.numCalls
	lm.numCalls = 0

	size := lm.size
	lm.size = 0

	duration := lm.duration
	lm.duration = 0
	lm.Unlock()

	log.Debug(lm.identifier+" time measurements",
		"num calls", numCalls,
		"total size in bytes", size,
		"total cumulated duration", duration,
	)
}

// AccountsDB is the struct used for accessing accounts. This struct is concurrent safe.
type AccountsDB struct {
	mainTrie               data.Trie
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	accountFactory         AccountFactory
	dbEvictionWaitingList  data.DBRemoveCacher
	pruningBuffer          atomicBuffer
	obsoleteDataTrieHashes map[string][][]byte

	lastRootHash []byte
	dataTries    TriesHolder
	entries      []JournalEntry
	mutOp        sync.RWMutex

	numCheckpoints       uint32
	loadCodeMeasurements *loadingMeasurements
	stopWatch            *core.StopWatch
}

var log = logger.GetOrCreate("state")

// NewAccountsDB creates a new account manager
func NewAccountsDB(
	trie data.Trie,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory AccountFactory,
	ewl data.DBRemoveCacher,
	pruningBufferLen uint32,
) (*AccountsDB, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(accountFactory) {
		return nil, ErrNilAccountFactory
	}
	if check.IfNil(ewl) {
		return nil, ErrNilEvictionWaitingList
	}

	numCheckpoints := getNumCheckpoints(trie.GetStorageManager())
	return &AccountsDB{
		mainTrie:               trie,
		hasher:                 hasher,
		marshalizer:            marshalizer,
		accountFactory:         accountFactory,
		dbEvictionWaitingList:  ewl,
		pruningBuffer:          newPruningBuffer(pruningBufferLen),
		entries:                make([]JournalEntry, 0),
		mutOp:                  sync.RWMutex{},
		dataTries:              NewDataTriesHolder(),
		obsoleteDataTrieHashes: make(map[string][][]byte),
		numCheckpoints:         numCheckpoints,
		stopWatch:              core.NewStopWatch(),
		loadCodeMeasurements: &loadingMeasurements{
			identifier: "load code",
		},
	}, nil
}

func getNumCheckpoints(trieStorageManager data.StorageManager) uint32 {
	val, err := trieStorageManager.Database().Get(numCheckpointsKey)
	if err != nil {
		return 0
	}

	bytesForUint32 := 4
	if len(val) < bytesForUint32 {
		return 0
	}

	return binary.BigEndian.Uint32(val)
}

//GetCode returns the code for the given account
func (adb *AccountsDB) GetCode(codeHash []byte) []byte {
	if len(codeHash) == 0 {
		return nil
	}

	var codeEntry CodeEntry

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		adb.loadCodeMeasurements.addMeasurement(len(codeEntry.Code), duration)
	}()

	val, err := adb.mainTrie.Get(codeHash)
	if err != nil {
		return nil
	}

	err = adb.marshalizer.Unmarshal(&codeEntry, val)
	if err != nil {
		return nil
	}

	return codeEntry.Code
}

// ImportAccount saves the account in the trie. It does not modify
func (adb *AccountsDB) ImportAccount(account AccountHandler) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(account) {
		return fmt.Errorf("%w in accountsDB ImportAccount", ErrNilAccountHandler)
	}

	return adb.saveAccountToTrie(account)
}

// SaveAccount saves in the trie all changes made to the account.
func (adb *AccountsDB) SaveAccount(account AccountHandler) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(account) {
		return fmt.Errorf("%w in accountsDB SaveAccount", ErrNilAccountHandler)
	}

	oldAccount, err := adb.getAccount(account.AddressBytes())
	if err != nil {
		return err
	}

	var entry JournalEntry
	if check.IfNil(oldAccount) {
		entry, err = NewJournalEntryAccountCreation(account.AddressBytes(), adb.mainTrie)
		if err != nil {
			return err
		}
		adb.journalize(entry)
	} else {
		entry, err = NewJournalEntryAccount(oldAccount)
		if err != nil {
			return err
		}
		adb.journalize(entry)
	}

	err = adb.saveCodeAndDataTrie(oldAccount, account)
	if err != nil {
		return err
	}

	return adb.saveAccountToTrie(account)
}

func (adb *AccountsDB) saveCodeAndDataTrie(oldAcc, newAcc AccountHandler) error {
	baseNewAcc, newAccOk := newAcc.(baseAccountHandler)
	baseOldAccount, _ := oldAcc.(baseAccountHandler)

	if !newAccOk {
		return nil
	}

	err := adb.saveDataTrie(baseNewAcc)
	if err != nil {
		return err
	}

	return adb.saveCode(baseNewAcc, baseOldAccount)
}

func (adb *AccountsDB) saveCode(newAcc, oldAcc baseAccountHandler) error {
	// TODO when state splitting is implemented, check how the code should be copied in different shards

	if !newAcc.HasNewCode() {
		return nil
	}

	var oldCodeHash []byte
	if !check.IfNil(oldAcc) {
		oldCodeHash = oldAcc.GetCodeHash()
	}

	userAcc, ok := newAcc.(*userAccount)
	if !ok {
		return ErrWrongTypeAssertion
	}

	newCode := userAcc.code
	var newCodeHash []byte
	if len(newCode) != 0 {
		newCodeHash = adb.hasher.Compute(string(newCode))
	}

	if bytes.Equal(oldCodeHash, newCodeHash) {
		newAcc.SetCodeHash(newCodeHash)
		return nil
	}

	unmodifiedOldCodeEntry, err := adb.updateOldCodeEntry(oldCodeHash)
	if err != nil {
		return err
	}

	err = adb.updateNewCodeEntry(newCodeHash, newCode)
	if err != nil {
		return err
	}

	entry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, newCodeHash, adb.mainTrie, adb.marshalizer)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	newAcc.SetCodeHash(newCodeHash)
	return nil
}

func (adb *AccountsDB) updateOldCodeEntry(oldCodeHash []byte) (*CodeEntry, error) {
	oldCodeEntry, err := getCodeEntry(oldCodeHash, adb.mainTrie, adb.marshalizer)
	if err != nil {
		return nil, err
	}

	if oldCodeEntry == nil {
		return nil, nil
	}

	unmodifiedOldCodeEntry := &CodeEntry{
		Code:          oldCodeEntry.Code,
		NumReferences: oldCodeEntry.NumReferences,
	}

	if oldCodeEntry.NumReferences <= 1 {
		err = adb.mainTrie.Update(oldCodeHash, nil)
		if err != nil {
			return nil, err
		}

		return unmodifiedOldCodeEntry, nil
	}

	oldCodeEntry.NumReferences--
	err = saveCodeEntry(oldCodeHash, oldCodeEntry, adb.mainTrie, adb.marshalizer)
	if err != nil {
		return nil, err
	}

	return unmodifiedOldCodeEntry, nil
}

func (adb *AccountsDB) updateNewCodeEntry(newCodeHash []byte, newCode []byte) error {
	if len(newCode) == 0 {
		return nil
	}

	newCodeEntry, err := getCodeEntry(newCodeHash, adb.mainTrie, adb.marshalizer)
	if err != nil {
		return err
	}

	if newCodeEntry == nil {
		newCodeEntry = &CodeEntry{
			Code: newCode,
		}
	}
	newCodeEntry.NumReferences++

	err = saveCodeEntry(newCodeHash, newCodeEntry, adb.mainTrie, adb.marshalizer)
	if err != nil {
		return err
	}

	return nil
}

func getCodeEntry(codeHash []byte, trie Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	val, err := trie.Get(codeHash)
	if err != nil {
		return nil, err
	}

	if len(val) == 0 {
		return nil, nil
	}

	var codeEntry CodeEntry
	err = marshalizer.Unmarshal(&codeEntry, val)
	if err != nil {
		return nil, err
	}

	return &codeEntry, nil
}

func saveCodeEntry(codeHash []byte, entry *CodeEntry, trie Updater, marshalizer marshal.Marshalizer) error {
	codeEntry, err := marshalizer.Marshal(entry)
	if err != nil {
		return err
	}

	err = trie.Update(codeHash, codeEntry)
	if err != nil {
		return err
	}

	return nil
}

// LoadDataTrie retrieves and saves the SC data inside accountHandler object.
// Errors if something went wrong
func (adb *AccountsDB) loadDataTrie(accountHandler baseAccountHandler) error {
	if len(accountHandler.GetRootHash()) == 0 {
		return nil
	}

	dataTrie := adb.dataTries.Get(accountHandler.AddressBytes())
	if dataTrie != nil {
		accountHandler.SetDataTrie(dataTrie)
		return nil
	}

	dataTrie, err := adb.mainTrie.Recreate(accountHandler.GetRootHash())
	if err != nil {
		return NewErrMissingTrie(accountHandler.GetRootHash())
	}

	accountHandler.SetDataTrie(dataTrie)
	adb.dataTries.Put(accountHandler.AddressBytes(), dataTrie)
	return nil
}

// SaveDataTrie is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) saveDataTrie(accountHandler baseAccountHandler) error {
	if check.IfNil(accountHandler.DataTrieTracker()) {
		return ErrNilTrackableDataTrie
	}
	if len(accountHandler.DataTrieTracker().DirtyData()) == 0 {
		return nil
	}

	log.Trace("accountsDB.SaveDataTrie",
		"address", hex.EncodeToString(accountHandler.AddressBytes()),
		"nonce", accountHandler.GetNonce(),
	)

	if check.IfNil(accountHandler.DataTrie()) {
		newDataTrie, err := adb.mainTrie.Recreate(make([]byte, 0))
		if err != nil {
			return err
		}

		accountHandler.SetDataTrie(newDataTrie)
		adb.dataTries.Put(accountHandler.AddressBytes(), newDataTrie)
	}

	trackableDataTrie := accountHandler.DataTrieTracker()
	dataTrie := trackableDataTrie.DataTrie()
	oldValues := make(map[string][]byte)

	for k, v := range trackableDataTrie.DirtyData() {
		val, err := dataTrie.Get([]byte(k))
		if err != nil {
			return err
		}

		oldValues[k] = val

		err = dataTrie.Update([]byte(k), v)
		if err != nil {
			return err
		}
	}

	entry, err := NewJournalEntryDataTrieUpdates(oldValues, accountHandler)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	rootHash, err := trackableDataTrie.DataTrie().RootHash()
	if err != nil {
		return err
	}

	accountHandler.SetRootHash(rootHash)
	trackableDataTrie.ClearDataCaches()

	log.Trace("accountsDB.SaveDataTrie",
		"address", hex.EncodeToString(accountHandler.AddressBytes()),
		"new root hash", accountHandler.GetRootHash(),
	)

	return nil
}

func (adb *AccountsDB) saveAccountToTrie(accountHandler AccountHandler) error {
	log.Trace("accountsDB.saveAccountToTrie",
		"address", hex.EncodeToString(accountHandler.AddressBytes()),
	)

	//pass the reference to marshalizer, otherwise it will fail marshalizing balance
	buff, err := adb.marshalizer.Marshal(accountHandler)
	if err != nil {
		return err
	}

	return adb.mainTrie.Update(accountHandler.AddressBytes(), buff)
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(address []byte) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if len(address) == 0 {
		return fmt.Errorf("%w in RemoveAccount", ErrNilAddress)
	}

	acnt, err := adb.getAccount(address)
	if err != nil {
		return err
	}
	if check.IfNil(acnt) {
		return fmt.Errorf("%w in RemoveAccount for address %s", ErrAccNotFound, address)
	}

	entry, err := NewJournalEntryAccount(acnt)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	err = adb.removeCodeAndDataTrie(acnt)
	if err != nil {
		return err
	}

	log.Trace("accountsDB.RemoveAccount",
		"address", hex.EncodeToString(address),
	)

	return adb.mainTrie.Update(address, make([]byte, 0))
}

func (adb *AccountsDB) removeCodeAndDataTrie(acnt AccountHandler) error {
	baseAcc, ok := acnt.(baseAccountHandler)
	if !ok {
		return nil
	}

	err := adb.removeCode(baseAcc)
	if err != nil {
		return err
	}

	err = adb.removeDataTrie(baseAcc)
	if err != nil {
		return err
	}

	return nil
}

func (adb *AccountsDB) removeDataTrie(baseAcc baseAccountHandler) error {
	rootHash := baseAcc.GetRootHash()
	if len(rootHash) == 0 {
		return nil
	}

	dataTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return err
	}

	hashes, err := dataTrie.GetAllHashes()
	if err != nil {
		return err
	}

	adb.obsoleteDataTrieHashes[string(rootHash)] = hashes

	entry, err := NewJournalEntryDataTrieRemove(rootHash, adb.obsoleteDataTrieHashes)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	return nil
}

func (adb *AccountsDB) removeCode(baseAcc baseAccountHandler) error {
	oldCodeHash := baseAcc.GetCodeHash()
	unmodifiedOldCodeEntry, err := adb.updateOldCodeEntry(oldCodeHash)
	if err != nil {
		return err
	}

	codeChangeEntry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, nil, adb.mainTrie, adb.marshalizer)
	if err != nil {
		return err
	}
	adb.journalize(codeChangeEntry)

	return nil
}

// LoadAccount fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) LoadAccount(address []byte) (AccountHandler, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if len(address) == 0 {
		return nil, fmt.Errorf("%w in LoadAccount", ErrNilAddress)
	}

	log.Trace("accountsDB.LoadAccount",
		"address", hex.EncodeToString(address),
	)

	acnt, err := adb.getAccount(address)
	if err != nil {
		return nil, err
	}
	if check.IfNil(acnt) {
		return adb.accountFactory.CreateAccount(address)
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		err = adb.loadDataTrie(baseAcc)
		if err != nil {
			return nil, err
		}
	}

	return acnt, nil
}

func (adb *AccountsDB) getAccount(address []byte) (AccountHandler, error) {
	val, err := adb.mainTrie.Get(address)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}

	acnt, err := adb.accountFactory.CreateAccount(address)
	if err != nil {
		return nil, err
	}

	err = adb.marshalizer.Unmarshal(acnt, val)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// GetExistingAccount returns an existing account if exists or nil if missing
func (adb *AccountsDB) GetExistingAccount(address []byte) (AccountHandler, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if len(address) == 0 {
		return nil, fmt.Errorf("%w in GetExistingAccount", ErrNilAddress)
	}

	log.Trace("accountsDB.GetExistingAccount",
		"address", hex.EncodeToString(address),
	)

	acnt, err := adb.getAccount(address)
	if err != nil {
		return nil, err
	}
	if check.IfNil(acnt) {
		return nil, ErrAccNotFound
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		err = adb.loadDataTrie(baseAcc)
		if err != nil {
			return nil, err
		}
	}

	return acnt, nil
}

// loadCode retrieves and saves the SC code inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) loadCode(accountHandler baseAccountHandler) error {
	if len(accountHandler.GetCodeHash()) == 0 {
		return nil
	}

	val, err := adb.mainTrie.Get(accountHandler.GetCodeHash())
	if err != nil {
		return err
	}

	var codeEntry CodeEntry
	err = adb.marshalizer.Unmarshal(&codeEntry, val)
	if err != nil {
		return err
	}

	accountHandler.SetCode(codeEntry.Code)
	return nil
}

// RevertToSnapshot apply Revert method over accounts object and removes entries from the list
// Calling with 0 will revert everything. If the snapshot value is out of bounds, an err will be returned
func (adb *AccountsDB) RevertToSnapshot(snapshot int) error {
	log.Trace("accountsDB.RevertToSnapshot started",
		"snapshot", snapshot,
	)

	adb.mutOp.Lock()

	defer func() {
		log.Trace("accountsDB.RevertToSnapshot ended")
		adb.mutOp.Unlock()
	}()

	if snapshot > len(adb.entries) || snapshot < 0 {
		return ErrSnapshotValueOutOfBounds
	}

	if snapshot == 0 {
		log.Trace("revert snapshot to adb.lastRootHash", "hash", adb.lastRootHash)
		return adb.recreateTrie(adb.lastRootHash)
	}

	for i := len(adb.entries) - 1; i >= snapshot; i-- {
		account, err := adb.entries[i].Revert()
		if err != nil {
			return err
		}

		if !check.IfNil(account) {
			err = adb.saveAccountToTrie(account)
			if err != nil {
				return err
			}
		}
	}

	adb.entries = adb.entries[:snapshot]

	return nil
}

// JournalLen will return the number of entries
func (adb *AccountsDB) JournalLen() int {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	length := len(adb.entries)
	log.Trace("accountsDB.JournalLen",
		"length", length,
	)

	return length
}

// Commit will persist all data inside the trie
func (adb *AccountsDB) Commit() ([]byte, error) {
	adb.mutOp.Lock()
	defer func() {
		adb.mutOp.Unlock()
		adb.loadCodeMeasurements.resetAndPrint()
	}()

	log.Trace("accountsDB.Commit started")
	adb.entries = make([]JournalEntry, 0)

	oldHashes := make(data.ModifiedHashes)
	newHashes := make(data.ModifiedHashes)
	//Step 1. commit all data tries
	dataTries := adb.dataTries.GetAll()
	for i := 0; i < len(dataTries); i++ {
		err := adb.commitTrie(dataTries[i], oldHashes, newHashes)
		if err != nil {
			return nil, err
		}
	}
	adb.dataTries.Reset()

	oldRoot := adb.mainTrie.GetOldRoot()

	//Step 2. commit main trie
	err := adb.commitTrie(adb.mainTrie, oldHashes, newHashes)
	if err != nil {
		return nil, err
	}

	newRoot, err := adb.mainTrie.RootHash()
	if err != nil {
		log.Trace("accountsDB.Commit ended", "error", err.Error())
		return nil, err
	}

	if adb.mainTrie.GetStorageManager().IsPruningEnabled() {
		for _, hashes := range adb.obsoleteDataTrieHashes {
			for i := range hashes {
				oldHashes[string(hashes[i])] = struct{}{}
			}
		}

		err = adb.markForEviction(oldRoot, newRoot, oldHashes, newHashes)
		if err != nil {
			return nil, err
		}
	}

	adb.lastRootHash = newRoot
	adb.obsoleteDataTrieHashes = make(map[string][][]byte)

	log.Trace("accountsDB.Commit ended", "root hash", newRoot)

	return newRoot, nil
}

func (adb *AccountsDB) markForEviction(
	oldRoot []byte,
	newRoot []byte,
	oldHashes data.ModifiedHashes,
	newHashes data.ModifiedHashes,
) error {
	if bytes.Equal(newRoot, oldRoot) {
		log.Trace("old root and new root are identical", "rootHash", newRoot)
		return nil
	}

	log.Trace("trie hashes sizes", "newHashes", len(newHashes), "oldHashes", len(oldHashes))
	removeDuplicatedKeys(oldHashes, newHashes)

	if len(newHashes) > 0 && len(newRoot) > 0 {
		newRoot = append(newRoot, byte(data.NewRoot))
		err := adb.dbEvictionWaitingList.Put(newRoot, newHashes)
		if err != nil {
			return err
		}

		logMapWithTrace("MarkForEviction newHashes", "hash", newHashes)
	}

	if len(oldHashes) > 0 && len(oldRoot) > 0 {
		oldRoot = append(oldRoot, byte(data.OldRoot))
		err := adb.dbEvictionWaitingList.Put(oldRoot, oldHashes)
		if err != nil {
			return err
		}

		logMapWithTrace("MarkForEviction oldHashes", "hash", oldHashes)
	}
	return nil
}

func removeDuplicatedKeys(oldHashes map[string]struct{}, newHashes map[string]struct{}) {
	for key := range oldHashes {
		_, ok := newHashes[key]
		if ok {
			delete(oldHashes, key)
			delete(newHashes, key)
			log.Trace("found in newHashes and oldHashes", "hash", key)
		}
	}
}

func logMapWithTrace(message string, paramName string, hashes data.ModifiedHashes) {
	if log.GetLevel() == logger.LogTrace {
		for key := range hashes {
			log.Trace(message, paramName, key)
		}
	}
}

func (adb *AccountsDB) commitTrie(tr data.Trie, oldHashes data.ModifiedHashes, newHashes data.ModifiedHashes) error {
	if adb.mainTrie.GetStorageManager().IsPruningEnabled() {
		oldTrieHashes := tr.GetObsoleteHashes()
		newTrieHashes, err := tr.GetDirtyHashes()
		if err != nil {
			return err
		}

		for _, hash := range oldTrieHashes {
			oldHashes[string(hash)] = struct{}{}
		}

		for hash := range newTrieHashes {
			newHashes[hash] = struct{}{}
		}
	}

	return tr.Commit()
}

// RootHash returns the main trie's root hash
func (adb *AccountsDB) RootHash() ([]byte, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	rootHash, err := adb.mainTrie.RootHash()
	log.Trace("accountsDB.RootHash",
		"root hash", rootHash,
		"err", err,
	)

	return rootHash, err
}

// RecreateTrie is used to reload the trie based on an existing rootHash
func (adb *AccountsDB) RecreateTrie(rootHash []byte) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	err := adb.recreateTrie(rootHash)
	if err != nil {
		return err
	}
	adb.lastRootHash = rootHash

	return nil
}

func (adb *AccountsDB) recreateTrie(rootHash []byte) error {
	log.Trace("accountsDB.RecreateTrie", "root hash", rootHash)
	defer func() {
		log.Trace("accountsDB.RecreateTrie ended")
	}()

	adb.obsoleteDataTrieHashes = make(map[string][][]byte)
	adb.dataTries.Reset()
	adb.entries = make([]JournalEntry, 0)
	newTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return err
	}
	if check.IfNil(newTrie) {
		return ErrNilTrie
	}

	adb.mainTrie = newTrie
	return nil
}

// RecreateAllTries recreates all the tries from the accounts DB
func (adb *AccountsDB) RecreateAllTries(rootHash []byte, ctx context.Context) (map[string]data.Trie, error) {
	leavesChannel, err := adb.mainTrie.GetAllLeavesOnChannel(rootHash, ctx)
	if err != nil {
		return nil, err
	}

	recreatedTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return nil, err
	}

	allTries := make(map[string]data.Trie)
	allTries[string(rootHash)] = recreatedTrie

	for leaf := range leavesChannel {
		account := &userAccount{}
		err = adb.marshalizer.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) > 0 {
			dataTrie, errRecreate := adb.mainTrie.Recreate(account.RootHash)
			if errRecreate != nil {
				return nil, errRecreate
			}

			allTries[string(account.RootHash)] = dataTrie
		}
	}

	return allTries, nil
}

// Journalize adds a new object to entries list.
func (adb *AccountsDB) journalize(entry JournalEntry) {
	if check.IfNil(entry) {
		return
	}

	adb.entries = append(adb.entries, entry)
	log.Trace("accountsDB.Journalize", "new length", len(adb.entries))
}

// PruneTrie removes old values from the trie database
func (adb *AccountsDB) PruneTrie(rootHash []byte, identifier data.TriePruningIdentifier) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.PruneTrie", "root hash", rootHash)

	rootHash = append(rootHash, byte(identifier))

	if adb.mainTrie.GetStorageManager().IsPruningBlocked() {
		if identifier == data.NewRoot {
			adb.cancelPrune(rootHash)
			return
		}

		rootHash = append(rootHash, byte(prune))
		adb.pruningBuffer.add(rootHash)

		return
	}

	oldHashes := adb.pruningBuffer.removeAll()
	adb.resolveBufferedHashes(oldHashes)
	adb.prune(rootHash)
}

func (adb *AccountsDB) cancelPrune(rootHash []byte) {
	log.Trace("trie storage manager cancel prune", "root", rootHash)
	_, _ = adb.dbEvictionWaitingList.Evict(rootHash)
}

func (adb *AccountsDB) resolveBufferedHashes(oldHashes [][]byte) {
	for _, rootHash := range oldHashes {
		lastBytePos := len(rootHash) - 1
		if lastBytePos < 0 {
			continue
		}

		pruneOperation := pruningOperation(rootHash[lastBytePos])
		rootHash = rootHash[:lastBytePos]

		switch pruneOperation {
		case prune:
			adb.prune(rootHash)
		case cancelPrune:
			adb.cancelPrune(rootHash)
		default:
			log.Error("invalid pruning operation", "operation id", pruneOperation)
		}
	}
}

func (adb *AccountsDB) prune(rootHash []byte) {
	log.Trace("trie storage manager prune", "root", rootHash)

	err := adb.removeFromDb(rootHash)
	if err != nil {
		log.Error("trie storage manager remove from db", "error", err, "rootHash", hex.EncodeToString(rootHash))
	}
}

func (adb *AccountsDB) removeFromDb(rootHash []byte) error {
	hashes, err := adb.dbEvictionWaitingList.Evict(rootHash)
	if err != nil {
		return err
	}

	log.Debug("trie removeFromDb", "rootHash", rootHash)

	lastBytePos := len(rootHash) - 1
	if lastBytePos < 0 {
		return ErrInvalidIdentifier
	}
	identifier := data.TriePruningIdentifier(rootHash[lastBytePos])

	sw := core.NewStopWatch()
	sw.Start("removeFromDb")
	defer func() {
		sw.Stop("removeFromDb")
		log.Debug("trieStorageManager.removeFromDb", sw.GetMeasurements()...)
	}()

	for key := range hashes {
		shouldKeepHash, err := adb.dbEvictionWaitingList.ShouldKeepHash(key, identifier)
		if err != nil {
			return err
		}
		if shouldKeepHash {
			continue
		}

		hash := []byte(key)

		log.Trace("remove hash from trie db", "hash", hex.EncodeToString(hash))
		err = adb.mainTrie.GetStorageManager().Database().Remove(hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// CancelPrune clears the trie's evictionWaitingList
func (adb *AccountsDB) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.CancelPrune", "root hash", rootHash)

	rootHash = append(rootHash, byte(identifier))

	if adb.mainTrie.GetStorageManager().IsPruningBlocked() || adb.pruningBuffer.len() != 0 {
		rootHash = append(rootHash, byte(cancelPrune))
		adb.pruningBuffer.add(rootHash)

		return
	}

	adb.cancelPrune(rootHash)
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *AccountsDB) SnapshotState(rootHash []byte, ctx context.Context) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	trieStorageManager := adb.mainTrie.GetStorageManager()
	log.Trace("accountsDB.SnapshotState", "root hash", rootHash)
	trieStorageManager.EnterPruningBufferingMode()

	go func() {
		adb.stopWatch.Start("snapshotState")
		trieStorageManager.TakeSnapshot(rootHash)
		adb.snapshotUserAccountDataTrie(rootHash, ctx)
		adb.stopWatch.Stop("snapshotState")

		log.Debug("snapshotState", adb.stopWatch.GetMeasurements()...)
		trieStorageManager.ExitPruningBufferingMode()

		adb.increaseNumCheckpoints()
	}()
}

func (adb *AccountsDB) snapshotUserAccountDataTrie(rootHash []byte, ctx context.Context) {
	leavesChannel, err := adb.GetAllLeaves(rootHash, ctx)
	if err != nil {
		log.Error("incomplete snapshot as getAllLeaves error", "error", err)
		return
	}

	for leaf := range leavesChannel {
		account := &userAccount{}
		err = adb.marshalizer.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) > 0 {
			adb.mainTrie.GetStorageManager().SetCheckpoint(account.RootHash)
		}
	}
}

// SetStateCheckpoint sets a checkpoint for the state trie
func (adb *AccountsDB) SetStateCheckpoint(rootHash []byte, ctx context.Context) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	trieStorageManager := adb.mainTrie.GetStorageManager()
	log.Trace("accountsDB.SetStateCheckpoint", "root hash", rootHash)
	trieStorageManager.EnterPruningBufferingMode()

	go func() {
		adb.stopWatch.Start("setStateCheckpoint")
		trieStorageManager.SetCheckpoint(rootHash)
		adb.snapshotUserAccountDataTrie(rootHash, ctx)
		adb.stopWatch.Stop("setStateCheckpoint")

		log.Debug("setStateCheckpoint", adb.stopWatch.GetMeasurements()...)
		trieStorageManager.ExitPruningBufferingMode()

		adb.increaseNumCheckpoints()
	}()
}

func (adb *AccountsDB) increaseNumCheckpoints() {
	trieStorageManager := adb.mainTrie.GetStorageManager()
	time.Sleep(time.Duration(trieStorageManager.GetSnapshotDbBatchDelay()) * time.Second)
	atomic.AddUint32(&adb.numCheckpoints, 1)

	numCheckpoints := atomic.LoadUint32(&adb.numCheckpoints)
	numCheckpointsVal := make([]byte, 4)
	binary.BigEndian.PutUint32(numCheckpointsVal, numCheckpoints)

	err := trieStorageManager.Database().Put(numCheckpointsKey, numCheckpointsVal)
	if err != nil {
		log.Warn("could not add num checkpoints to database", "error", err)
	}
}

// IsPruningEnabled returns true if state pruning is enabled
func (adb *AccountsDB) IsPruningEnabled() bool {
	return adb.mainTrie.GetStorageManager().IsPruningEnabled()
}

// GetAllLeaves returns all the leaves from a given rootHash
func (adb *AccountsDB) GetAllLeaves(rootHash []byte, ctx context.Context) (chan core.KeyValueHolder, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	return adb.mainTrie.GetAllLeavesOnChannel(rootHash, ctx)
}

// GetNumCheckpoints returns the total number of state checkpoints
func (adb *AccountsDB) GetNumCheckpoints() uint32 {
	return atomic.LoadUint32(&adb.numCheckpoints)
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *AccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
