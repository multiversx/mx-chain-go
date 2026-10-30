//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. codeEntry.proto
package state

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/stateAnalysis"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

const (
	leavesChannelSize       = 100
	missingNodesChannelSize = 100
)

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

type snapshotInfo struct {
	rootHash []byte
	epoch    uint32
}

// AccountsDB is the struct used for accessing accounts. This struct is concurrent safe.
type AccountsDB struct {
	mainTrie               common.Trie
	hasher                 hashing.Hasher
	marshaller             marshal.Marshalizer
	accountFactory         AccountFactory
	storagePruningManager  StoragePruningManager
	obsoleteDataTrieHashes map[string][][]byte
	snapshotsManger        SnapshotsManager

	lastRootHash           []byte
	dataTries              common.TriesHolder
	entries                []JournalEntry
	stateAccessesCollector StateAccessesCollector
	mutOp                  sync.RWMutex
	loadCodeMeasurements   *loadingMeasurements
	addressConverter       core.PubkeyConverter
	pruningEnabled         bool
	analysisAdapter        int

	stackDebug []byte
}

var log = logger.GetOrCreate("state")

// ArgsAccountsDB is the arguments DTO for the AccountsDB instance
type ArgsAccountsDB struct {
	Trie                   common.Trie
	Hasher                 hashing.Hasher
	Marshaller             marshal.Marshalizer
	AccountFactory         AccountFactory
	StoragePruningManager  StoragePruningManager
	AddressConverter       core.PubkeyConverter
	SnapshotsManager       SnapshotsManager
	StateAccessesCollector StateAccessesCollector
	PruningEnabled         bool
}

// NewAccountsDB creates a new account manager
func NewAccountsDB(args ArgsAccountsDB) (*AccountsDB, error) {
	err := checkArgsAccountsDB(args)
	if err != nil {
		return nil, err
	}

	return createAccountsDb(args), nil
}

func createAccountsDb(args ArgsAccountsDB) *AccountsDB {
	return &AccountsDB{
		mainTrie:               args.Trie,
		hasher:                 args.Hasher,
		marshaller:             args.Marshaller,
		accountFactory:         args.AccountFactory,
		storagePruningManager:  args.StoragePruningManager,
		entries:                make([]JournalEntry, 0),
		mutOp:                  sync.RWMutex{},
		dataTries:              NewDataTriesHolder(),
		obsoleteDataTrieHashes: make(map[string][][]byte),
		loadCodeMeasurements: &loadingMeasurements{
			identifier: "load code",
		},
		addressConverter:       args.AddressConverter,
		snapshotsManger:        args.SnapshotsManager,
		stateAccessesCollector: args.StateAccessesCollector,
		pruningEnabled:         args.PruningEnabled,
	}
}

func checkArgsAccountsDB(args ArgsAccountsDB) error {
	if check.IfNil(args.Trie) {
		return ErrNilTrie
	}
	if check.IfNil(args.Hasher) {
		return ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return ErrNilMarshalizer
	}
	if check.IfNil(args.AccountFactory) {
		return ErrNilAccountFactory
	}
	if check.IfNil(args.StoragePruningManager) {
		return ErrNilStoragePruningManager
	}
	if check.IfNil(args.AddressConverter) {
		return ErrNilAddressConverter
	}
	if check.IfNil(args.SnapshotsManager) {
		return ErrNilSnapshotsManager
	}
	if check.IfNil(args.StateAccessesCollector) {
		return ErrNilStateAccessesCollector
	}

	return nil
}

func handleLoggingWhenError(message string, err error, extraArguments ...interface{}) {
	if err == nil {
		return
	}
	if core.IsClosingError(err) {
		args := []interface{}{"reason", err}
		log.Debug(message, append(args, extraArguments...)...)
		return
	}

	args := []interface{}{"error", err}
	log.Warn(message, append(args, extraArguments...)...)
}

// SetSyncer sets the given syncer as the syncer for the underlying trie
func (adb *AccountsDB) SetSyncer(syncer AccountsDBSyncer) error {
	return adb.snapshotsManger.SetSyncer(syncer)
}

// StartSnapshotIfNeeded starts the snapshot if the previous snapshot process was not fully completed
func (adb *AccountsDB) StartSnapshotIfNeeded() error {
	return adb.snapshotsManger.StartSnapshotAfterRestartIfNeeded(adb.getMainTrie().GetStorageManager())
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

	if stateAnalysis.Enabled() {
		stateAnalysis.Inc(stateAnalysis.CAdbGetCode.Adapter(adb.analysisAdapter))
		defer func() {
			stateAnalysis.Add(stateAnalysis.CAdbGetCodeBytes.Adapter(adb.analysisAdapter), int64(len(codeEntry.Code)))
			stateAnalysis.AddSince(stateAnalysis.CAdbGetCodeNs.Adapter(adb.analysisAdapter), startTime)
		}()
	}

	val, _, err := adb.getMainTrie().Get(codeHash)
	if err != nil {
		return nil
	}

	sc := &stateChange.StateAccess{
		Type:            stateChange.Read,
		MainTrieKey:     codeHash,
		MainTrieVal:     val,
		Operation:       stateChange.GetCode,
		DataTrieChanges: nil,
	}
	adb.stateAccessesCollector.AddStateAccess(sc)

	err = adb.marshaller.Unmarshal(&codeEntry, val)
	if err != nil {
		return nil
	}

	return codeEntry.Code
}

// ImportAccount saves the account in the trie. It does not modify
func (adb *AccountsDB) ImportAccount(account vmcommon.AccountHandler) error {
	if check.IfNil(account) {
		return fmt.Errorf("%w in accountsDB ImportAccount", ErrNilAccountHandler)
	}

	mainTrie := adb.getMainTrie()
	_, err := adb.saveAccountToTrie(account, mainTrie)
	return err
}

func (adb *AccountsDB) getMainTrie() common.Trie {
	adb.mutOp.RLock()
	defer adb.mutOp.RUnlock()

	return adb.mainTrie
}

// SaveAccount saves in the trie all changes made to the account.
func (adb *AccountsDB) SaveAccount(account vmcommon.AccountHandler) error {
	if check.IfNil(account) {
		return fmt.Errorf("%w in accountsDB SaveAccount", ErrNilAccountHandler)
	}

	// this is a critical section, do not remove the mutex
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	saveStart := stateAnalysis.Now()
	if stateAnalysis.Enabled() {
		stateAnalysis.Inc(stateAnalysis.CAdbSaveAccount.Adapter(adb.analysisAdapter))
		stateAnalysis.TouchAccount(adb.analysisAdapter, account.AddressBytes(), 0, true)
		defer func() {
			saveNs := int64(time.Since(saveStart))
			stateAnalysis.Add(stateAnalysis.CAdbSaveNs.Adapter(adb.analysisAdapter), saveNs)
			stateAnalysis.Max(stateAnalysis.MAdbSaveNs, saveNs)
		}()
	}

	oldAccount, oldAccountBytes, err := adb.getAccountAndBytesWithTouch(account.AddressBytes(), adb.mainTrie, false)
	stateAnalysis.AddSince(stateAnalysis.CAdbSaveRereadNs.Adapter(adb.analysisAdapter), saveStart)
	if err != nil {
		return err
	}

	var entry JournalEntry
	if check.IfNil(oldAccount) {
		stateAnalysis.Inc(stateAnalysis.CAdbSaveNewAccount.Adapter(adb.analysisAdapter))
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

	newDataTrieValues, err := adb.saveCodeAndDataTrie(oldAccount, account)
	if err != nil {
		return err
	}

	marshalledAccount, err := adb.saveAccountToTrie(account, adb.mainTrie)
	if err != nil {
		return err
	}
	if stateAnalysis.Enabled() && bytes.Equal(oldAccountBytes, marshalledAccount) {
		stateAnalysis.Inc(stateAnalysis.CAdbSaveUnchanged.Adapter(adb.analysisAdapter))
	}

	sc := &stateChange.StateAccess{
		Type:            stateChange.Write,
		MainTrieKey:     account.AddressBytes(),
		MainTrieVal:     marshalledAccount,
		DataTrieChanges: newDataTrieValues,
		Operation:       stateChange.SaveAccount,
		Index:           int32(len(adb.entries)),
		AccountChanges:  adb.stateAccessesCollector.GetAccountChanges(oldAccount, account),
	}

	adb.stateAccessesCollector.AddStateAccess(sc)

	return nil
}

func (adb *AccountsDB) saveCodeAndDataTrie(oldAcc, newAcc vmcommon.AccountHandler) ([]*stateChange.DataTrieChange, error) {
	baseNewAcc, newAccOk := newAcc.(baseAccountHandler)
	baseOldAccount, _ := oldAcc.(baseAccountHandler)

	if !newAccOk {
		return nil, nil
	}

	newValues, err := adb.saveDataTrie(baseNewAcc)
	if err != nil {
		return nil, err
	}

	err = adb.saveCode(baseNewAcc, baseOldAccount)
	if err != nil {
		return nil, err
	}

	return newValues, nil
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

	newCode := newAcc.GetCode()
	var newCodeHash []byte
	if len(newCode) != 0 {
		newCodeHash = adb.hasher.Compute(string(newCode))
	}

	if bytes.Equal(oldCodeHash, newCodeHash) {
		newAcc.SetCodeHash(newCodeHash)
		return nil
	}

	oldCodeEntry, err := adb.getCodeEntry(oldCodeHash)
	if err != nil {
		return err
	}

	newCodeEntry, err := adb.getCodeEntry(newCodeHash)
	if err != nil {
		return err
	}

	entry, err := NewJournalEntryCode(oldCodeEntry, oldCodeHash, newCodeEntry, newCodeHash, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	err = adb.updateOldCodeEntry(oldCodeHash, oldCodeEntry)
	if err != nil {
		return err
	}

	err = adb.updateNewCodeEntry(newCodeHash, newCodeEntry, newCode)
	if err != nil {
		return err
	}

	newAcc.SetCodeHash(newCodeHash)
	return nil
}

func (adb *AccountsDB) getCodeEntry(hash []byte) (*CodeEntry, error) {
	stateAnalysis.Inc(stateAnalysis.CAdbCodeEntryGet.Adapter(adb.analysisAdapter))
	codeEntry, err := getCodeEntry(hash, adb.mainTrie, adb.marshaller)
	if err != nil {
		return nil, err
	}

	sc := &stateChange.StateAccess{
		Type:            stateChange.Read,
		MainTrieKey:     hash,
		MainTrieVal:     nil,
		Operation:       stateChange.GetCode,
		DataTrieChanges: nil,
	}
	adb.stateAccessesCollector.AddStateAccess(sc)

	return codeEntry, nil
}

func (adb *AccountsDB) updateOldCodeEntry(oldCodeHash []byte, oldCodeEntry *CodeEntry) error {
	if oldCodeEntry == nil {
		return nil
	}

	codeEntryClone := &CodeEntry{
		Code:          oldCodeEntry.Code,
		NumReferences: oldCodeEntry.NumReferences,
	}

	if codeEntryClone.NumReferences <= 1 {
		err := adb.mainTrie.Delete(oldCodeHash)
		if err != nil {
			return err
		}

		sc1 := &stateChange.StateAccess{
			Type:            stateChange.Write,
			MainTrieKey:     oldCodeHash,
			MainTrieVal:     nil,
			Operation:       stateChange.WriteCode,
			DataTrieChanges: nil,
		}
		adb.stateAccessesCollector.AddStateAccess(sc1)

		return nil
	}

	codeEntryClone.NumReferences--
	stateAnalysis.Inc(stateAnalysis.CAdbCodeEntrySave.Adapter(adb.analysisAdapter))
	codeEntryBytes, err := saveCodeEntry(oldCodeHash, codeEntryClone, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}

	sc := &stateChange.StateAccess{
		Type:            stateChange.Write,
		MainTrieKey:     oldCodeHash,
		MainTrieVal:     codeEntryBytes,
		Operation:       stateChange.WriteCode,
		DataTrieChanges: nil,
	}
	adb.stateAccessesCollector.AddStateAccess(sc)

	return nil
}

func (adb *AccountsDB) updateNewCodeEntry(newCodeHash []byte, newCodeEntry *CodeEntry, newCode []byte) error {
	if len(newCode) == 0 {
		return nil
	}

	codeEntry := &CodeEntry{}
	codeEntry.Code = newCode
	if newCodeEntry != nil {
		codeEntry.NumReferences = newCodeEntry.NumReferences
	}
	codeEntry.NumReferences++

	stateAnalysis.Inc(stateAnalysis.CAdbCodeEntrySave.Adapter(adb.analysisAdapter))
	codeEntryBytes, err := saveCodeEntry(newCodeHash, codeEntry, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}

	sc := &stateChange.StateAccess{
		Type:            stateChange.Write,
		MainTrieKey:     newCodeHash,
		MainTrieVal:     codeEntryBytes,
		Operation:       stateChange.WriteCode,
		DataTrieChanges: nil,
	}
	adb.stateAccessesCollector.AddStateAccess(sc)

	return nil
}

func getCodeEntry(codeHash []byte, trie Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	val, _, err := trie.Get(codeHash)
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

func saveCodeEntry(codeHash []byte, entry *CodeEntry, trie Updater, marshalizer marshal.Marshalizer) ([]byte, error) {
	codeEntry, err := marshalizer.Marshal(entry)
	if err != nil {
		return nil, err
	}

	err = trie.Update(codeHash, codeEntry)
	if err != nil {
		return nil, err
	}

	return codeEntry, nil
}

// loadDataTrieConcurrentSafe retrieves and saves the SC data inside accountHandler object.
// Errors if something went wrong
func (adb *AccountsDB) loadDataTrieConcurrentSafe(accountHandler baseAccountHandler, mainTrie common.Trie) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if len(accountHandler.GetRootHash()) != 0 {
		stateAnalysis.Inc(stateAnalysis.CAdbAccountsWithRootHash.Adapter(adb.analysisAdapter))
	}

	dataTrie := adb.dataTries.Get(accountHandler.AddressBytes())
	if dataTrie != nil {
		stateAnalysis.Inc(stateAnalysis.CAdbDataTrieHolderHit.Adapter(adb.analysisAdapter))
		stateAnalysis.TouchDataTrie(adb.analysisAdapter, accountHandler.AddressBytes(), false)
		accountHandler.SetDataTrie(dataTrie)
		return nil
	}

	if len(accountHandler.GetRootHash()) == 0 {
		return nil
	}

	recreateStart := stateAnalysis.Now()
	if stateAnalysis.Enabled() {
		stateAnalysis.Inc(stateAnalysis.CAdbDataTrieRecreate.Adapter(adb.analysisAdapter))
		stateAnalysis.TouchDataTrie(adb.analysisAdapter, accountHandler.AddressBytes(), false)
		defer stateAnalysis.AddSince(stateAnalysis.CAdbDataTrieRecreateNs.Adapter(adb.analysisAdapter), recreateStart)
	}

	rootHashHolder := holders.NewDefaultRootHashesHolder(accountHandler.GetRootHash())
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
	dataTrie, err := mainTrie.Recreate(rootHashHolder)
	stateAnalysis.SetTrieKind(prevKind)
	if err != nil {
		return fmt.Errorf("trie was not found for hash, rootHash = %s, err = %w", hex.EncodeToString(accountHandler.GetRootHash()), err)
	}

	accountHandler.SetDataTrie(dataTrie)
	adb.dataTries.Put(accountHandler.AddressBytes(), dataTrie)
	return nil
}

// saveDataTrie is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) saveDataTrie(accountHandler baseAccountHandler) ([]*stateChange.DataTrieChange, error) {
	saveStart := stateAnalysis.Now()
	newValues, oldValues, err := accountHandler.SaveDirtyData(adb.mainTrie)
	if err != nil {
		return nil, err
	}
	if len(oldValues) == 0 {
		return nil, nil
	}
	if stateAnalysis.Enabled() {
		stateAnalysis.Inc(stateAnalysis.CAdbSaveDataTrieDirty.Adapter(adb.analysisAdapter))
		stateAnalysis.TouchDataTrie(adb.analysisAdapter, accountHandler.AddressBytes(), true)
		defer stateAnalysis.AddSince(stateAnalysis.CAdbSaveDataTrieNs.Adapter(adb.analysisAdapter), saveStart)
	}

	entry, err := NewJournalEntryDataTrieUpdates(oldValues, accountHandler)
	if err != nil {
		return nil, err
	}
	adb.journalize(entry)

	rootHashStart := stateAnalysis.Now()
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
	rootHash, err := accountHandler.DataTrie().RootHash()
	stateAnalysis.SetTrieKind(prevKind)
	stateAnalysis.AddSince(stateAnalysis.CAdbSaveDataTrieRootHashNs.Adapter(adb.analysisAdapter), rootHashStart)
	if err != nil {
		return nil, err
	}
	accountHandler.SetRootHash(rootHash)
	log.Trace("saveDataTrie: rootHash changed", "address", accountHandler.AddressBytes(), "rootHash", rootHash)

	if check.IfNil(adb.dataTries.Get(accountHandler.AddressBytes())) {
		trie, ok := accountHandler.DataTrie().(common.Trie)
		if !ok {
			log.Warn("wrong type conversion", "trie type", fmt.Sprintf("%T", accountHandler.DataTrie()))
			return nil, nil
		}

		adb.dataTries.Put(accountHandler.AddressBytes(), trie)
	}

	return newValues, nil
}

func (adb *AccountsDB) saveAccountToTrie(accountHandler vmcommon.AccountHandler, mainTrie common.Trie) ([]byte, error) {
	log.Trace("accountsDB.saveAccountToTrie", "address", accountHandler.AddressBytes())

	// pass the reference to marshaller, otherwise it will fail marshalling balance
	marshalStart := stateAnalysis.Now()
	buff, err := adb.marshaller.Marshal(accountHandler)
	if err != nil {
		return nil, err
	}
	if stateAnalysis.Enabled() {
		stateAnalysis.AddSince(stateAnalysis.CAdbSaveMarshalNs.Adapter(adb.analysisAdapter), marshalStart)
		stateAnalysis.Add(stateAnalysis.CAdbSaveBytes.Adapter(adb.analysisAdapter), int64(len(buff)))
	}

	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindForAdapter(adb.analysisAdapter))
	defer stateAnalysis.SetTrieKind(prevKind)
	err = mainTrie.Update(accountHandler.AddressBytes(), buff)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(address []byte) error {
	if len(address) == 0 {
		return fmt.Errorf("%w in RemoveAccount", ErrNilAddress)
	}

	// this is a critical section, do not remove the mutex
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	stateAnalysis.Inc(stateAnalysis.CAdbRemoveAccount.Adapter(adb.analysisAdapter))
	acnt, err := adb.getAccount(address, adb.mainTrie)
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

	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindForAdapter(adb.analysisAdapter))
	defer stateAnalysis.SetTrieKind(prevKind)
	return adb.mainTrie.Delete(address)
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

	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
	defer stateAnalysis.SetTrieKind(prevKind)
	dataTrie, err := adb.mainTrie.Recreate(rootHashHolder)
	if err != nil {
		return err
	}

	hashes, err := dataTrie.GetAllHashes()
	if err != nil {
		return err
	}
	stateAnalysis.Inc(stateAnalysis.CAdbRemoveDataTrie.Adapter(adb.analysisAdapter))
	stateAnalysis.Add(stateAnalysis.CAdbRemoveDataTrieHashes.Adapter(adb.analysisAdapter), int64(len(hashes)))

	adb.obsoleteDataTrieHashes[string(rootHash)] = hashes

	sc := &stateChange.StateAccess{
		Type:            stateChange.Write,
		MainTrieKey:     baseAcc.AddressBytes(),
		Operation:       stateChange.RemoveDataTrie,
		DataTrieChanges: nil,
	}
	adb.stateAccessesCollector.AddStateAccess(sc)
	entry, err := NewJournalEntryDataTrieRemove(rootHash, adb.obsoleteDataTrieHashes)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	return nil
}

func (adb *AccountsDB) removeCode(baseAcc baseAccountHandler) error {
	oldCodeHash := baseAcc.GetCodeHash()

	oldCodeEntry, err := adb.getCodeEntry(oldCodeHash)
	if err != nil {
		return err
	}

	entry, err := NewJournalEntryCode(oldCodeEntry, oldCodeHash, nil, nil, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	return adb.updateOldCodeEntry(oldCodeHash, oldCodeEntry)
}

// LoadAccount fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("%w in LoadAccount", ErrNilAddress)
	}

	log.Trace("accountsDB.LoadAccount", "address", address)
	stateAnalysis.Inc(stateAnalysis.CAdbLoadAccount.Adapter(adb.analysisAdapter))

	mainTrie := adb.getMainTrie()
	acnt, err := adb.getAccount(address, mainTrie)
	if err != nil {
		return nil, err
	}
	if check.IfNil(acnt) {
		return adb.accountFactory.CreateAccount(address)
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		err = adb.loadDataTrieConcurrentSafe(baseAcc, mainTrie)
		if err != nil {
			return nil, err
		}
	}

	return acnt, nil
}

func (adb *AccountsDB) getAccount(address []byte, mainTrie common.Trie) (vmcommon.AccountHandler, error) {
	acnt, _, err := adb.getAccountAndBytes(address, mainTrie)
	return acnt, err
}

func (adb *AccountsDB) getAccountAndBytes(address []byte, mainTrie common.Trie) (vmcommon.AccountHandler, []byte, error) {
	return adb.getAccountAndBytesWithTouch(address, mainTrie, true)
}

// getAccountAndBytesWithTouch reads the account; touch=false is used by SaveAccount whose re-read is not a new access of the address
func (adb *AccountsDB) getAccountAndBytesWithTouch(address []byte, mainTrie common.Trie, touch bool) (vmcommon.AccountHandler, []byte, error) {
	getStart := stateAnalysis.Now()
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindForAdapter(adb.analysisAdapter))
	val, _, err := mainTrie.Get(address)
	stateAnalysis.SetTrieKind(prevKind)
	if stateAnalysis.Enabled() {
		stateAnalysis.AddSince(stateAnalysis.CAdbGetAccountNs.Adapter(adb.analysisAdapter), getStart)
		if touch {
			stateAnalysis.TouchAccount(adb.analysisAdapter, address, len(val), false)
		}
		if val != nil {
			stateAnalysis.Inc(stateAnalysis.CAdbGetAccountFound.Adapter(adb.analysisAdapter))
			stateAnalysis.Add(stateAnalysis.CAdbGetAccountBytes.Adapter(adb.analysisAdapter), int64(len(val)))
		}
	}
	if err != nil {
		return nil, nil, err
	}
	if val == nil {
		return nil, nil, nil
	}

	sc := &stateChange.StateAccess{
		Type:            stateChange.Read,
		MainTrieKey:     address,
		MainTrieVal:     val,
		Operation:       stateChange.GetAccount,
		DataTrieChanges: nil,
	}
	adb.stateAccessesCollector.AddStateAccess(sc)

	acnt, err := adb.accountFactory.CreateAccount(address)
	if err != nil {
		return nil, nil, err
	}

	unmarshalStart := stateAnalysis.Now()
	err = adb.marshaller.Unmarshal(acnt, val)
	stateAnalysis.AddSince(stateAnalysis.CAdbUnmarshalNs.Adapter(adb.analysisAdapter), unmarshalStart)
	if err != nil {
		if adb.isReservedAccountAddress(address, val) {
			return nil, nil, ErrAccountAddressIsReserved
		}

		return nil, nil, err
	}

	return acnt, val, nil
}

func (adb *AccountsDB) isReservedAccountAddress(address []byte, value []byte) bool {
	var codeEntry CodeEntry
	err := adb.marshaller.Unmarshal(&codeEntry, value)
	if err != nil || len(codeEntry.Code) == 0 {
		return false
	}

	computedHash := adb.hasher.Compute(string(codeEntry.Code))
	return bytes.Equal(computedHash, address)
}

// GetExistingAccount returns an existing account if exists or nil if missing
func (adb *AccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("%w in GetExistingAccount", ErrNilAddress)
	}

	log.Trace("accountsDB.GetExistingAccount", "address", address)
	stateAnalysis.Inc(stateAnalysis.CAdbGetExisting.Adapter(adb.analysisAdapter))

	mainTrie := adb.getMainTrie()
	acnt, err := adb.getAccount(address, mainTrie)
	if err != nil {
		return nil, err
	}
	if check.IfNil(acnt) {
		stateAnalysis.Inc(stateAnalysis.CAdbGetExistingMissing.Adapter(adb.analysisAdapter))
		return nil, ErrAccNotFound
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		err = adb.loadDataTrieConcurrentSafe(baseAcc, mainTrie)
		if err != nil {
			return nil, err
		}
	}

	return acnt, nil
}

// GetAccountFromBytes returns an account from the given bytes
func (adb *AccountsDB) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("%w in GetAccountFromBytes", ErrNilAddress)
	}

	acnt, err := adb.accountFactory.CreateAccount(address)
	if err != nil {
		return nil, err
	}

	err = adb.marshaller.Unmarshal(acnt, accountBytes)
	if err != nil {
		return nil, err
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if !ok {
		return acnt, nil
	}

	err = adb.loadDataTrieConcurrentSafe(baseAcc, adb.getMainTrie())
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

	val, _, err := adb.mainTrie.Get(accountHandler.GetCodeHash())
	if err != nil {
		return err
	}

	var codeEntry CodeEntry
	err = adb.marshaller.Unmarshal(&codeEntry, val)
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
		stateAnalysis.Inc(stateAnalysis.CAdbRevertFull.Adapter(adb.analysisAdapter))
		stateAnalysis.Add(stateAnalysis.CAdbRevertEntries.Adapter(adb.analysisAdapter), int64(len(adb.entries)))
		return adb.recreateTrie(holders.NewDefaultRootHashesHolder(adb.lastRootHash))
	}

	stateAnalysis.Inc(stateAnalysis.CAdbRevertPartial.Adapter(adb.analysisAdapter))
	stateAnalysis.Add(stateAnalysis.CAdbRevertEntries.Adapter(adb.analysisAdapter), int64(len(adb.entries)-snapshot))
	mainKind := stateAnalysis.KindForAdapter(adb.analysisAdapter)
	prevKind := stateAnalysis.SetTrieKind(mainKind)
	defer stateAnalysis.SetTrieKind(prevKind)
	for i := len(adb.entries) - 1; i >= snapshot; i-- {
		switch adb.entries[i].(type) {
		case *journalEntryDataTrieUpdates, *journalEntryDataTrieRemove:
			stateAnalysis.SetTrieKind(stateAnalysis.KindData)
		default:
			stateAnalysis.SetTrieKind(mainKind)
		}
		account, err := adb.entries[i].Revert()
		if err != nil {
			adb.entries = adb.entries[:i+1]
			return err
		}

		if !check.IfNil(account) {
			_, err = adb.saveAccountToTrie(account, adb.mainTrie)
			if err != nil {
				adb.entries = adb.entries[:i+1]
				return err
			}
		}
	}

	adb.entries = adb.entries[:snapshot]

	return adb.stateAccessesCollector.RevertToIndex(snapshot)
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

// CommitInEpoch will commit the current trie state in the given epoch
func (adb *AccountsDB) CommitInEpoch(currentEpoch uint32, epochToCommit uint32) ([]byte, error) {
	adb.mutOp.Lock()
	defer func() {
		adb.mainTrie.GetStorageManager().SetEpochForPutOperation(currentEpoch)
		adb.mainTrie.GetStorageManager().GetStateStatsHandler().Reset()
		adb.mutOp.Unlock()
		adb.loadCodeMeasurements.resetAndPrint()
	}()

	adb.mainTrie.GetStorageManager().SetEpochForPutOperation(epochToCommit)

	return adb.commit()
}

func (adb *AccountsDB) printTrieStorageStatistics() {
	stats := adb.mainTrie.GetStorageManager().GetStateStatsHandler().ProcessingStats()
	if stats != nil {
		log.Debug("trie storage statistics",
			"stats", stats,
		)
	}

}

// Commit will persist all data inside the trie
func (adb *AccountsDB) Commit() ([]byte, error) {
	adb.mutOp.Lock()
	defer func() {
		adb.mutOp.Unlock()
		adb.loadCodeMeasurements.resetAndPrint()
	}()

	return adb.commit()
}

func (adb *AccountsDB) commit() ([]byte, error) {
	log.Trace("accountsDB.Commit started")
	commitStart := stateAnalysis.Now()
	adapter := adb.analysisAdapter
	if stateAnalysis.Enabled() {
		stateAnalysis.Inc(stateAnalysis.CAdbCommit.Adapter(adapter))
		stateAnalysis.Add(stateAnalysis.CAdbJournalLenAtCommit.Adapter(adapter), int64(len(adb.entries)))
		defer stateAnalysis.AddSince(stateAnalysis.CAdbCommitNs.Adapter(adapter), commitStart)
	}
	adb.entries = make([]JournalEntry, 0)

	oldHashes := make(common.ModifiedHashes)
	newHashes := make(common.ModifiedHashes)
	// Step 1. commit all data tries
	dataTries := adb.dataTries.GetAll()
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
	for i := 0; i < len(dataTries); i++ {
		err := adb.commitTrie(dataTries[i], oldHashes, newHashes)
		if err != nil {
			stateAnalysis.SetTrieKind(prevKind)
			return nil, err
		}
	}
	if stateAnalysis.Enabled() {
		stateAnalysis.Add(stateAnalysis.CAdbCommitDataTries.Adapter(adapter), int64(len(dataTries)))
		stateAnalysis.AddSince(stateAnalysis.CAdbCommitDataTriesNs.Adapter(adapter), commitStart)
		stateAnalysis.SetGauge(stateAnalysis.GHolderDataTries.At(adapter), int64(len(dataTries)))
		walkNs := adb.sampleResidentMemory(dataTries)
		// the walk is instrumentation cost, keep it out of the commit timers and of the outlier detection
		stateAnalysis.Add(stateAnalysis.CAdbCommitNs.Adapter(adapter), -walkNs)
		stateAnalysis.ExcludeFromBlockCommit(walkNs)
	}
	adb.dataTries.Reset()

	oldRoot := adb.mainTrie.GetOldRoot()

	// Step 2. commit main trie
	mainStart := stateAnalysis.Now()
	stateAnalysis.SetTrieKind(stateAnalysis.KindForAdapter(adapter))
	err := adb.commitTrie(adb.mainTrie, oldHashes, newHashes)
	stateAnalysis.SetTrieKind(prevKind)
	stateAnalysis.AddSince(stateAnalysis.CAdbCommitMainTrieNs.Adapter(adapter), mainStart)
	if err != nil {
		return nil, err
	}

	newRoot, err := adb.rootHash(adb.mainTrie)
	if err != nil {
		log.Trace("accountsDB.Commit ended", "error", err.Error())
		return nil, err
	}

	stateAccStart := stateAnalysis.Now()
	err = adb.stateAccessesCollector.CommitCollectedAccesses(newRoot)
	stateAnalysis.AddSince(stateAnalysis.CAdbCommitStateAccNs.Adapter(adapter), stateAccStart)
	if err != nil {
		return nil, err
	}

	evictStart := stateAnalysis.Now()
	err = adb.markForEviction(oldRoot, newRoot, oldHashes, newHashes)
	if stateAnalysis.Enabled() {
		stateAnalysis.AddSince(stateAnalysis.CAdbCommitEvictNs.Adapter(adapter), evictStart)
		stateAnalysis.Add(stateAnalysis.CAdbCommitOldHashes.Adapter(adapter), int64(len(oldHashes)))
		stateAnalysis.Add(stateAnalysis.CAdbCommitNewHashes.Adapter(adapter), int64(len(newHashes)))
	}
	if err != nil {
		return nil, err
	}

	adb.lastRootHash = newRoot
	adb.obsoleteDataTrieHashes = make(map[string][][]byte)

	log.Trace("accountsDB.Commit ended", "root hash", newRoot)

	adb.printTrieStorageStatistics()

	return newRoot, nil
}

func (adb *AccountsDB) markForEviction(
	oldRoot []byte,
	newRoot []byte,
	oldHashes common.ModifiedHashes,
	newHashes common.ModifiedHashes,
) error {
	if !adb.pruningEnabled {
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
	if adb.pruningEnabled {
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
	rootHash, err := adb.rootHash(adb.getMainTrie())
	log.Trace("accountsDB.RootHash",
		"root hash", rootHash,
		"err", err,
	)

	return rootHash, err
}

func (adb *AccountsDB) rootHash(tr common.Trie) ([]byte, error) {
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindForAdapter(adb.analysisAdapter))
	defer stateAnalysis.SetTrieKind(prevKind)

	return tr.RootHash()
}

// RecreateTrie is used to reload the trie based on the provided options
func (adb *AccountsDB) RecreateTrie(options common.RootHashHolder) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	err := adb.recreateTrie(options)
	if err != nil {
		return err
	}
	adb.lastRootHash = options.GetRootHash()

	return nil
}

// RecreateTrieIfNeeded is used to reload the trie based on the provided options if the root hash is different than the current one
func (adb *AccountsDB) RecreateTrieIfNeeded(options common.RootHashHolder) error {
	err := adb.recreateTrieIfNeeded(options)
	if err != nil {
		return err
	}

	adb.mutOp.Lock()
	adb.lastRootHash = options.GetRootHash()
	adb.mutOp.Unlock()

	return nil
}

func (adb *AccountsDB) recreateTrie(options common.RootHashHolder) error {
	log.Trace("accountsDB.RecreateTrie", "root hash holder", options.String())
	defer func() {
		log.Trace("accountsDB.RecreateTrie ended")
	}()

	adb.obsoleteDataTrieHashes = make(map[string][][]byte)
	adb.dataTries.Reset()
	adb.entries = make([]JournalEntry, 0)

	adb.stateAccessesCollector.Reset()
	stateAnalysis.Inc(stateAnalysis.CAdbRecreateTrie.Adapter(adb.analysisAdapter))
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindForAdapter(adb.analysisAdapter))
	defer stateAnalysis.SetTrieKind(prevKind)
	newTrie, err := adb.mainTrie.Recreate(options)
	if err != nil {
		return err
	}
	if check.IfNil(newTrie) {
		return ErrNilTrie
	}

	adb.mainTrie = newTrie
	return nil
}

func (adb *AccountsDB) recreateTrieIfNeeded(options common.RootHashHolder) error {
	currentRootHash, err := adb.rootHash(adb.getMainTrie())
	if err != nil {
		return err
	}

	if bytes.Equal(currentRootHash, options.GetRootHash()) {
		log.Trace("accountsDB.RecreateTrieIfNeeded - no need to recreate", "root hash", currentRootHash)
		return nil
	}

	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	return adb.recreateTrie(options)
}

// RecreateAllTries recreates all the tries from the accounts DB
func (adb *AccountsDB) RecreateAllTries(rootHash []byte) (map[string]common.Trie, error) {
	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, leavesChannelSize),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	mainTrie := adb.getMainTrie()
	err := mainTrie.GetAllLeavesOnChannel(
		leavesChannels,
		context.Background(),
		rootHash,
		keyBuilder.NewKeyBuilder(),
		parsers.NewMainTrieLeafParser(),
	)
	if err != nil {
		return nil, err
	}

	allTries, err := adb.recreateMainTrie(rootHash)
	if err != nil {
		return nil, err
	}

	for leaf := range leavesChannels.LeavesChan {
		userAccount, skipAccount, err := getUserAccountFromBytes(adb.accountFactory, adb.marshaller, leaf.Key(), leaf.Value())
		if err != nil {
			return nil, err
		}
		if skipAccount {
			continue
		}

		userAccountRootHash := userAccount.GetRootHash()
		if len(userAccountRootHash) > 0 {
			rootHashHolder := holders.NewDefaultRootHashesHolder(userAccountRootHash)
			dataTrie, errRecreate := mainTrie.Recreate(rootHashHolder)
			if errRecreate != nil {
				return nil, errRecreate
			}

			allTries[string(userAccountRootHash)] = dataTrie
		}
	}

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return allTries, nil
}

func getUserAccountFromBytes(accountFactory AccountFactory, marshaller marshal.Marshalizer, address []byte, accountBytes []byte) (UserAccountHandler, bool, error) {
	account, err := accountFactory.CreateAccount(address)
	if err != nil {
		return nil, true, err
	}

	err = marshaller.Unmarshal(account, accountBytes)
	if err != nil {
		log.Trace("this must be a leaf with code", "err", err)
		return nil, true, nil
	}

	userAccount, ok := account.(UserAccountHandler)
	if !ok {
		return nil, true, nil
	}

	return userAccount, false, nil
}

func (adb *AccountsDB) recreateMainTrie(rootHash []byte) (map[string]common.Trie, error) {
	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	recreatedTrie, err := adb.getMainTrie().Recreate(rootHashHolder)
	if err != nil {
		return nil, err
	}

	allTries := make(map[string]common.Trie)
	allTries[string(rootHash)] = recreatedTrie

	return allTries, nil
}

// GetTrie returns the trie that has the given rootHash
func (adb *AccountsDB) GetTrie(rootHash []byte) (common.Trie, error) {
	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	return adb.getMainTrie().Recreate(rootHashHolder)
}

// journalize adds a new object to entries list.
func (adb *AccountsDB) journalize(entry JournalEntry) {
	if check.IfNil(entry) {
		return
	}

	adb.entries = append(adb.entries, entry)
	stateAnalysis.Inc(stateAnalysis.CAdbJournalize.Adapter(adb.analysisAdapter))
	log.Trace("accountsDB.Journalize",
		"new length", len(adb.entries),
		"entry type", fmt.Sprintf("%T", entry),
	)

	err := adb.stateAccessesCollector.SetIndexToLatestStateAccesses(len(adb.entries))
	if err != nil {
		log.Trace("failed to set index to latest state accesses",
			"num journal entries", len(adb.entries),
			"error", err.Error(),
		)
	}

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
func (adb *AccountsDB) PruneTrie(rootHash []byte, identifier TriePruningIdentifier, handler PruningHandler) {
	log.Trace("accountsDB.PruneTrie", "root hash", rootHash)

	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	adb.storagePruningManager.PruneTrie(rootHash, identifier, adb.mainTrie.GetStorageManager(), handler)
}

// CancelPrune clears the trie's evictionWaitingList
func (adb *AccountsDB) CancelPrune(rootHash []byte, identifier TriePruningIdentifier) {
	log.Trace("accountsDB.CancelPrune", "root hash", rootHash)

	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	adb.storagePruningManager.CancelPrune(rootHash, identifier, adb.mainTrie.GetStorageManager())
}

// GetEvictionWaitingListSize returns the number of entries in the eviction waiting list cache
func (adb *AccountsDB) GetEvictionWaitingListSize() int {
	adb.mutOp.RLock()
	defer adb.mutOp.RUnlock()
	return adb.storagePruningManager.EvictionWaitingListCacheLen()
}

// ResetPruning will reset all collected data needed for pruning
func (adb *AccountsDB) ResetPruning() {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	adb.storagePruningManager.Reset()
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *AccountsDB) SnapshotState(rootHash []byte, epoch uint32) {
	start := stateAnalysis.Now()
	stateAnalysis.Inc(stateAnalysis.CAdbSnapshotStateCalls.Adapter(adb.analysisAdapter))
	adb.snapshotsManger.SnapshotState(rootHash, epoch, adb.getMainTrie().GetStorageManager())
	stateAnalysis.AddSince(stateAnalysis.CBlockSnapshotWait, start)
}

// sampleResidentMemory walks the resident part of the main trie and the given data tries every N blocks and
// returns the duration of the walk in nanoseconds
func (adb *AccountsDB) sampleResidentMemory(dataTries []common.Trie) int64 {
	if !stateAnalysis.ShouldSampleMemory() {
		return 0
	}

	start := time.Now()
	mainKind := stateAnalysis.KindForAdapter(adb.analysisAdapter)
	if provider, ok := adb.mainTrie.(stateAnalysis.ResidentStatsProvider); ok {
		stats := provider.ResidentStats()
		stateAnalysis.SetResidentStats(mainKind, stats)
	}

	if adb.analysisAdapter != stateAnalysis.AdapterUser {
		walkNs := int64(time.Since(start))
		stateAnalysis.SetGauge(stateAnalysis.GTrieResidentWalkNs, walkNs)
		return walkNs
	}

	total := stateAnalysis.ResidentStats{}
	for _, tr := range dataTries {
		provider, ok := tr.(stateAnalysis.ResidentStatsProvider)
		if !ok {
			continue
		}
		total.Add(provider.ResidentStats())
	}
	stateAnalysis.SetResidentStats(stateAnalysis.KindData, total)
	walkNs := int64(time.Since(start))
	stateAnalysis.SetGauge(stateAnalysis.GTrieResidentWalkNs, walkNs)
	return walkNs
}

func (adb *AccountsDB) getTrieStorageManagerAndLatestEpoch(mainTrie common.Trie) (common.StorageManager, uint32, error) {
	trieStorageManager := mainTrie.GetStorageManager()
	epoch, err := trieStorageManager.GetLatestStorageEpoch()
	if err != nil {
		return nil, 0, fmt.Errorf("%w while getting the latest storage epoch", err)
	}

	return trieStorageManager, epoch, nil
}

func emptyErrChanReturningHadContained(errChan chan error) bool {
	contained := false
	for {
		select {
		case <-errChan:
			contained = true
		default:
			return contained
		}
	}
}

// IsPruningEnabled returns true if state pruning is enabled
func (adb *AccountsDB) IsPruningEnabled() bool {
	return adb.pruningEnabled
}

// GetAllLeaves returns all the leaves from a given rootHash
func (adb *AccountsDB) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
	return adb.getMainTrie().GetAllLeavesOnChannel(leavesChannels, ctx, rootHash, keyBuilder.NewKeyBuilder(), trieLeafParser)
}

// Close will handle the closing of the underlying components
func (adb *AccountsDB) Close() error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	_ = adb.mainTrie.Close()
	return adb.storagePruningManager.Close()
}

// GetStatsForRootHash will get trie statistics for the given rootHash
func (adb *AccountsDB) GetStatsForRootHash(rootHash []byte) (common.TriesStatisticsCollector, error) {
	stats := statistics.NewTrieStatisticsCollector()
	mainTrie := adb.getMainTrie()

	tr, ok := mainTrie.(common.TrieStats)
	if !ok {
		return nil, fmt.Errorf("invalid trie, type is %T", mainTrie)
	}

	collectStats(tr, stats, rootHash, "", common.MainTrie)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, leavesChannelSize),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := mainTrie.GetAllLeavesOnChannel(
		iteratorChannels,
		context.Background(),
		rootHash,
		keyBuilder.NewKeyBuilder(),
		parsers.NewMainTrieLeafParser(),
	)
	if err != nil {
		return nil, err
	}

	for leaf := range iteratorChannels.LeavesChan {
		userAccount, skipAccount, err := getUserAccountFromBytes(adb.accountFactory, adb.marshaller, leaf.Key(), leaf.Value())
		if err != nil {
			return nil, err
		}
		if skipAccount {
			continue
		}

		if common.IsEmptyTrie(userAccount.GetRootHash()) {
			continue
		}

		accountAddress, err := adb.addressConverter.Encode(userAccount.AddressBytes())
		if err != nil {
			return nil, err
		}

		collectStats(tr, stats, userAccount.GetRootHash(), accountAddress, common.DataTrie)
	}

	err = iteratorChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func collectStats(
	tr common.TrieStats,
	stats common.TriesStatisticsCollector,
	rootHash []byte,
	address string,
	trieType common.TrieType,
) {
	trieStats, err := tr.GetTrieStats(address, rootHash)
	if err != nil {
		log.Error(err.Error())
		return
	}
	stats.Add(trieStats, trieType)

	log.Debug(strings.Join(trieStats.ToString(), " "))
}

// SetTxHashForLatestStateAccesses will return the state changes since the last call of this method
func (adb *AccountsDB) SetTxHashForLatestStateAccesses(txHash []byte) {
	adb.stateAccessesCollector.AddTxHashToCollectedStateAccesses(txHash)
}

// IsSnapshotInProgress returns true if there is a snapshot in progress
func (adb *AccountsDB) IsSnapshotInProgress() bool {
	return adb.snapshotsManger.IsSnapshotInProgress()
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *AccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
