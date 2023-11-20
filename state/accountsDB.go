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
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/stateMetrics"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const (
	leavesChannelSize             = 100
	missingNodesChannelSize       = 100
	lastSnapshot                  = "lastSnapshot"
	waitTimeForSnapshotEpochCheck = time.Millisecond * 100
	snapshotWaitTimeout           = time.Minute
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

	lastRootHash []byte
	dataTries    common.TriesHolder
	entries      []JournalEntry

	mutOp                sync.RWMutex
	loadCodeMeasurements *loadingMeasurements
	addressConverter     core.PubkeyConverter
	enableEpochsHandler  common.EnableEpochsHandler

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
	EnableEpochsHandler      common.EnableEpochsHandler
}

// NewAccountsDB creates a new account manager
func NewAccountsDB(args ArgsAccountsDB) (*AccountsDB, error) {
	err := checkArgsAccountsDB(args)
	if err != nil {
		return nil, err
	}

	argStateMetrics := stateMetrics.ArgsStateMetrics{
		SnapshotInProgressKey:   common.MetricAccountsSnapshotInProgress,
		LastSnapshotDurationKey: common.MetricLastAccountsSnapshotDurationSec,
		SnapshotMessage:         stateMetrics.UserTrieSnapshotMsg,
	}
	sm, err := stateMetrics.NewStateMetrics(argStateMetrics, args.AppStatusHandler)
	if err != nil {
		return nil, err
	}

	argsSnapshotsManager := ArgsNewSnapshotsManager{
		ShouldSerializeSnapshots: args.ShouldSerializeSnapshots,
		ProcessingMode:           args.ProcessingMode,
		Marshaller:               args.Marshaller,
		AddressConverter:         args.AddressConverter,
		ProcessStatusHandler:     args.ProcessStatusHandler,
		StateMetrics:             sm,
		ChannelsProvider:         iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
		AccountFactory:           args.AccountFactory,
	}
	snapshotManager, err := NewSnapshotsManager(argsSnapshotsManager)
	if err != nil {
		return nil, err
	}

	return createAccountsDb(args, snapshotManager), nil
}

func createAccountsDb(args ArgsAccountsDB, snapshotManager SnapshotsManager) *AccountsDB {
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
		addressConverter:    args.AddressConverter,
		snapshotsManger:     snapshotManager,
		enableEpochsHandler: args.EnableEpochsHandler,
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
	if check.IfNil(args.EnableEpochsHandler) {
		return ErrNilEnableEpochsHandler
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

	var code []byte

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		adb.loadCodeMeasurements.addMeasurement(len(code), duration)
	}()

	return adb.getCode(codeHash)
}

func (adb *AccountsDB) getCode(codeHash []byte) []byte {
	code, err := adb.getMainTrie().GetStorageManager().Get(codeHash)
	if err == nil {
		return code
	}

	codeEntry, err := getCodeEntry(codeHash, adb.mainTrie, adb.marshaller)
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

func (adb *AccountsDB) migrateCode(
	newAcc baseAccountHandler,
	oldAccVersion, newAccVersion uint8,
	newCodeHash, oldCodeHash, newCode []byte,
) error {
	unmodifiedOldCodeEntry, err := adb.updateOldCodeEntry(oldCodeHash, oldAccVersion)
	if err != nil {
		return err
	}

	err = adb.mainTrie.GetStorageManager().Put(newCodeHash, newCode)
	if err != nil {
		return err
	}

	entry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, newCodeHash, adb.mainTrie, adb.marshaller, adb.enableEpochsHandler, oldAccVersion, newAccVersion)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	newAcc.SetCodeHash(newCodeHash)

	return nil
}

func (adb *AccountsDB) saveCode(newAcc, oldAcc baseAccountHandler) error {
	// TODO when state splitting is implemented, check how the code should be copied in different shards

	oldAccVersion := uint8(0)
	if !check.IfNil(oldAcc) {
		oldAccVersion = oldAcc.GetVersion()
	}

	newAccVersion := uint8(0)
	if !check.IfNil(newAcc) {
		newAccVersion = newAcc.GetVersion()
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

	if newAcc.GetVersion() == uint8(core.WithoutCodeLeaf) && oldAccVersion == uint8(core.NotSpecified) {
		return adb.migrateCode(newAcc, oldAccVersion, newAccVersion, newCodeHash, oldCodeHash, newCode)
	}

	if !newAcc.HasNewCode() {
		return nil
	}

	if bytes.Equal(oldCodeHash, newCodeHash) {
		newAcc.SetCodeHash(newCodeHash)
		return nil
	}

	unmodifiedOldCodeEntry, err := adb.updateOldCodeEntry(oldCodeHash, oldAccVersion)
	if err != nil {
		return err
	}

	err = adb.updateNewCodeEntry(newCodeHash, newCode, newAccVersion)
	if err != nil {
		return err
	}

	entry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, newCodeHash, adb.mainTrie, adb.marshaller, adb.enableEpochsHandler, oldAccVersion, newAccVersion)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	newAcc.SetCodeHash(newCodeHash)
	return nil
}

func (adb *AccountsDB) updateOldCodeEntry(oldCodeHash []byte, oldAccVersion uint8) (*CodeEntry, error) {
	if adb.enableEpochsHandler.IsFlagEnabled(common.RemoveCodeLeafFlag) &&
		oldAccVersion == uint8(core.WithoutCodeLeaf) {
		// do not update old code entry since num references is not used anymore
		return nil, nil
	}

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
		err = adb.mainTrie.Delete(oldCodeHash)
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

func (adb *AccountsDB) updateNewCodeEntry(newCodeHash []byte, newCode []byte, newAccVersion uint8) error {
	if len(newCode) == 0 {
		return nil
	}

	if adb.enableEpochsHandler.IsFlagEnabled(common.RemoveCodeLeafFlag) &&
		newAccVersion == uint8(core.WithoutCodeLeaf) {
		return adb.mainTrie.GetStorageManager().Put(newCodeHash, newCode)
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

// this has to be used for trie with code entry
func getCodeEntry(codeHash []byte, updater Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	leafHolder, err := updater.Get(codeHash)
	if err != nil {
		return nil, err
	}

	if len(leafHolder.Value()) == 0 {
		return nil, nil
	}

	var codeEntry CodeEntry
	err = marshalizer.Unmarshal(&codeEntry, leafHolder.Value())
	if err != nil {
		return nil, err
	}

	return &codeEntry, nil
}

// this has to be used for trie with code entry
func saveCodeEntry(codeHash []byte, entry *CodeEntry, updater Updater, marshalizer marshal.Marshalizer) error {
	codeEntry, err := marshalizer.Marshal(entry)
	if err != nil {
		return err
	}

	err = updater.Update(codeHash, codeEntry)
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

	//TODO in order to avoid recomputing the root hash after every transaction for the same data trie,
	// benchmark if it is better to cache the account and compute the rootHash only when the state is committed.
	// For this to work, LoadAccount should check that cache first, and only after load from the trie.
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

	unmodifiedOldCodeEntry, err := adb.updateOldCodeEntry(oldCodeHash, baseAcc.GetVersion())
	if err != nil {
		return err
	}

	codeChangeEntry, err := NewJournalEntryCode(unmodifiedOldCodeEntry, oldCodeHash, nil, adb.mainTrie, adb.marshaller, adb.enableEpochsHandler, baseAcc.GetVersion(), 0)
	if err != nil {
		return err
	}
	adb.journalize(codeChangeEntry)

	return nil
}

// MigrateCodeLeaf will remove code leaf node from main trie and put it to trie storage
func (adb *AccountsDB) MigrateCodeLeaf(account vmcommon.AccountHandler) error {
	if !adb.enableEpochsHandler.IsFlagEnabled(common.RemoveCodeLeafFlag) {
		log.Warn("remove code leaf operation not enabled")
		return nil
	}

	baseAcc, ok := account.(baseAccountHandler)
	if !ok {
		return errors.ErrWrongTypeAssertion
	}

	if baseAcc.GetVersion() == uint8(core.WithoutCodeLeaf) {
		return nil
	}

	codeHash := baseAcc.GetCodeHash()
	if len(codeHash) != adb.hasher.Size() {
		return nil
	}

	leafHolder, err := adb.mainTrie.Get(codeHash)
	if err != nil {
		return err
	}

	codeData := leafHolder.Value()
	if codeData == nil {
		return nil
	}

	// check if node is code leaf
	var oldCodeEntry CodeEntry
	err = adb.marshaller.Unmarshal(&oldCodeEntry, codeData)
	if err != nil {
		return err
	}

	baseAcc.SetVersion(uint8(core.WithoutCodeLeaf))

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
	leafHolder, err := mainTrie.Get(address)
	if err != nil {
		return nil, err
	}
	if leafHolder.Value() == nil {
		return nil, nil
	}

	acnt, err := adb.accountFactory.CreateAccount(address)
	if err != nil {
		return nil, err
	}

	err = adb.marshaller.Unmarshal(acnt, leafHolder.Value())
	if err != nil {
		return nil, err
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		baseAcc.SetVersion(uint8(leafHolder.Version()))
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

	leafHolder, err := adb.mainTrie.Get(accountHandler.GetCodeHash())
	if err != nil {
		return err
	}

	var codeEntry CodeEntry
	err = adb.marshaller.Unmarshal(&codeEntry, leafHolder.Value())
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
		adb.snapshotsManger.SetStateCheckpoint(newRoot, adb.mainTrie.GetStorageManager())
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
			dataTrie, errRecreate := mainTrie.Recreate(userAccountRootHash)
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
func (adb *AccountsDB) SnapshotState(rootHash []byte, epoch uint32) {
	adb.snapshotsManger.SnapshotState(rootHash, epoch, adb.getMainTrie().GetStorageManager())
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

// SetStateCheckpoint sets a checkpoint for the state trie
func (adb *AccountsDB) SetStateCheckpoint(rootHash []byte) {
	adb.snapshotsManger.SetStateCheckpoint(rootHash, adb.getMainTrie().GetStorageManager())
}

// IsPruningEnabled returns true if state pruning is enabled
func (adb *AccountsDB) IsPruningEnabled() bool {
	return adb.getMainTrie().GetStorageManager().IsPruningEnabled()
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

// IsSnapshotInProgress returns true if there is a snapshot in progress
func (adb *AccountsDB) IsSnapshotInProgress() bool {
	return adb.snapshotsManger.IsSnapshotInProgress()
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *AccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
