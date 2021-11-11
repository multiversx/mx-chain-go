package trie

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
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
}

type snapshotsQueueEntry struct {
	rootHash   []byte
	newDb      bool
	leavesChan chan core.KeyValueHolder
}

// NewTrieStorageManagerArgs holds the arguments needed for creating a new trieStorageManager
type NewTrieStorageManagerArgs struct {
	MainStorer             common.DBWriteCacher
	CheckpointsStorer      common.DBWriteCacher
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	GeneralConfig          config.TrieStorageManagerConfig
	CheckpointHashesHolder CheckpointHashesHolder
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

	ctx, cancelFunc := context.WithCancel(context.Background())

	tsm := &trieStorageManager{
		mainStorer:             args.MainStorer,
		checkpointsStorer:      args.CheckpointsStorer,
		snapshotReq:            make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		checkpointReq:          make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		pruningBlockingOps:     0,
		cancelFunc:             cancelFunc,
		checkpointHashesHolder: args.CheckpointHashesHolder,
		closer:                 closing.NewSafeChanCloser(),
	}

	go tsm.doCheckpointsAndSnapshots(ctx, args.Marshalizer, args.Hasher)
	return tsm, nil
}

func (tsm *trieStorageManager) doCheckpointsAndSnapshots(ctx context.Context, msh marshal.Marshalizer, hsh hashing.Hasher) {
	tsm.doProcessLoop(ctx, msh, hsh)
	tsm.cleanupChans()
}

func (tsm *trieStorageManager) doProcessLoop(ctx context.Context, msh marshal.Marshalizer, hsh hashing.Hasher) {
	for {
		select {
		case snapshotRequest := <-tsm.snapshotReq:
			tsm.takeSnapshot(snapshotRequest, msh, hsh, ctx)
		case snapshotRequest := <-tsm.checkpointReq:
			tsm.takeCheckpoint(snapshotRequest, msh, hsh, ctx)
		case <-ctx.Done():
			log.Debug("trieStorageManager.storageProcessLoop go routine is closing...")
			return
		}
	}
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

//Get checks all the storers for the given key, and returns it if it is found
func (tsm *trieStorageManager) Get(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	val, _ := tsm.mainStorer.Get(key)
	if len(val) != 0 {
		return val, nil
	}

	val, _ = tsm.checkpointsStorer.Get(key)
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

	return tsm.mainStorer.Put(key, val)
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

// TakeSnapshot creates a new snapshot, or if there is another snapshot or checkpoint in progress,
// it adds this snapshot in the queue.
func (tsm *trieStorageManager) TakeSnapshot(rootHash []byte, newDb bool, leavesChan chan core.KeyValueHolder) {
	if tsm.isClosed() {
		tsm.safelyCloseChan(leavesChan)
		return
	}

	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not snapshot an empty trie")
		tsm.safelyCloseChan(leavesChan)
		return
	}

	tsm.EnterPruningBufferingMode()
	tsm.checkpointHashesHolder.RemoveCommitted(rootHash)

	snapshotEntry := &snapshotsQueueEntry{
		rootHash:   rootHash,
		newDb:      newDb,
		leavesChan: leavesChan,
	}
	select {
	case tsm.snapshotReq <- snapshotEntry:
	case <-tsm.closer.ChanClose():
		tsm.ExitPruningBufferingMode()
		tsm.safelyCloseChan(leavesChan)
	}
}

// SetCheckpoint creates a new checkpoint, or if there is another snapshot or checkpoint in progress,
// it adds this checkpoint in the queue. The checkpoint operation creates a new snapshot file
// only if there was no snapshot done prior to this
func (tsm *trieStorageManager) SetCheckpoint(rootHash []byte, leavesChan chan core.KeyValueHolder) {
	if tsm.isClosed() {
		tsm.safelyCloseChan(leavesChan)
		return
	}

	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not set checkpoint for empty trie")
		tsm.safelyCloseChan(leavesChan)
		return
	}

	tsm.EnterPruningBufferingMode()

	checkpointEntry := &snapshotsQueueEntry{
		rootHash:   rootHash,
		newDb:      false,
		leavesChan: leavesChan,
	}
	select {
	case tsm.checkpointReq <- checkpointEntry:
	case <-tsm.closer.ChanClose():
		tsm.ExitPruningBufferingMode()
		tsm.safelyCloseChan(leavesChan)
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
}

func (tsm *trieStorageManager) takeSnapshot(snapshotEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher, ctx context.Context) {
	defer tsm.finishOperation(snapshotEntry, "trie snapshot finished")

	log.Trace("trie snapshot started", "rootHash", snapshotEntry.rootHash)

	newRoot, err := newSnapshotNode(tsm, msh, hsh, snapshotEntry.rootHash)
	if err != nil {
		log.Error("takeSnapshot: trie storage manager: newSnapshotTrie", "hash", snapshotEntry.rootHash, "error", err.Error())
		return
	}

	stsm, err := newSnapshotTrieStorageManager(tsm)
	if err != nil {
		log.Error("takeSnapshot: trie storage manager: newSnapshotTrieStorageManager", "err", err.Error())
		return
	}

	err = newRoot.commitSnapshot(stsm, snapshotEntry.leavesChan, ctx)
	if err == ErrContextClosing {
		log.Debug("context closing while in commitSnapshot operation")
		return
	}
	if err != nil {
		log.Error("trie storage manager: commit", "error", err.Error())
	}
}

func (tsm *trieStorageManager) takeCheckpoint(checkpointEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher, ctx context.Context) {
	defer tsm.finishOperation(checkpointEntry, "trie checkpoint finished")

	log.Trace("trie checkpoint started", "rootHash", checkpointEntry.rootHash)

	newRoot, err := newSnapshotNode(tsm, msh, hsh, checkpointEntry.rootHash)
	if err != nil {
		log.Error("takeCheckpoint: trie storage manager: newSnapshotTrie", "hash", checkpointEntry.rootHash, "error", err.Error())
		return
	}

	err = newRoot.commitCheckpoint(tsm, tsm.checkpointsStorer, tsm.checkpointHashesHolder, checkpointEntry.leavesChan, ctx)
	if err == ErrContextClosing {
		log.Debug("context closing while in commitCheckpoint operation")
		return
	}
	if err != nil {
		log.Error("trie storage manager: commit", "error", err.Error())
	}
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

// IsInterfaceNil returns true if there is no value under the interface
func (tsm *trieStorageManager) IsInterfaceNil() bool {
	return tsm == nil
}
