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
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const (
	leavesChannelSize       = 100
	missingNodesChannelSize = 100
	lastSnapshot            = "lastSnapshot"
	userTrieSnapshotMsg     = "snapshotState user trie"
	peerTrieSnapshotMsg     = "snapshotState peer trie"
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

type accountMetrics struct {
	snapshotInProgressKey   string
	lastSnapshotDurationKey string
	snapshotMessage         string
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
	trieSyncer             AccountsDBSyncer

	isSnapshotInProgress atomic.Flag
	lastSnapshot         *snapshotInfo
	lastRootHash         []byte
	dataTries            common.TriesHolder
	entries              []JournalEntry

	mutOp                    sync.RWMutex
	processingMode           common.NodeProcessingMode
	shouldSerializeSnapshots bool
	loadCodeMeasurements     *loadingMeasurements
	processStatusHandler     common.ProcessStatusHandler
	appStatusHandler         core.AppStatusHandler
	addressConverter         core.PubkeyConverter

	stackDebug []byte
}

var log = logger.GetOrCreate("state")

// ArgsAccountsDB is the arguments DTO for the AccountsDB instance
type ArgsAccountsDB struct {
	Trie                     common.Trie
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	AccountFactory           AccountFactory
	StoragePruningManager    StoragePruningManager
	ProcessingMode           common.NodeProcessingMode
	ShouldSerializeSnapshots bool
	ProcessStatusHandler     common.ProcessStatusHandler
	AppStatusHandler         core.AppStatusHandler
	AddressConverter         core.PubkeyConverter
}

// NewAccountsDB creates a new account manager
func NewAccountsDB(args ArgsAccountsDB) (*AccountsDB, error) {
	err := checkArgsAccountsDB(args)
	if err != nil {
		return nil, err
	}

	args.AppStatusHandler.SetUInt64Value(common.MetricAccountsSnapshotInProgress, 0)

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
		processingMode:           args.ProcessingMode,
		shouldSerializeSnapshots: args.ShouldSerializeSnapshots,
		lastSnapshot:             &snapshotInfo{},
		processStatusHandler:     args.ProcessStatusHandler,
		appStatusHandler:         args.AppStatusHandler,
		isSnapshotInProgress:     atomic.Flag{},
		addressConverter:         args.AddressConverter,
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
	if check.IfNil(args.ProcessStatusHandler) {
		return ErrNilProcessStatusHandler
	}
	if check.IfNil(args.AppStatusHandler) {
		return ErrNilAppStatusHandler
	}
	if check.IfNil(args.AddressConverter) {
		return ErrNilAddressConverter
	}

	return nil
}

func startSnapshotAfterRestart(adb AccountsAdapter, tsm common.StorageManager, processingMode common.NodeProcessingMode) {
	epoch, err := tsm.GetLatestStorageEpoch()
	if err != nil {
		log.Error("could not get latest storage epoch")
	}
	putActiveDBMarker := epoch == 0 && err == nil
	isInImportDBMode := processingMode == common.ImportDb
	putActiveDBMarker = putActiveDBMarker || isInImportDBMode
	if putActiveDBMarker {
		log.Debug("marking activeDB", "epoch", epoch, "error", err, "processing mode", processingMode)
		err = tsm.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
		handleLoggingWhenError("error while putting active DB value into main storer", err)
		return
	}

	rootHash, err := tsm.GetFromCurrentEpoch([]byte(lastSnapshot))
	if err != nil {
		log.Debug("startSnapshotAfterRestart root hash", "error", err)
		return
	}
	log.Debug("snapshot hash after restart", "hash", rootHash)

	if tsm.ShouldTakeSnapshot() {
		log.Debug("startSnapshotAfterRestart")
		adb.SnapshotState(rootHash)
		return
	}
}

func handleLoggingWhenError(message string, err error, extraArguments ...interface{}) {
	if err == nil {
		return
	}
	if errors.IsClosingError(err) {
		args := []interface{}{"reason", err}
		log.Debug(message, append(args, extraArguments...)...)
		return
	}

	args := []interface{}{"error", err}
	log.Warn(message, append(args, extraArguments...)...)
}

// SetSyncer sets the given syncer as the syncer for the underlying trie
func (adb *AccountsDB) SetSyncer(syncer AccountsDBSyncer) error {
	if check.IfNil(syncer) {
		return ErrNilTrieSyncer
	}

	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	adb.trieSyncer = syncer
	return nil
}

// StartSnapshotIfNeeded starts the snapshot if the previous snapshot process was not fully completed
func (adb *AccountsDB) StartSnapshotIfNeeded() error {
	return startSnapshotIfNeeded(adb, adb.getTrieSyncer(), adb.getMainTrie().GetStorageManager(), adb.processingMode)
}

func startSnapshotIfNeeded(
	adb AccountsAdapter,
	trieSyncer AccountsDBSyncer,
	trieStorageManager common.StorageManager,
	processingMode common.NodeProcessingMode,
) error {
	if check.IfNil(trieSyncer) {
		return ErrNilTrieSyncer
	}

	val, err := trieStorageManager.GetFromCurrentEpoch([]byte(common.ActiveDBKey))
	if err != nil || !bytes.Equal(val, []byte(common.ActiveDBVal)) {
		startSnapshotAfterRestart(adb, trieStorageManager, processingMode)
	}

	return nil
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

	val, _, err := adb.getMainTrie().Get(codeHash)
	if err != nil {
		return nil
	}

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
	return adb.saveAccountToTrie(account, mainTrie)
}

func (adb *AccountsDB) getMainTrie() common.Trie {
	adb.mutOp.RLock()
	defer adb.mutOp.RUnlock()

	return adb.mainTrie
}

func (adb *AccountsDB) getTrieSyncer() AccountsDBSyncer {
	adb.mutOp.RLock()
	defer adb.mutOp.RUnlock()

	return adb.trieSyncer
}

// SaveAccount saves in the trie all changes made to the account.
func (adb *AccountsDB) SaveAccount(account vmcommon.AccountHandler) error {
	if check.IfNil(account) {
		return fmt.Errorf("%w in accountsDB SaveAccount", ErrNilAccountHandler)
	}

	// this is a critical section, do not remove the mutex
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	oldAccount, err := adb.getAccount(account.AddressBytes(), adb.mainTrie)
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

	return adb.saveAccountToTrie(account, adb.mainTrie)
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

	entry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, newCodeHash, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	newAcc.SetCodeHash(newCodeHash)
	return nil
}

func (adb *AccountsDB) updateOldCodeEntry(oldCodeHash []byte) (*CodeEntry, error) {
	oldCodeEntry, err := getCodeEntry(oldCodeHash, adb.mainTrie, adb.marshaller)
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
	err = saveCodeEntry(oldCodeHash, oldCodeEntry, adb.mainTrie, adb.marshaller)
	if err != nil {
		return nil, err
	}

	return unmodifiedOldCodeEntry, nil
}

func (adb *AccountsDB) updateNewCodeEntry(newCodeHash []byte, newCode []byte) error {
	if len(newCode) == 0 {
		return nil
	}

	newCodeEntry, err := getCodeEntry(newCodeHash, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}

	if newCodeEntry == nil {
		newCodeEntry = &CodeEntry{
			Code: newCode,
		}
	}
	newCodeEntry.NumReferences++

	err = saveCodeEntry(newCodeHash, newCodeEntry, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}

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

// loadDataTrieConcurrentSafe retrieves and saves the SC data inside accountHandler object.
// Errors if something went wrong
func (adb *AccountsDB) loadDataTrieConcurrentSafe(accountHandler baseAccountHandler, mainTrie common.Trie) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	dataTrie := adb.dataTries.Get(accountHandler.AddressBytes())
	if dataTrie != nil {
		accountHandler.SetDataTrie(dataTrie)
		return nil
	}

	if len(accountHandler.GetRootHash()) == 0 {
		return nil
	}

	dataTrie, err := mainTrie.Recreate(accountHandler.GetRootHash())
	if err != nil {
		return fmt.Errorf("trie was not found for hash, rootHash = %s, err = %w", hex.EncodeToString(accountHandler.GetRootHash()), err)
	}

	accountHandler.SetDataTrie(dataTrie)
	adb.dataTries.Put(accountHandler.AddressBytes(), dataTrie)
	return nil
}

// SaveDataTrie is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) saveDataTrie(accountHandler baseAccountHandler) error {
	oldValues, err := accountHandler.SaveDirtyData(adb.mainTrie)
	if err != nil {
		return err
	}
	if len(oldValues) == 0 {
		return nil
	}

	entry, err := NewJournalEntryDataTrieUpdates(oldValues, accountHandler)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	rootHash, err := accountHandler.DataTrie().RootHash()
	if err != nil {
		return err
	}
	accountHandler.SetRootHash(rootHash)

	if check.IfNil(adb.dataTries.Get(accountHandler.AddressBytes())) {
		trie, ok := accountHandler.DataTrie().(common.Trie)
		if !ok {
			log.Warn("wrong type conversion", "trie type", fmt.Sprintf("%T", accountHandler.DataTrie()))
			return nil
		}

		adb.dataTries.Put(accountHandler.AddressBytes(), trie)
	}

	return nil
}

func (adb *AccountsDB) saveAccountToTrie(accountHandler vmcommon.AccountHandler, mainTrie common.Trie) error {
	log.Trace("accountsDB.saveAccountToTrie",
		"address", hex.EncodeToString(accountHandler.AddressBytes()),
	)

	// pass the reference to marshaller, otherwise it will fail marshalling balance
	buff, err := adb.marshaller.Marshal(accountHandler)
	if err != nil {
		return err
	}

	return mainTrie.Update(accountHandler.AddressBytes(), buff)
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

	codeChangeEntry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, nil, adb.mainTrie, adb.marshaller)
	if err != nil {
		return err
	}
	adb.journalize(codeChangeEntry)

	return nil
}

// LoadAccount fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("%w in LoadAccount", ErrNilAddress)
	}

	log.Trace("accountsDB.LoadAccount",
		"address", hex.EncodeToString(address),
	)

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
	val, _, err := mainTrie.Get(address)
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

	err = adb.marshaller.Unmarshal(acnt, val)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// GetExistingAccount returns an existing account if exists or nil if missing
func (adb *AccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("%w in GetExistingAccount", ErrNilAddress)
	}

	log.Trace("accountsDB.GetExistingAccount",
		"address", hex.EncodeToString(address),
	)

	mainTrie := adb.getMainTrie()
	acnt, err := adb.getAccount(address, mainTrie)
	if err != nil {
		return nil, err
	}
	if check.IfNil(acnt) {
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
		return adb.recreateTrie(holders.NewRootHashHolder(adb.lastRootHash, core.OptionalUint32{}))
	}

	for i := len(adb.entries) - 1; i >= snapshot; i-- {
		account, err := adb.entries[i].Revert()
		if err != nil {
			return err
		}

		if !check.IfNil(account) {
			err = adb.saveAccountToTrie(account, adb.mainTrie)
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

// CommitInEpoch will commit the current trie state in the given epoch
func (adb *AccountsDB) CommitInEpoch(currentEpoch uint32, epochToCommit uint32) ([]byte, error) {
	adb.mutOp.Lock()
	defer func() {
		adb.mainTrie.GetStorageManager().SetEpochForPutOperation(currentEpoch)
		adb.mutOp.Unlock()
		adb.loadCodeMeasurements.resetAndPrint()
	}()

	adb.mainTrie.GetStorageManager().SetEpochForPutOperation(epochToCommit)

	return adb.commit()
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
	rootHash, err := adb.getMainTrie().RootHash()
	log.Trace("accountsDB.RootHash",
		"root hash", rootHash,
		"err", err,
	)

	return rootHash, err
}

// RecreateTrie is used to reload the trie based on an existing rootHash
func (adb *AccountsDB) RecreateTrie(rootHash []byte) error {
	return adb.RecreateTrieFromEpoch(holders.NewRootHashHolder(rootHash, core.OptionalUint32{}))
}

// RecreateTrieFromEpoch is used to reload the trie based on the provided options
func (adb *AccountsDB) RecreateTrieFromEpoch(options common.RootHashHolder) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	err := adb.recreateTrie(options)
	if err != nil {
		return err
	}
	adb.lastRootHash = options.GetRootHash()

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
	newTrie, err := adb.mainTrie.RecreateFromEpoch(options)
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
	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, leavesChannelSize),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	mainTrie := adb.getMainTrie()
	err := mainTrie.GetAllLeavesOnChannel(leavesChannels, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	if err != nil {
		return nil, err
	}

	allTries, err := adb.recreateMainTrie(rootHash)
	if err != nil {
		return nil, err
	}

	for leaf := range leavesChannels.LeavesChan {
		account := &userAccount{}
		err = adb.marshaller.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) > 0 {
			dataTrie, errRecreate := mainTrie.Recreate(account.RootHash)
			if errRecreate != nil {
				return nil, errRecreate
			}

			allTries[string(account.RootHash)] = dataTrie
		}
	}

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return allTries, nil
}

func (adb *AccountsDB) recreateMainTrie(rootHash []byte) (map[string]common.Trie, error) {
	recreatedTrie, err := adb.getMainTrie().Recreate(rootHash)
	if err != nil {
		return nil, err
	}

	allTries := make(map[string]common.Trie)
	allTries[string(rootHash)] = recreatedTrie

	return allTries, nil
}

// GetTrie returns the trie that has the given rootHash
func (adb *AccountsDB) GetTrie(rootHash []byte) (common.Trie, error) {
	return adb.getMainTrie().Recreate(rootHash)
}

// Journalize adds a new object to entries list.
func (adb *AccountsDB) journalize(entry JournalEntry) {
	if check.IfNil(entry) {
		return
	}

	adb.entries = append(adb.entries, entry)
	log.Trace("accountsDB.Journalize",
		"new length", len(adb.entries),
		"entry type", fmt.Sprintf("%T", entry),
	)

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

	adb.storagePruningManager.PruneTrie(rootHash, identifier, adb.getMainTrie().GetStorageManager(), handler)
}

// CancelPrune clears the trie's evictionWaitingList
func (adb *AccountsDB) CancelPrune(rootHash []byte, identifier TriePruningIdentifier) {
	log.Trace("accountsDB.CancelPrune", "root hash", rootHash)

	adb.storagePruningManager.CancelPrune(rootHash, identifier, adb.getMainTrie().GetStorageManager())
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *AccountsDB) SnapshotState(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	trieStorageManager, epoch, shouldTakeSnapshot := adb.prepareSnapshot(rootHash)
	if !shouldTakeSnapshot {
		return
	}

	log.Info("starting snapshot user trie", "rootHash", rootHash, "epoch", epoch)
	missingNodesChannel := make(chan []byte, missingNodesChannelSize)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, leavesChannelSize),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	stats := newSnapshotStatistics(1, 1)

	accountMetricsInstance := &accountMetrics{
		snapshotInProgressKey:   common.MetricAccountsSnapshotInProgress,
		lastSnapshotDurationKey: common.MetricLastAccountsSnapshotDurationSec,
		snapshotMessage:         userTrieSnapshotMsg,
	}
	adb.updateMetricsOnSnapshotStart(accountMetricsInstance)

	go func() {
		stats.NewSnapshotStarted()

		trieStorageManager.TakeSnapshot("", rootHash, rootHash, iteratorChannels, missingNodesChannel, stats, epoch)
		adb.snapshotUserAccountDataTrie(true, rootHash, iteratorChannels, missingNodesChannel, stats, epoch)

		stats.SnapshotFinished()
	}()

	go adb.syncMissingNodes(missingNodesChannel, iteratorChannels.ErrChan, stats, adb.trieSyncer)

	go adb.processSnapshotCompletion(stats, trieStorageManager, missingNodesChannel, iteratorChannels.ErrChan, rootHash, accountMetricsInstance, epoch)

	adb.waitForCompletionIfAppropriate(stats)
}

func (adb *AccountsDB) prepareSnapshot(rootHash []byte) (common.StorageManager, uint32, bool) {
	trieStorageManager, epoch, err := adb.getTrieStorageManagerAndLatestEpoch(adb.mainTrie)
	if err != nil {
		log.Error("prepareSnapshot error", "err", err.Error())
		return nil, 0, false
	}

	defer func() {
		err = trieStorageManager.PutInEpoch([]byte(lastSnapshot), rootHash, epoch)
		handleLoggingWhenError("could not set lastSnapshot", err, "rootHash", rootHash)
	}()

	if !adb.shouldTakeSnapshot(trieStorageManager, rootHash, epoch) {
		log.Debug("skipping snapshot",
			"last snapshot rootHash", adb.lastSnapshot.rootHash,
			"rootHash", rootHash,
			"last snapshot epoch", adb.lastSnapshot.epoch,
			"epoch", epoch,
			"isSnapshotInProgress", adb.isSnapshotInProgress.IsSet(),
		)
		return nil, 0, false
	}

	adb.isSnapshotInProgress.SetValue(true)
	adb.lastSnapshot.rootHash = rootHash
	adb.lastSnapshot.epoch = epoch
	trieStorageManager.EnterPruningBufferingMode()

	return trieStorageManager, epoch, true
}

func (adb *AccountsDB) getTrieStorageManagerAndLatestEpoch(mainTrie common.Trie) (common.StorageManager, uint32, error) {
	trieStorageManager := mainTrie.GetStorageManager()
	epoch, err := trieStorageManager.GetLatestStorageEpoch()
	if err != nil {
		return nil, 0, fmt.Errorf("%w while getting the latest storage epoch", err)
	}

	return trieStorageManager, epoch, nil
}

func (adb *AccountsDB) shouldTakeSnapshot(trieStorageManager common.StorageManager, rootHash []byte, epoch uint32) bool {
	snapshotAlreadyTaken := bytes.Equal(adb.lastSnapshot.rootHash, rootHash) && adb.lastSnapshot.epoch == epoch
	if snapshotAlreadyTaken {
		return false
	}

	if adb.isSnapshotInProgress.IsSet() {
		return false
	}

	return trieStorageManager.ShouldTakeSnapshot()
}

func (adb *AccountsDB) finishSnapshotOperation(
	rootHash []byte,
	stats *snapshotStatistics,
	missingNodesCh chan []byte,
	message string,
	trieStorageManager common.StorageManager,
) {
	stats.WaitForSnapshotsToFinish()
	close(missingNodesCh)
	stats.WaitForSyncToFinish()

	trieStorageManager.ExitPruningBufferingMode()

	stats.PrintStats(message, rootHash)
}

func (adb *AccountsDB) updateMetricsOnSnapshotStart(metrics *accountMetrics) {
	adb.appStatusHandler.SetUInt64Value(metrics.snapshotInProgressKey, 1)
	adb.appStatusHandler.SetInt64Value(metrics.lastSnapshotDurationKey, 0)
}

func (adb *AccountsDB) updateMetricsOnSnapshotCompletion(metrics *accountMetrics, stats *snapshotStatistics) {
	adb.appStatusHandler.SetUInt64Value(metrics.snapshotInProgressKey, 0)
	adb.appStatusHandler.SetInt64Value(metrics.lastSnapshotDurationKey, stats.GetSnapshotDuration())
	if metrics.snapshotMessage == userTrieSnapshotMsg {
		adb.appStatusHandler.SetUInt64Value(common.MetricAccountsSnapshotNumNodes, stats.GetSnapshotNumNodes())
	}
}

func (adb *AccountsDB) processSnapshotCompletion(
	stats *snapshotStatistics,
	trieStorageManager common.StorageManager,
	missingNodesCh chan []byte,
	errChan common.BufferedErrChan,
	rootHash []byte,
	metrics *accountMetrics,
	epoch uint32,
) {
	adb.finishSnapshotOperation(rootHash, stats, missingNodesCh, metrics.snapshotMessage, trieStorageManager)

	defer func() {
		adb.isSnapshotInProgress.Reset()
		adb.updateMetricsOnSnapshotCompletion(metrics, stats)
		errChan.Close()
	}()

	errorDuringSnapshot := errChan.ReadFromChanNonBlocking()
	shouldNotMarkActive := trieStorageManager.IsClosed() || errorDuringSnapshot != nil
	if shouldNotMarkActive {
		log.Debug("will not set activeDB in epoch as the snapshot might be incomplete",
			"epoch", epoch, "trie storage manager closed", trieStorageManager.IsClosed(),
			"errors during snapshot found", errorDuringSnapshot)
		return
	}

	err := trieStorageManager.RemoveFromAllEpochs([]byte(lastSnapshot))
	handleLoggingWhenError("could not remove lastSnapshot", err, "rootHash", rootHash)

	log.Debug("set activeDB in epoch", "epoch", epoch)
	errPut := trieStorageManager.PutInEpochWithoutCache([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal), epoch)
	handleLoggingWhenError("error while putting active DB value into main storer", errPut)
}

func (adb *AccountsDB) syncMissingNodes(missingNodesChan chan []byte, errChan common.BufferedErrChan, stats *snapshotStatistics, syncer AccountsDBSyncer) {
	defer stats.SyncFinished()

	if check.IfNil(syncer) {
		log.Error("can not sync missing nodes", "error", ErrNilTrieSyncer.Error())
		for missingNode := range missingNodesChan {
			log.Warn("could not sync node", "hash", missingNode)
		}
		errChan.WriteInChanNonBlocking(ErrNilTrieSyncer)
		return
	}

	for missingNode := range missingNodesChan {
		err := syncer.SyncAccounts(missingNode)
		if err != nil {
			log.Error("could not sync missing node",
				"missing node hash", missingNode,
				"error", err,
			)
			errChan.WriteInChanNonBlocking(err)
		}
	}
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

func (adb *AccountsDB) snapshotUserAccountDataTrie(
	isSnapshot bool,
	mainTrieRootHash []byte,
	iteratorChannels *common.TrieIteratorChannels,
	missingNodesChannel chan []byte,
	stats common.SnapshotStatisticsHandler,
	epoch uint32,
) {
	for leaf := range iteratorChannels.LeavesChan {
		account := &userAccount{}
		err := adb.marshaller.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) == 0 {
			continue
		}

		stats.NewSnapshotStarted()

		iteratorChannelsForDataTries := &common.TrieIteratorChannels{
			LeavesChan: nil,
			ErrChan:    iteratorChannels.ErrChan,
		}
		if isSnapshot {
			address := adb.addressConverter.Encode(account.Address)
			adb.mainTrie.GetStorageManager().TakeSnapshot(address, account.RootHash, mainTrieRootHash, iteratorChannelsForDataTries, missingNodesChannel, stats, epoch)
			continue
		}

		adb.mainTrie.GetStorageManager().SetCheckpoint(account.RootHash, mainTrieRootHash, iteratorChannelsForDataTries, missingNodesChannel, stats)
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

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, leavesChannelSize),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	missingNodesChannel := make(chan []byte, missingNodesChannelSize)
	stats := newSnapshotStatistics(1, 1)
	go func() {
		stats.NewSnapshotStarted()
		trieStorageManager.SetCheckpoint(rootHash, rootHash, iteratorChannels, missingNodesChannel, stats)
		adb.snapshotUserAccountDataTrie(false, rootHash, iteratorChannels, missingNodesChannel, stats, 0)

		stats.SnapshotFinished()
	}()

	go adb.syncMissingNodes(missingNodesChannel, iteratorChannels.ErrChan, stats, adb.trieSyncer)

	// TODO decide if we need to take some actions whenever we hit an error that occurred in the checkpoint process
	//  that will be present in the errChan var
	go adb.finishSnapshotOperation(rootHash, stats, missingNodesChannel, "setStateCheckpoint user trie", trieStorageManager)

	adb.waitForCompletionIfAppropriate(stats)
}

func (adb *AccountsDB) waitForCompletionIfAppropriate(stats common.SnapshotStatisticsHandler) {
	shouldSerializeSnapshots := adb.shouldSerializeSnapshots || adb.processingMode == common.ImportDb
	if !shouldSerializeSnapshots {
		return
	}

	log.Debug("manually setting idle on the process status handler in order to be able to start & complete the snapshotting/checkpointing process")
	adb.processStatusHandler.SetIdle()

	stats.WaitForSnapshotsToFinish()
}

// IsPruningEnabled returns true if state pruning is enabled
func (adb *AccountsDB) IsPruningEnabled() bool {
	return adb.getMainTrie().GetStorageManager().IsPruningEnabled()
}

// GetAllLeaves returns all the leaves from a given rootHash
func (adb *AccountsDB) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte) error {
	return adb.getMainTrie().GetAllLeavesOnChannel(leavesChannels, ctx, rootHash, keyBuilder.NewKeyBuilder())
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

	collectStats(tr, stats, rootHash, "")

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, leavesChannelSize),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := mainTrie.GetAllLeavesOnChannel(iteratorChannels, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	if err != nil {
		return nil, err
	}

	for leaf := range iteratorChannels.LeavesChan {
		account := &userAccount{}
		err = adb.marshaller.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if common.IsEmptyTrie(account.RootHash) {
			continue
		}

		address := adb.addressConverter.Encode(account.Address)
		collectStats(tr, stats, account.RootHash, address)
	}

	err = iteratorChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func collectStats(tr common.TrieStats, stats common.TriesStatisticsCollector, rootHash []byte, address string) {
	trieStats, err := tr.GetTrieStats(address, rootHash)
	if err != nil {
		log.Error(err.Error())
		return
	}
	stats.Add(trieStats)

	log.Debug(strings.Join(trieStats.ToString(), " "))
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *AccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
