package trie

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// trieStorageManager manages all the storage operations of the trie (commit, snapshot, checkpoint, pruning)
type trieStorageManager struct {
	mainStorer             common.DBWriteCacher
	pruningBlockingOps     uint32
	snapshotReq            chan *snapshotsQueueEntry
	checkpointReq          chan *snapshotsQueueEntry
	checkpointsStorer      common.DBWriteCacher
	checkpointHashesHolder CheckpointHashesHolder
	storageOperationMutex  sync.RWMutex
	cancelFunc             context.CancelFunc
	closer                 core.SafeCloser
	closed                 bool

	// TODO remove these fields after the new implementation is in production
	db                     common.DBWriteCacher
	snapshots              []common.SnapshotDbHandler
	snapshotId             int
	snapshotDbCfg          config.DBConfig
	maxSnapshots           uint32
	keepSnapshots          bool
	flagDisableOldStorage  atomic.Flag
	disableOldStorageEpoch uint32
	oldStorageClosed       bool
}

type snapshotsQueueEntry struct {
	rootHash         []byte
	mainTrieRootHash []byte
	leavesChan       chan core.KeyValueHolder
	stats            common.SnapshotStatisticsHandler
	epoch            uint32
}

// NewTrieStorageManagerArgs holds the arguments needed for creating a new trieStorageManager
type NewTrieStorageManagerArgs struct {
	EpochNotifier              EpochNotifier
	DisableOldTrieStorageEpoch uint32
	DB                         common.DBWriteCacher
	MainStorer                 common.DBWriteCacher
	CheckpointsStorer          common.DBWriteCacher
	Marshalizer                marshal.Marshalizer
	Hasher                     hashing.Hasher
	SnapshotDbConfig           config.DBConfig
	GeneralConfig              config.TrieStorageManagerConfig
	CheckpointHashesHolder     CheckpointHashesHolder
}

// NewTrieStorageManager creates a new instance of trieStorageManager
func NewTrieStorageManager(args NewTrieStorageManagerArgs) (*trieStorageManager, error) {
	if check.IfNil(args.DB) {
		return nil, ErrNilDatabase
	}
	if check.IfNil(args.MainStorer) {
		return nil, fmt.Errorf("%w for main storer", ErrNilStorer)
	}
	if check.IfNil(args.CheckpointsStorer) {
		return nil, fmt.Errorf("%w for checkpoints storer", ErrNilStorer)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.CheckpointHashesHolder) {
		return nil, ErrNilCheckpointHashesHolder
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, ErrNilEpochNotifier
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	tsm := &trieStorageManager{
		db:                     args.DB,
		mainStorer:             args.MainStorer,
		checkpointsStorer:      args.CheckpointsStorer,
		snapshotDbCfg:          args.SnapshotDbConfig,
		snapshotReq:            make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		checkpointReq:          make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		pruningBlockingOps:     0,
		maxSnapshots:           args.GeneralConfig.MaxSnapshots,
		keepSnapshots:          args.GeneralConfig.KeepSnapshots,
		cancelFunc:             cancelFunc,
		checkpointHashesHolder: args.CheckpointHashesHolder,
		closer:                 closing.NewSafeChanCloser(),
		disableOldStorageEpoch: args.DisableOldTrieStorageEpoch,
		oldStorageClosed:       false,
	}
	goRoutinesThrottler, err := throttler.NewNumGoRoutinesThrottler(int32(args.GeneralConfig.SnapshotsGoroutineNum))
	if err != nil {
		return nil, err
	}

	log.Debug("epoch for disabling old trie storage", "epoch", tsm.disableOldStorageEpoch)
	args.EpochNotifier.RegisterNotifyHandler(tsm)

	err = tsm.mainStorer.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	if err != nil {
		log.Warn("newTrieStorageManager error while putting active DB value into main storer", "error", err)
	}

	if tsm.flagDisableOldStorage.IsSet() {
		err := tsm.db.Close()
		if err != nil {
			return nil, err
		}
		tsm.oldStorageClosed = true

		go tsm.doCheckpointsAndSnapshots(ctx, args.Marshalizer, args.Hasher, goRoutinesThrottler)
		return tsm, nil
	}

	snapshots, snapshotId, err := getSnapshotsAndSnapshotId(args.SnapshotDbConfig)
	if err != nil {
		log.Debug("get snapshot", "error", err.Error())
	}

	tsm.snapshots = snapshots
	tsm.snapshotId = snapshotId

	go tsm.doCheckpointsAndSnapshots(ctx, args.Marshalizer, args.Hasher, goRoutinesThrottler)
	return tsm, nil
}

func (tsm *trieStorageManager) doCheckpointsAndSnapshots(ctx context.Context, msh marshal.Marshalizer, hsh hashing.Hasher, goRoutinesThrottler core.Throttler) {
	tsm.doProcessLoop(ctx, msh, hsh, goRoutinesThrottler)
	tsm.cleanupChans()
}

func (tsm *trieStorageManager) doProcessLoop(ctx context.Context, msh marshal.Marshalizer, hsh hashing.Hasher, goRoutinesThrottler core.Throttler) {
	defer log.Debug("trieStorageManager.storageProcessLoop go routine is closing...")

	for {
		select {
		case snapshotRequest := <-tsm.snapshotReq:
			err := tsm.checkGoRoutinesThrottler(ctx, goRoutinesThrottler, snapshotRequest)
			if err != nil {
				return
			}

			goRoutinesThrottler.StartProcessing()
			go tsm.takeSnapshot(snapshotRequest, msh, hsh, ctx, goRoutinesThrottler)
		case snapshotRequest := <-tsm.checkpointReq:
			err := tsm.checkGoRoutinesThrottler(ctx, goRoutinesThrottler, snapshotRequest)
			if err != nil {
				return
			}

			goRoutinesThrottler.StartProcessing()
			go tsm.takeCheckpoint(snapshotRequest, msh, hsh, ctx, goRoutinesThrottler)
		case <-ctx.Done():
			return
		}
	}
}

func (tsm *trieStorageManager) checkGoRoutinesThrottler(
	ctx context.Context,
	goRoutinesThrottler core.Throttler,
	snapshotRequest *snapshotsQueueEntry,
) error {
	for {
		if goRoutinesThrottler.CanProcess() {
			break
		}

		select {
		case <-time.After(time.Millisecond * 100):
			continue
		case <-ctx.Done():
			tsm.finishOperation(snapshotRequest, "did not start snapshot, goroutione is closing")
			return ErrTimeIsOut
		}
	}

	return nil
}

func (tsm *trieStorageManager) cleanupChans() {
	<-tsm.closer.ChanClose()
	//at this point we can not add new entries in the snapshot/checkpoint chans
	for {
		select {
		case entry := <-tsm.snapshotReq:
			tsm.finishOperation(entry, "trie snapshot finished on cleanup")
		case entry := <-tsm.checkpointReq:
			tsm.finishOperation(entry, "trie checkpoint finished on cleanup")
		default:
			log.Debug("finished trieStorageManager.cleanupChans")
			return
		}
	}
}

func getOrderedSnapshots(snapshotsMap map[int]common.SnapshotDbHandler) []common.SnapshotDbHandler {
	snapshots := make([]common.SnapshotDbHandler, 0)
	keys := make([]int, 0)

	for key := range snapshotsMap {
		keys = append(keys, key)
	}

	sort.Ints(keys)
	for _, key := range keys {
		snapshots = append(snapshots, snapshotsMap[key])
	}

	return snapshots
}

func getSnapshotsAndSnapshotId(snapshotDbCfg config.DBConfig) ([]common.SnapshotDbHandler, int, error) {
	snapshotsMap := make(map[int]common.SnapshotDbHandler)
	snapshotId := 0

	if !directoryExists(snapshotDbCfg.FilePath) {
		return getOrderedSnapshots(snapshotsMap), snapshotId, nil
	}

	files, err := ioutil.ReadDir(snapshotDbCfg.FilePath)
	if err != nil {
		log.Debug("there is no snapshot in path", "path", snapshotDbCfg.FilePath)
		return getOrderedSnapshots(snapshotsMap), snapshotId, err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		var snapshotName int
		snapshotName, err = strconv.Atoi(f.Name())
		if err != nil {
			return getOrderedSnapshots(snapshotsMap), snapshotId, err
		}

		var db storage.Persister
		arg := storageUnit.ArgDB{
			DBType:            storageUnit.DBType(snapshotDbCfg.Type),
			Path:              path.Join(snapshotDbCfg.FilePath, f.Name()),
			BatchDelaySeconds: snapshotDbCfg.BatchDelaySeconds,
			MaxBatchSize:      snapshotDbCfg.MaxBatchSize,
			MaxOpenFiles:      snapshotDbCfg.MaxOpenFiles,
		}
		db, err = storageUnit.NewDB(arg)
		if err != nil {
			return getOrderedSnapshots(snapshotsMap), snapshotId, err
		}

		if snapshotName > snapshotId {
			snapshotId = snapshotName
		}

		newSnapshot := &snapshotDb{
			DBWriteCacher: db,
		}

		log.Debug("restored snapshot", "snapshot ID", snapshotName)
		snapshotsMap[snapshotName] = newSnapshot
	}

	if len(snapshotsMap) != 0 {
		snapshotId++
	}

	return getOrderedSnapshots(snapshotsMap), snapshotId, nil
}

//Get checks all the storers for the given key, and returns it if it is found
func (tsm *trieStorageManager) Get(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	if tsm.closed {
		log.Debug("trieStorageManager get context closing", "key", key)
		return nil, ErrContextClosing
	}

	val, err := tsm.mainStorer.Get(key)
	if isClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		return val, nil
	}

	return tsm.getFromOtherStorers(key)
}

func (tsm *trieStorageManager) getFromOtherStorers(key []byte) ([]byte, error) {
	val, err := tsm.checkpointsStorer.Get(key)
	if isClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		return val, nil
	}

	if tsm.flagDisableOldStorage.IsSet() {
		return nil, ErrKeyNotFound
	}

	val, err = tsm.db.Get(key)
	if isClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		return val, nil
	}

	for i := len(tsm.snapshots) - 1; i >= 0; i-- {
		val, _ = tsm.snapshots[i].Get(key)
		if len(val) != 0 {
			return val, nil
		}
	}

	return nil, ErrKeyNotFound
}

func isClosingError(err error) bool {
	if err == ErrContextClosing || err == storage.ErrSerialDBIsClosed {
		return true
	}

	return false
}

// Put adds the given value to the main storer
func (tsm *trieStorageManager) Put(key []byte, val []byte) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()
	log.Trace("put hash in tsm", "hash", key)

	if tsm.closed {
		log.Debug("trieStorageManager put context closing", "key", key, "value", val)
		return ErrContextClosing
	}

	return tsm.mainStorer.Put(key, val)
}

// PutInEpoch adds the given value to the main storer in the specified epoch
func (tsm *trieStorageManager) PutInEpoch(key []byte, val []byte, epoch uint32) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()
	log.Trace("put hash in tsm in epoch", "hash", key, "epoch", epoch)

	if tsm.closed {
		log.Debug("trieStorageManager put context closing", "key", key, "value", val, "epoch", epoch)
		return ErrContextClosing
	}

	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return fmt.Errorf("invalid storer type for PutInEpoch")
	}

	return storer.PutInEpochWithoutCache(key, val, epoch)
}

// EnterPruningBufferingMode increases the counter that tracks how many operations
// that block the pruning process are in progress
func (tsm *trieStorageManager) EnterPruningBufferingMode() {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.pruningBlockingOps++

	log.Trace("enter pruning buffering state", "operations in progress that block pruning", tsm.pruningBlockingOps)
}

// ExitPruningBufferingMode decreases the counter that tracks how many operations
// that block the pruning process are in progress
func (tsm *trieStorageManager) ExitPruningBufferingMode() {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	if tsm.pruningBlockingOps < 1 {
		log.Error("ExitPruningBufferingMode called too many times")
		return
	}

	tsm.pruningBlockingOps--

	log.Trace("exit pruning buffering state", "operations in progress that block pruning", tsm.pruningBlockingOps)
}

// GetLatestStorageEpoch returns the epoch for the latest opened persister
func (tsm *trieStorageManager) GetLatestStorageEpoch() (uint32, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		log.Debug("GetLatestStorageEpoch", "error", fmt.Sprintf("%T", tsm.mainStorer))
		return 0, fmt.Errorf("invalid storer type for GetLatestStorageEpoch")
	}

	return storer.GetLatestStorageEpoch()
}

// TakeSnapshot creates a new snapshot, or if there is another snapshot or checkpoint in progress,
// it adds this snapshot in the queue.
func (tsm *trieStorageManager) TakeSnapshot(
	rootHash []byte,
	mainTrieRootHash []byte,
	leavesChan chan core.KeyValueHolder,
	stats common.SnapshotStatisticsHandler,
	epoch uint32,
) {
	if tsm.isClosed() {
		tsm.safelyCloseChan(leavesChan)
		stats.SnapshotFinished()
		return
	}

	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not snapshot an empty trie")
		tsm.safelyCloseChan(leavesChan)
		stats.SnapshotFinished()
		return
	}

	tsm.EnterPruningBufferingMode()
	tsm.checkpointHashesHolder.RemoveCommitted(rootHash)

	snapshotEntry := &snapshotsQueueEntry{
		rootHash:         rootHash,
		mainTrieRootHash: mainTrieRootHash,
		leavesChan:       leavesChan,
		stats:            stats,
		epoch:            epoch,
	}
	select {
	case tsm.snapshotReq <- snapshotEntry:
	case <-tsm.closer.ChanClose():
		tsm.ExitPruningBufferingMode()
		tsm.safelyCloseChan(leavesChan)
		stats.SnapshotFinished()
	}
}

// SetCheckpoint creates a new checkpoint, or if there is another snapshot or checkpoint in progress,
// it adds this checkpoint in the queue. The checkpoint operation creates a new snapshot file
// only if there was no snapshot done prior to this
func (tsm *trieStorageManager) SetCheckpoint(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, stats common.SnapshotStatisticsHandler) {
	if tsm.isClosed() {
		tsm.safelyCloseChan(leavesChan)
		stats.SnapshotFinished()
		return
	}

	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not set checkpoint for empty trie")
		tsm.safelyCloseChan(leavesChan)
		stats.SnapshotFinished()
		return
	}

	tsm.EnterPruningBufferingMode()

	checkpointEntry := &snapshotsQueueEntry{
		rootHash:         rootHash,
		mainTrieRootHash: mainTrieRootHash,
		leavesChan:       leavesChan,
		stats:            stats,
	}
	select {
	case tsm.checkpointReq <- checkpointEntry:
	case <-tsm.closer.ChanClose():
		tsm.ExitPruningBufferingMode()
		tsm.safelyCloseChan(leavesChan)
		stats.SnapshotFinished()
	}
}

func (tsm *trieStorageManager) safelyCloseChan(ch chan core.KeyValueHolder) {
	if ch != nil {
		close(ch)
	}
}

func (tsm *trieStorageManager) finishOperation(snapshotEntry *snapshotsQueueEntry, message string) {
	tsm.ExitPruningBufferingMode()
	log.Trace(message, "rootHash", snapshotEntry.rootHash)
	tsm.safelyCloseChan(snapshotEntry.leavesChan)
	snapshotEntry.stats.SnapshotFinished()
}

func (tsm *trieStorageManager) takeSnapshot(snapshotEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher, ctx context.Context, goRoutinesThrottler core.Throttler) {
	defer func() {
		tsm.finishOperation(snapshotEntry, "trie snapshot finished")
		goRoutinesThrottler.EndProcessing()
	}()

	log.Trace("trie snapshot started", "rootHash", snapshotEntry.rootHash)

	newRoot, err := newSnapshotNode(tsm, msh, hsh, snapshotEntry.rootHash)
	if err != nil {
		treatSnapshotError(err,
			"trie storage manager: newSnapshotNode takeSnapshot",
			snapshotEntry.rootHash,
			snapshotEntry.mainTrieRootHash,
		)
		return
	}

	stsm, err := newSnapshotTrieStorageManager(tsm, snapshotEntry.epoch)
	if err != nil {
		log.Error("takeSnapshot: trie storage manager: newSnapshotTrieStorageManager",
			"rootHash", snapshotEntry.rootHash,
			"main trie rootHash", snapshotEntry.mainTrieRootHash,
			"err", err.Error())
		return
	}

	err = newRoot.commitSnapshot(stsm, snapshotEntry.leavesChan, ctx, snapshotEntry.stats)
	if err != nil {
		treatSnapshotError(err,
			"trie storage manager: takeSnapshot commit",
			snapshotEntry.rootHash,
			snapshotEntry.mainTrieRootHash,
		)
		return
	}
}

func (tsm *trieStorageManager) takeCheckpoint(checkpointEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher, ctx context.Context, goRoutinesThrottler core.Throttler) {
	defer func() {
		tsm.finishOperation(checkpointEntry, "trie checkpoint finished")
		goRoutinesThrottler.EndProcessing()
	}()

	log.Trace("trie checkpoint started", "rootHash", checkpointEntry.rootHash)

	newRoot, err := newSnapshotNode(tsm, msh, hsh, checkpointEntry.rootHash)
	if err != nil {
		treatSnapshotError(err,
			"trie storage manager: newSnapshotNode takeCheckpoint",
			checkpointEntry.rootHash,
			checkpointEntry.mainTrieRootHash,
		)
		return
	}

	err = newRoot.commitCheckpoint(tsm, tsm.checkpointsStorer, tsm.checkpointHashesHolder, checkpointEntry.leavesChan, ctx, checkpointEntry.stats)
	if err != nil {
		treatSnapshotError(err,
			"trie storage manager: takeCheckpoint commit",
			checkpointEntry.rootHash,
			checkpointEntry.mainTrieRootHash,
		)
		return
	}
}

func treatSnapshotError(err error, message string, rootHash []byte, mainTrieRootHash []byte) {
	if isClosingError(err) {
		log.Debug("context closing", "message", message, "rootHash", rootHash, "mainTrieRootHash", mainTrieRootHash)
		return
	}

	log.Error(message, "rootHash", rootHash, "mainTrieRootHash", mainTrieRootHash, "err", err.Error())
}

func newSnapshotNode(
	db common.DBWriteCacher,
	msh marshal.Marshalizer,
	hsh hashing.Hasher,
	rootHash []byte,
) (snapshotNode, error) {
	newRoot, err := getNodeFromDBAndDecode(rootHash, db, msh, hsh)
	if err != nil {
		return nil, err
	}

	return newRoot, nil
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsPruningEnabled returns true if the trie pruning is enabled
func (tsm *trieStorageManager) IsPruningEnabled() bool {
	return true
}

// IsPruningBlocked returns true if there is any pruningBlockingOperation in progress
func (tsm *trieStorageManager) IsPruningBlocked() bool {
	tsm.storageOperationMutex.RLock()
	defer tsm.storageOperationMutex.RUnlock()

	return tsm.pruningBlockingOps != 0
}

// GetSnapshotDbBatchDelay returns the batch write delay in seconds
func (tsm *trieStorageManager) GetSnapshotDbBatchDelay() int {
	return tsm.snapshotDbCfg.BatchDelaySeconds
}

// AddDirtyCheckpointHashes adds the given hashes to the checkpoint hashes holder
func (tsm *trieStorageManager) AddDirtyCheckpointHashes(rootHash []byte, hashes common.ModifiedHashes) bool {
	return tsm.checkpointHashesHolder.Put(rootHash, hashes)
}

// Remove removes the given hash form the storage and from the checkpoint hashes holder
func (tsm *trieStorageManager) Remove(hash []byte) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.checkpointHashesHolder.Remove(hash)
	return tsm.mainStorer.Remove(hash)
}

func (tsm *trieStorageManager) isClosed() bool {
	tsm.storageOperationMutex.RLock()
	defer tsm.storageOperationMutex.RUnlock()

	return tsm.closed
}

// Close - closes all underlying components
func (tsm *trieStorageManager) Close() error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.cancelFunc()
	tsm.closed = true

	//calling close on the SafeCloser instance should be the last instruction called
	//(just to close some go routines started as edge cases that would otherwise hang)
	defer tsm.closer.Close()

	var err error
	if !tsm.flagDisableOldStorage.IsSet() {
		err = tsm.closeOldTrieStorage()
	}

	errMainStorerClose := tsm.mainStorer.Close()
	if errMainStorerClose != nil {
		log.Error("trieStorageManager.Close mainStorerClose", "error", errMainStorerClose)
		err = errMainStorerClose
	}

	errCheckpointsStorerClose := tsm.checkpointsStorer.Close()
	if errCheckpointsStorerClose != nil {
		log.Error("trieStorageManager.Close checkpointsStorerClose", "error", errCheckpointsStorerClose)
		err = errCheckpointsStorerClose
	}

	if err != nil {
		return fmt.Errorf("trieStorageManager close failed: %w", err)
	}

	return nil
}

func (tsm *trieStorageManager) closeOldTrieStorage() error {
	err := tsm.db.Close()

	for _, sdb := range tsm.snapshots {
		errSnapshotClose := sdb.Close()
		if errSnapshotClose != nil {
			log.Error("trieStorageManager.Close snapshotClose", "error", errSnapshotClose)
			err = errSnapshotClose
		}
	}

	tsm.oldStorageClosed = true
	return err
}

// SetEpochForPutOperation will set the storer for the given epoch as the current storer
func (tsm *trieStorageManager) SetEpochForPutOperation(epoch uint32) {
	storer, ok := tsm.mainStorer.(epochStorer)
	if !ok {
		log.Error("invalid storer for ChangeEpochForPutOperations", "epoch", epoch)
		return
	}

	storer.SetEpochForPutOperation(epoch)
}

// ShouldTakeSnapshot returns true if the conditions for a new snapshot are met
func (tsm *trieStorageManager) ShouldTakeSnapshot() bool {
	stsm, err := newSnapshotTrieStorageManager(tsm, 0)
	if err != nil {
		log.Error("shouldTakeSnapshot error", "err", err.Error())
		return false
	}

	val, err := stsm.GetFromLastEpoch([]byte(common.ActiveDBKey))
	if err != nil {
		log.Debug("shouldTakeSnapshot get error", "err", err.Error())
		return false
	}

	if bytes.Equal(val, []byte(common.ActiveDBVal)) {
		return true
	}

	log.Debug("shouldTakeSnapshot invalid value for activeDBKey", "value", val)
	return false
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (tsm *trieStorageManager) EpochConfirmed(epoch uint32, _ uint64) {
	tsm.flagDisableOldStorage.SetValue(epoch >= tsm.disableOldStorageEpoch)
	log.Debug("old trie storage", "disabled", tsm.flagDisableOldStorage.IsSet())

	if tsm.flagDisableOldStorage.IsSet() && !tsm.oldStorageClosed {
		err := tsm.closeOldTrieStorage()
		if err != nil {
			log.Error("could not close old trie storage", "error", err.Error())
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsm *trieStorageManager) IsInterfaceNil() bool {
	return tsm == nil
}
