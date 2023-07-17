package trie

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

// trieStorageManager manages all the storage operations of the trie (commit, snapshot, checkpoint, pruning)
type trieStorageManager struct {
	mainStorer             common.StorerWithStats
	checkpointsStorer      common.BaseStorer
	pruningBlockingOps     uint32
	snapshotReq            chan *snapshotsQueueEntry
	checkpointReq          chan *snapshotsQueueEntry
	checkpointHashesHolder CheckpointHashesHolder
	storageOperationMutex  sync.RWMutex
	cancelFunc             context.CancelFunc
	closer                 core.SafeCloser
	closed                 bool
	idleProvider           IdleNodeProvider
	identifier             string
	statsCollector         common.StateStatisticsHandler
}

type snapshotsQueueEntry struct {
	address          string
	rootHash         []byte
	mainTrieRootHash []byte
	iteratorChannels *common.TrieIteratorChannels
	missingNodesChan chan []byte
	stats            common.SnapshotStatisticsHandler
	epoch            uint32
}

// NewTrieStorageManagerArgs holds the arguments needed for creating a new trieStorageManager
type NewTrieStorageManagerArgs struct {
	MainStorer             common.BaseStorer
	CheckpointsStorer      common.BaseStorer
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	GeneralConfig          config.TrieStorageManagerConfig
	CheckpointHashesHolder CheckpointHashesHolder
	IdleProvider           IdleNodeProvider
	Identifier             string
	StatsCollector         common.StateStatisticsHandler
}

// NewTrieStorageManager creates a new instance of trieStorageManager
func NewTrieStorageManager(args NewTrieStorageManagerArgs) (*trieStorageManager, error) {
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
	if check.IfNil(args.IdleProvider) {
		return nil, ErrNilIdleNodeProvider
	}
	if len(args.Identifier) == 0 {
		return nil, ErrInvalidIdentifier
	}
	if check.IfNil(args.StatsCollector) {
		return nil, storage.ErrNilStatsCollector
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	storerWithStats, ok := args.MainStorer.(common.StorerWithStats)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	tsm := &trieStorageManager{
		mainStorer:             storerWithStats,
		checkpointsStorer:      args.CheckpointsStorer,
		snapshotReq:            make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		checkpointReq:          make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		pruningBlockingOps:     0,
		cancelFunc:             cancelFunc,
		checkpointHashesHolder: args.CheckpointHashesHolder,
		closer:                 closing.NewSafeChanCloser(),
		idleProvider:           args.IdleProvider,
		identifier:             args.Identifier,
		statsCollector:         args.StatsCollector,
	}
	goRoutinesThrottler, err := throttler.NewNumGoRoutinesThrottler(int32(args.GeneralConfig.SnapshotsGoroutineNum))
	if err != nil {
		return nil, err
	}

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
	// at this point we can not add new entries in the snapshot/checkpoint chans
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

// Get checks all the storers for the given key, and returns it if it is found
func (tsm *trieStorageManager) Get(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	if tsm.closed {
		log.Trace("trieStorageManager get context closing", "key", key)
		return nil, core.ErrContextClosing
	}

	val, foundInCache, err := tsm.mainStorer.GetWithStats(key)
	if core.IsClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		if foundInCache {
			tsm.statsCollector.IncrCacheOp()
		} else {
			tsm.statsCollector.IncrPersisterOp()
		}

		return val, nil
	}

	return tsm.getFromOtherStorers(key)
}

func (tsm *trieStorageManager) GetStatsCollector() common.StateStatisticsHandler {
	return tsm.statsCollector
}

// GetFromCurrentEpoch checks only the current storer for the given key, and returns it if it is found
func (tsm *trieStorageManager) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()

	if tsm.closed {
		log.Trace("trieStorageManager get context closing", "key", key)
		tsm.storageOperationMutex.Unlock()
		return nil, core.ErrContextClosing
	}

	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		storerType := fmt.Sprintf("%T", tsm.mainStorer)
		tsm.storageOperationMutex.Unlock()
		return nil, fmt.Errorf("invalid storer, type is %s", storerType)
	}

	tsm.storageOperationMutex.Unlock()

	return storer.GetFromCurrentEpoch(key)
}

func (tsm *trieStorageManager) getFromOtherStorers(key []byte) ([]byte, error) {
	val, err := tsm.checkpointsStorer.Get(key)
	if core.IsClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		return val, nil
	}

	return nil, ErrKeyNotFound
}

// Put adds the given value to the main storer
func (tsm *trieStorageManager) Put(key []byte, val []byte) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()
	log.Trace("put hash in tsm", "hash", key)

	if tsm.closed {
		log.Trace("trieStorageManager put context closing", "key", key, "value", val)
		return core.ErrContextClosing
	}

	return tsm.mainStorer.Put(key, val)
}

// PutInEpoch adds the given value to the main storer in the specified epoch
func (tsm *trieStorageManager) PutInEpoch(key []byte, val []byte, epoch uint32) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()
	log.Trace("put hash in tsm in epoch", "hash", key, "epoch", epoch)

	if tsm.closed {
		log.Trace("trieStorageManager putInEpoch context closing", "key", key, "value", val, "epoch", epoch)
		return core.ErrContextClosing
	}

	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return fmt.Errorf("invalid storer type for PutInEpoch")
	}

	return storer.PutInEpoch(key, val, epoch)
}

// PutInEpochWithoutCache adds the given value to the main storer in the specified epoch without saving it to cache
func (tsm *trieStorageManager) PutInEpochWithoutCache(key []byte, val []byte, epoch uint32) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()
	log.Trace("put hash in tsm in epoch without cache", "hash", key, "epoch", epoch)

	if tsm.closed {
		log.Trace("trieStorageManager putInEpochWithoutCache context closing", "key", key, "value", val, "epoch", epoch)
		return core.ErrContextClosing
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
	address string,
	rootHash []byte,
	mainTrieRootHash []byte,
	iteratorChannels *common.TrieIteratorChannels,
	missingNodesChan chan []byte,
	stats common.SnapshotStatisticsHandler,
	epoch uint32,
) {
	if iteratorChannels.ErrChan == nil {
		log.Error("programming error in trieStorageManager.TakeSnapshot, cannot take snapshot because errChan is nil")
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
		return
	}
	if tsm.IsClosed() {
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
		return
	}

	if bytes.Equal(rootHash, common.EmptyTrieHash) {
		log.Trace("should not snapshot an empty trie")
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
		return
	}

	tsm.EnterPruningBufferingMode()
	tsm.checkpointHashesHolder.RemoveCommitted(rootHash)

	snapshotEntry := &snapshotsQueueEntry{
		address:          address,
		rootHash:         rootHash,
		mainTrieRootHash: mainTrieRootHash,
		iteratorChannels: iteratorChannels,
		missingNodesChan: missingNodesChan,
		stats:            stats,
		epoch:            epoch,
	}
	select {
	case tsm.snapshotReq <- snapshotEntry:
	case <-tsm.closer.ChanClose():
		tsm.ExitPruningBufferingMode()
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
	}
}

// SetCheckpoint creates a new checkpoint, or if there is another snapshot or checkpoint in progress,
// it adds this checkpoint in the queue. The checkpoint operation creates a new snapshot file
// only if there was no snapshot done prior to this
func (tsm *trieStorageManager) SetCheckpoint(
	rootHash []byte,
	mainTrieRootHash []byte,
	iteratorChannels *common.TrieIteratorChannels,
	missingNodesChan chan []byte,
	stats common.SnapshotStatisticsHandler,
) {
	if iteratorChannels.ErrChan == nil {
		log.Error("programming error in trieStorageManager.SetCheckpoint, cannot set checkpoint because errChan is nil")
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
		return
	}
	if tsm.IsClosed() {
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
		return
	}

	if bytes.Equal(rootHash, common.EmptyTrieHash) {
		log.Trace("should not set checkpoint for empty trie")
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
		return
	}

	tsm.EnterPruningBufferingMode()

	checkpointEntry := &snapshotsQueueEntry{
		rootHash:         rootHash,
		mainTrieRootHash: mainTrieRootHash,
		iteratorChannels: iteratorChannels,
		missingNodesChan: missingNodesChan,
		stats:            stats,
	}
	select {
	case tsm.checkpointReq <- checkpointEntry:
	case <-tsm.closer.ChanClose():
		tsm.ExitPruningBufferingMode()
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
		stats.SnapshotFinished()
	}
}

func (tsm *trieStorageManager) finishOperation(snapshotEntry *snapshotsQueueEntry, message string) {
	tsm.ExitPruningBufferingMode()
	log.Trace(message, "rootHash", snapshotEntry.rootHash)
	common.CloseKeyValueHolderChan(snapshotEntry.iteratorChannels.LeavesChan)
	snapshotEntry.stats.SnapshotFinished()
}

func (tsm *trieStorageManager) takeSnapshot(snapshotEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher, ctx context.Context, goRoutinesThrottler core.Throttler) {
	defer func() {
		tsm.finishOperation(snapshotEntry, "trie snapshot finished")
		goRoutinesThrottler.EndProcessing()
	}()

	log.Trace("trie snapshot started", "rootHash", snapshotEntry.rootHash)

	stsm, err := newSnapshotTrieStorageManager(tsm, snapshotEntry.epoch)
	if err != nil {
		snapshotEntry.iteratorChannels.ErrChan.WriteInChanNonBlocking(err)
		log.Error("takeSnapshot: trie storage manager: newSnapshotTrieStorageManager",
			"rootHash", snapshotEntry.rootHash,
			"main trie rootHash", snapshotEntry.mainTrieRootHash,
			"err", err.Error())
		return
	}

	newRoot, err := newSnapshotNode(stsm, msh, hsh, snapshotEntry.rootHash, snapshotEntry.missingNodesChan)
	if err != nil {
		snapshotEntry.iteratorChannels.ErrChan.WriteInChanNonBlocking(err)
		treatSnapshotError(err,
			"trie storage manager: newSnapshotNode takeSnapshot",
			snapshotEntry.rootHash,
			snapshotEntry.mainTrieRootHash,
		)
		return
	}

	stats := statistics.NewTrieStatistics()
	err = newRoot.commitSnapshot(stsm, snapshotEntry.iteratorChannels.LeavesChan, snapshotEntry.missingNodesChan, ctx, stats, tsm.idleProvider, rootDepthLevel)
	if err != nil {
		snapshotEntry.iteratorChannels.ErrChan.WriteInChanNonBlocking(err)
		treatSnapshotError(err,
			"trie storage manager: takeSnapshot commit",
			snapshotEntry.rootHash,
			snapshotEntry.mainTrieRootHash,
		)
		return
	}

	stats.AddAccountInfo(snapshotEntry.address, snapshotEntry.rootHash)
	snapshotEntry.stats.AddTrieStats(stats, getTrieTypeFromAddress(snapshotEntry.address))
}

func getTrieTypeFromAddress(address string) common.TrieType {
	if len(address) == 0 {
		return common.MainTrie
	}

	return common.DataTrie
}

func (tsm *trieStorageManager) takeCheckpoint(checkpointEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher, ctx context.Context, goRoutinesThrottler core.Throttler) {
	defer func() {
		tsm.finishOperation(checkpointEntry, "trie checkpoint finished")
		goRoutinesThrottler.EndProcessing()
	}()

	log.Trace("trie checkpoint started", "rootHash", checkpointEntry.rootHash)

	newRoot, err := newSnapshotNode(tsm, msh, hsh, checkpointEntry.rootHash, checkpointEntry.missingNodesChan)
	if err != nil {
		checkpointEntry.iteratorChannels.ErrChan.WriteInChanNonBlocking(err)
		treatSnapshotError(err,
			"trie storage manager: newSnapshotNode takeCheckpoint",
			checkpointEntry.rootHash,
			checkpointEntry.mainTrieRootHash,
		)
		return
	}

	stats := statistics.NewTrieStatistics()
	err = newRoot.commitCheckpoint(tsm, tsm.checkpointsStorer, tsm.checkpointHashesHolder, checkpointEntry.iteratorChannels.LeavesChan, ctx, stats, tsm.idleProvider, rootDepthLevel)
	if err != nil {
		checkpointEntry.iteratorChannels.ErrChan.WriteInChanNonBlocking(err)
		treatSnapshotError(err,
			"trie storage manager: takeCheckpoint commit",
			checkpointEntry.rootHash,
			checkpointEntry.mainTrieRootHash,
		)
		return
	}

	stats.AddAccountInfo(checkpointEntry.address, checkpointEntry.rootHash)
	checkpointEntry.stats.AddTrieStats(stats, getTrieTypeFromAddress(checkpointEntry.address))
}

func treatSnapshotError(err error, message string, rootHash []byte, mainTrieRootHash []byte) {
	if core.IsClosingError(err) {
		log.Debug("context closing", "message", message, "rootHash", rootHash, "mainTrieRootHash", mainTrieRootHash)
		return
	}

	log.Error(message, "rootHash", rootHash, "mainTrieRootHash", mainTrieRootHash, "err", err.Error())
}

func newSnapshotNode(
	db common.TrieStorageInteractor,
	msh marshal.Marshalizer,
	hsh hashing.Hasher,
	rootHash []byte,
	missingNodesCh chan []byte,
) (snapshotNode, error) {
	newRoot, err := getNodeFromDBAndDecode(rootHash, db, msh, hsh)
	if err != nil {
		if strings.Contains(err.Error(), core.GetNodeFromDBErrorString) {
			treatCommitSnapshotError(err, rootHash, missingNodesCh)
		}
		return nil, err
	}

	return newRoot, nil
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

// AddDirtyCheckpointHashes adds the given hashes to the checkpoint hashes holder
func (tsm *trieStorageManager) AddDirtyCheckpointHashes(rootHash []byte, hashes common.ModifiedHashes) bool {
	return tsm.checkpointHashesHolder.Put(rootHash, hashes)
}

// Remove removes the given hash form the storage and from the checkpoint hashes holder
func (tsm *trieStorageManager) Remove(hash []byte) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.checkpointHashesHolder.Remove(hash)
	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return tsm.mainStorer.Remove(hash)
	}

	return storer.RemoveFromCurrentEpoch(hash)
}

// RemoveFromAllActiveEpochs removes the given hash from all epochs
func (tsm *trieStorageManager) RemoveFromAllActiveEpochs(hash []byte) error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.checkpointHashesHolder.Remove(hash)
	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return fmt.Errorf("trie storage manager: main storer does not implement snapshotPruningStorer interface: %T", tsm.mainStorer)
	}

	return storer.RemoveFromAllActiveEpochs(hash)
}

// RemoveFromCheckpointHashesHolder removes the given hash from the checkpointHashesHolder
func (tsm *trieStorageManager) RemoveFromCheckpointHashesHolder(hash []byte) {
	//TODO check if the mutex is really needed here
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	log.Trace("trie storage manager: RemoveFromCheckpointHashesHolder", "hash", hash)

	tsm.checkpointHashesHolder.Remove(hash)
}

// IsClosed returns true if the trie storage manager has been closed
func (tsm *trieStorageManager) IsClosed() bool {
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

	// calling close on the SafeCloser instance should be the last instruction called
	// (just to close some go routines started as edge cases that would otherwise hang)
	defer tsm.closer.Close()

	var err error

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

	if isTrieSynced(stsm) {
		return false
	}

	return isActiveDB(stsm)
}

func isActiveDB(stsm *snapshotTrieStorageManager) bool {
	val, err := stsm.Get([]byte(common.ActiveDBKey))
	if err != nil {
		log.Debug("isActiveDB get error", "err", err.Error())
		return false
	}

	if bytes.Equal(val, []byte(common.ActiveDBVal)) {
		log.Debug("isActiveDB true")
		return true
	}

	log.Debug("isActiveDB invalid value", "value", val)
	return false
}

func isTrieSynced(stsm *snapshotTrieStorageManager) bool {
	val, err := stsm.GetFromCurrentEpoch([]byte(common.TrieSyncedKey))
	if err != nil {
		log.Debug("isTrieSynced get error", "err", err.Error())
		return false
	}

	if bytes.Equal(val, []byte(common.TrieSyncedVal)) {
		log.Debug("isTrieSynced true")
		return true
	}

	log.Debug("isTrieSynced invalid value", "value", val)
	return false
}

// GetBaseTrieStorageManager returns the trie storage manager
func (tsm *trieStorageManager) GetBaseTrieStorageManager() common.StorageManager {
	return tsm
}

// GetIdentifier returns the identifier of the main storer
func (tsm *trieStorageManager) GetIdentifier() string {
	return tsm.identifier
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsm *trieStorageManager) IsInterfaceNil() bool {
	return tsm == nil
}
