package state

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const (
	createNewDb       = true
	useExistingDb     = false
	leavesChannelSize = 100
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
	mainTrie               common.Trie
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	accountFactory         AccountFactory
	storagePruningManager  StoragePruningManager
	obsoleteDataTrieHashes map[string][][]byte

	lastRootHash []byte
	dataTries    TriesHolder
	entries      []JournalEntry
	mutOp        sync.RWMutex

	numCheckpoints       uint32
	loadCodeMeasurements *loadingMeasurements

	stackDebug []byte
}

var log = logger.GetOrCreate("state")

// NewAccountsDB creates a new account manager
func NewAccountsDB(
	trie common.Trie,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory AccountFactory,
	storagePruningManager StoragePruningManager,
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
	if check.IfNil(storagePruningManager) {
		return nil, ErrNilStoragePruningManager
	}

	numCheckpoints := getNumCheckpoints(trie.GetStorageManager())
	return &AccountsDB{
		mainTrie:               trie,
		hasher:                 hasher,
		marshalizer:            marshalizer,
		accountFactory:         accountFactory,
		storagePruningManager:  storagePruningManager,
		entries:                make([]JournalEntry, 0),
		mutOp:                  sync.RWMutex{},
		dataTries:              NewDataTriesHolder(),
		obsoleteDataTrieHashes: make(map[string][][]byte),
		numCheckpoints:         numCheckpoints,
		loadCodeMeasurements: &loadingMeasurements{
			identifier: "load code",
		},
	}, nil
}

func getNumCheckpoints(trieStorageManager common.StorageManager) uint32 {
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

// GetCode returns the code for the given account
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
func (adb *AccountsDB) ImportAccount(account vmcommon.AccountHandler) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(account) {
		return fmt.Errorf("%w in accountsDB ImportAccount", ErrNilAccountHandler)
	}

	return adb.saveAccountToTrie(account)
}

// SaveAccount saves in the trie all changes made to the account.
func (adb *AccountsDB) SaveAccount(account vmcommon.AccountHandler) error {
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

func (adb *AccountsDB) saveCodeAndDataTrie(oldAcc, newAcc vmcommon.AccountHandler) error {
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

	if check.IfNil(adb.dataTries.Get(accountHandler.AddressBytes())) {
		adb.dataTries.Put(accountHandler.AddressBytes(), accountHandler.DataTrie())
	}

	return nil
}

func (adb *AccountsDB) saveAccountToTrie(accountHandler vmcommon.AccountHandler) error {
	log.Trace("accountsDB.saveAccountToTrie",
		"address", hex.EncodeToString(accountHandler.AddressBytes()),
	)

	// pass the reference to marshalizer, otherwise it will fail marshalizing balance
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

func (adb *AccountsDB) removeCodeAndDataTrie(acnt vmcommon.AccountHandler) error {
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
func (adb *AccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
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

func (adb *AccountsDB) getAccount(address []byte) (vmcommon.AccountHandler, error) {
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
func (adb *AccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
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

// GetAccountFromBytes returns an account from the given bytes
func (adb *AccountsDB) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if len(address) == 0 {
		return nil, fmt.Errorf("%w in GetAccountFromBytes", ErrNilAddress)
	}

	acnt, err := adb.accountFactory.CreateAccount(address)
	if err != nil {
		return nil, err
	}

	err = adb.marshalizer.Unmarshal(acnt, accountBytes)
	if err != nil {
		return nil, err
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if !ok {
		return acnt, nil
	}

	err = adb.loadDataTrie(baseAcc)
	if err != nil {
		return nil, err
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

	oldHashes := make(common.ModifiedHashes)
	newHashes := make(common.ModifiedHashes)
	// Step 1. commit all data tries
	dataTries := adb.dataTries.GetAll()
	for i := 0; i < len(dataTries); i++ {
		err := adb.commitTrie(dataTries[i], oldHashes, newHashes)
		if err != nil {
			return nil, err
		}
	}
	adb.dataTries.Reset()

	oldRoot := adb.mainTrie.GetOldRoot()

	// Step 2. commit main trie
	err := adb.commitTrie(adb.mainTrie, oldHashes, newHashes)
	if err != nil {
		return nil, err
	}

	newRoot, err := adb.mainTrie.RootHash()
	if err != nil {
		log.Trace("accountsDB.Commit ended", "error", err.Error())
		return nil, err
	}

	err = adb.markForEviction(oldRoot, newRoot, oldHashes, newHashes)
	if err != nil {
		return nil, err
	}

	adb.lastRootHash = newRoot
	adb.obsoleteDataTrieHashes = make(map[string][][]byte)
	shouldCreateCheckpoint := adb.mainTrie.GetStorageManager().AddDirtyCheckpointHashes(newRoot, newHashes.Clone())

	if shouldCreateCheckpoint {
		log.Debug("checkpoint hashes holder is full - force state checkpoint")
		adb.setStateCheckpoint(newRoot)
	}

	log.Trace("accountsDB.Commit ended", "root hash", newRoot)

	return newRoot, nil
}

func (adb *AccountsDB) markForEviction(
	oldRoot []byte,
	newRoot []byte,
	oldHashes common.ModifiedHashes,
	newHashes common.ModifiedHashes,
) error {
	if !adb.mainTrie.GetStorageManager().IsPruningEnabled() {
		return nil
	}

	for _, hashes := range adb.obsoleteDataTrieHashes {
		for _, hash := range hashes {
			oldHashes[string(hash)] = struct{}{}
		}
	}

	return adb.storagePruningManager.MarkForEviction(oldRoot, newRoot, oldHashes, newHashes)
}

func (adb *AccountsDB) commitTrie(tr common.Trie, oldHashes common.ModifiedHashes, newHashes common.ModifiedHashes) error {
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
func (adb *AccountsDB) RecreateAllTries(rootHash []byte) (map[string]common.Trie, error) {
	leavesChannel, err := adb.mainTrie.GetAllLeavesOnChannel(rootHash)
	if err != nil {
		return nil, err
	}

	recreatedTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return nil, err
	}

	allTries := make(map[string]common.Trie)
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

// GetTrie returns the trie that has the given rootHash
func (adb *AccountsDB) GetTrie(rootHash []byte) (common.Trie, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	return adb.mainTrie.Recreate(rootHash)
}

// Journalize adds a new object to entries list.
func (adb *AccountsDB) journalize(entry JournalEntry) {
	if check.IfNil(entry) {
		return
	}

	adb.entries = append(adb.entries, entry)
	log.Trace("accountsDB.Journalize", "new length", len(adb.entries))

	if len(adb.entries) == 1 {
		adb.stackDebug = debug.Stack()
	}
}

// GetStackDebugFirstEntry will return the debug.Stack for the first entry from the adb.entries
func (adb *AccountsDB) GetStackDebugFirstEntry() []byte {
	adb.mutOp.RLock()
	defer adb.mutOp.RUnlock()

	return adb.stackDebug
}

// PruneTrie removes old values from the trie database
func (adb *AccountsDB) PruneTrie(rootHash []byte, identifier TriePruningIdentifier) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.PruneTrie", "root hash", rootHash)

	adb.storagePruningManager.PruneTrie(rootHash, identifier, adb.mainTrie.GetStorageManager())
}

// CancelPrune clears the trie's evictionWaitingList
func (adb *AccountsDB) CancelPrune(rootHash []byte, identifier TriePruningIdentifier) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.CancelPrune", "root hash", rootHash)

	adb.storagePruningManager.CancelPrune(rootHash, identifier, adb.mainTrie.GetStorageManager())
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *AccountsDB) SnapshotState(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	trieStorageManager := adb.mainTrie.GetStorageManager()
	log.Trace("accountsDB.SnapshotState", "root hash", rootHash)
	trieStorageManager.EnterPruningBufferingMode()

	go func() {
		stopWatch := core.NewStopWatch()
		stopWatch.Start("snapshotState")
		leavesChannel := make(chan core.KeyValueHolder, leavesChannelSize)
		trieStorageManager.TakeSnapshot(rootHash, createNewDb, leavesChannel)
		adb.snapshotUserAccountDataTrie(true, leavesChannel)
		stopWatch.Stop("snapshotState")

		log.Debug("snapshotState", stopWatch.GetMeasurements()...)
		trieStorageManager.ExitPruningBufferingMode()

		adb.increaseNumCheckpoints()
	}()
}

func (adb *AccountsDB) snapshotUserAccountDataTrie(isSnapshot bool, leavesChannel chan core.KeyValueHolder) {
	for leaf := range leavesChannel {
		account := &userAccount{}
		err := adb.marshalizer.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) == 0 {
			continue
		}

		if isSnapshot {
			adb.mainTrie.GetStorageManager().TakeSnapshot(account.RootHash, useExistingDb, nil)
			continue
		}

		adb.mainTrie.GetStorageManager().SetCheckpoint(account.RootHash, nil)
	}
}

// SetStateCheckpoint sets a checkpoint for the state trie
func (adb *AccountsDB) SetStateCheckpoint(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	adb.setStateCheckpoint(rootHash)
}

func (adb *AccountsDB) setStateCheckpoint(rootHash []byte) {
	trieStorageManager := adb.mainTrie.GetStorageManager()
	log.Trace("accountsDB.SetStateCheckpoint", "root hash", rootHash)
	trieStorageManager.EnterPruningBufferingMode()

	go func() {
		stopWatch := core.NewStopWatch()
		stopWatch.Start("setStateCheckpoint")
		leavesChannel := make(chan core.KeyValueHolder, leavesChannelSize)
		trieStorageManager.SetCheckpoint(rootHash, leavesChannel)
		adb.snapshotUserAccountDataTrie(false, leavesChannel)
		stopWatch.Stop("setStateCheckpoint")

		log.Debug("setStateCheckpoint", stopWatch.GetMeasurements()...)
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
func (adb *AccountsDB) GetAllLeaves(rootHash []byte) (chan core.KeyValueHolder, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	return adb.mainTrie.GetAllLeavesOnChannel(rootHash)
}

// GetNumCheckpoints returns the total number of state checkpoints
func (adb *AccountsDB) GetNumCheckpoints() uint32 {
	return atomic.LoadUint32(&adb.numCheckpoints)
}

// Close will handle the closing of the underlying components
func (adb *AccountsDB) Close() error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	_ = adb.mainTrie.Close()
	return adb.storagePruningManager.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *AccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
