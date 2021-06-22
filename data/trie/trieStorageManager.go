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

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const (
	snapshot   byte = 0
	checkpoint byte = 1
)

// trieStorageManager manages all the storage operations of the trie (commit, snapshot, checkpoint, pruning)
type trieStorageManager struct {
	db data.DBWriteCacher

	snapshots              []data.SnapshotDbHandler
	snapshotId             int
	snapshotDbCfg          config.DBConfig
	snapshotReq            chan *snapshotsQueueEntry
	checkpointHashesHolder data.CheckpointHashesHolder

	pruningBlockingOps uint32
	maxSnapshots       uint32
	keepSnapshots      bool
	cancelFunc         context.CancelFunc

	storageOperationMutex sync.RWMutex
}

type snapshotsQueueEntry struct {
	rootHash   []byte
	newDb      bool
	entryType  byte
	leavesChan chan core.KeyValueHolder
}

// NewTrieStorageManagerArgs holds the arguments needed for creating a new trieStorageManager
type NewTrieStorageManagerArgs struct {
	DB                     data.DBWriteCacher
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	SnapshotDbConfig       config.DBConfig
	GeneralConfig          config.TrieStorageManagerConfig
	CheckpointHashesHolder data.CheckpointHashesHolder
}

// NewTrieStorageManager creates a new instance of trieStorageManager
func NewTrieStorageManager(args NewTrieStorageManagerArgs) (*trieStorageManager, error) {
	if check.IfNil(args.DB) {
		return nil, ErrNilDatabase
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

	snapshots, snapshotId, err := getSnapshotsAndSnapshotId(args.SnapshotDbConfig)
	if err != nil {
		log.Debug("get snapshot", "error", err.Error())
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	tsm := &trieStorageManager{
		db:                     args.DB,
		snapshots:              snapshots,
		snapshotId:             snapshotId,
		snapshotDbCfg:          args.SnapshotDbConfig,
		snapshotReq:            make(chan *snapshotsQueueEntry, args.GeneralConfig.SnapshotsBufferLen),
		pruningBlockingOps:     0,
		maxSnapshots:           args.GeneralConfig.MaxSnapshots,
		keepSnapshots:          args.GeneralConfig.KeepSnapshots,
		cancelFunc:             cancelFunc,
		checkpointHashesHolder: args.CheckpointHashesHolder,
	}

	go tsm.storageProcessLoop(ctx, args.Marshalizer, args.Hasher)
	return tsm, nil
}

//nolint
func (tsm *trieStorageManager) storageProcessLoop(ctx context.Context, msh marshal.Marshalizer, hsh hashing.Hasher) {
	for {
		select {
		case snapshot := <-tsm.snapshotReq:
			tsm.takeSnapshot(snapshot, msh, hsh)
		case <-ctx.Done():
			return
		}
	}
}

func getOrderedSnapshots(snapshotsMap map[int]data.SnapshotDbHandler) []data.SnapshotDbHandler {
	snapshots := make([]data.SnapshotDbHandler, 0)
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

func getSnapshotsAndSnapshotId(snapshotDbCfg config.DBConfig) ([]data.SnapshotDbHandler, int, error) {
	snapshotsMap := make(map[int]data.SnapshotDbHandler)
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

		snapshot := &snapshotDb{
			DBWriteCacher: db,
		}

		log.Debug("restored snapshot", "snapshot ID", snapshotName)
		snapshotsMap[snapshotName] = snapshot
	}

	if len(snapshotsMap) != 0 {
		snapshotId++
	}

	return getOrderedSnapshots(snapshotsMap), snapshotId, nil
}

// Database returns the main database
func (tsm *trieStorageManager) Database() data.DBWriteCacher {
	return tsm.db
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

// GetSnapshotThatContainsHash returns the snapshot that contains the given hash
func (tsm *trieStorageManager) GetSnapshotThatContainsHash(rootHash []byte) data.SnapshotDbHandler {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	for i := range tsm.snapshots {
		_, err := tsm.snapshots[i].Get(rootHash)

		hashPresent := err == nil
		if hashPresent {
			log.Trace("hash present in snapshot trie db", "hash", rootHash)
			tsm.snapshots[i].IncreaseNumReferences()
			return tsm.snapshots[i]
		}
	}

	return nil
}

// TakeSnapshot creates a new snapshot, or if there is another snapshot or checkpoint in progress,
// it adds this snapshot in the queue.
func (tsm *trieStorageManager) TakeSnapshot(rootHash []byte, newDb bool, leavesChan chan core.KeyValueHolder) {
	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not snapshot an empty trie")
		return
	}

	tsm.EnterPruningBufferingMode()
	tsm.checkpointHashesHolder.RemoveCommitted(rootHash)

	snapshotEntry := &snapshotsQueueEntry{
		rootHash:   rootHash,
		newDb:      newDb,
		entryType:  snapshot,
		leavesChan: leavesChan,
	}
	tsm.writeOnChan(snapshotEntry)
}

// SetCheckpoint creates a new checkpoint, or if there is another snapshot or checkpoint in progress,
// it adds this checkpoint in the queue. The checkpoint operation creates a new snapshot file
// only if there was no snapshot done prior to this
func (tsm *trieStorageManager) SetCheckpoint(rootHash []byte, leavesChan chan core.KeyValueHolder) {
	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not set checkpoint for empty trie")
		return
	}

	tsm.EnterPruningBufferingMode()

	checkpointEntry := &snapshotsQueueEntry{
		rootHash:   rootHash,
		newDb:      false,
		entryType:  checkpoint,
		leavesChan: leavesChan,
	}
	tsm.writeOnChan(checkpointEntry)
}

func (tsm *trieStorageManager) writeOnChan(entry *snapshotsQueueEntry) {
	tsm.snapshotReq <- entry
}

func (tsm *trieStorageManager) takeSnapshot(snapshotEntry *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher) {
	defer func() {
		tsm.ExitPruningBufferingMode()
		log.Trace("trie snapshot finished", "rootHash", snapshotEntry.rootHash)
		if snapshotEntry.leavesChan != nil {
			close(snapshotEntry.leavesChan)
		}
	}()

	if tsm.isPresentInLastSnapshotDb(snapshotEntry.rootHash) {
		log.Trace("snapshot for rootHash already taken", "rootHash", snapshotEntry.rootHash)
		return
	}

	log.Trace("trie snapshot started", "rootHash", snapshotEntry.rootHash, "newDB", snapshotEntry.newDb)

	newRoot, err := newSnapshotNode(tsm.db, msh, hsh, snapshotEntry.rootHash)
	if err != nil {
		log.Error("trie storage manager: newSnapshotTrie", "error", err.Error())
		return
	}
	db := tsm.getSnapshotDb(snapshotEntry.newDb)
	if check.IfNil(db) {
		return
	}

	if snapshotEntry.entryType == snapshot {
		err = newRoot.commitSnapshot(tsm.db, db, snapshotEntry.leavesChan)
		if err != nil {
			log.Error("trie storage manager: commit", "error", err.Error())
		}

		return
	}

	err = newRoot.commitCheckpoint(tsm.db, db, tsm.checkpointHashesHolder, snapshotEntry.leavesChan)
	if err != nil {
		log.Error("trie storage manager: commit", "error", err.Error())
	}
}

func (tsm *trieStorageManager) isPresentInLastSnapshotDb(rootHash []byte) bool {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	lastSnapshotIndex := len(tsm.snapshots) - 1
	if lastSnapshotIndex < 0 {
		return false
	}

	val, err := tsm.snapshots[lastSnapshotIndex].Get(rootHash)
	if err != nil || val == nil {
		return false
	}

	return true
}

func (tsm *trieStorageManager) getSnapshotDb(newDb bool) data.DBWriteCacher {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	createNewDb := newDb || len(tsm.snapshots) == 0
	if !createNewDb {
		return tsm.snapshots[len(tsm.snapshots)-1]
	}

	db, err := tsm.newSnapshotDb()
	if err != nil {
		log.Error("trie storage manager: getSnapshotDb", "error", err.Error())
		return nil
	}

	if uint32(len(tsm.snapshots)) > tsm.maxSnapshots {
		if tsm.keepSnapshots {
			tsm.disconnectSnapshot()
		} else {
			tsm.removeSnapshot()
		}
	}

	return db
}

func (tsm *trieStorageManager) disconnectSnapshot() {
	if len(tsm.snapshots) <= 0 {
		return
	}
	snapshot := tsm.snapshots[0]
	tsm.snapshots = tsm.snapshots[1:]

	if snapshot.IsInUse() {
		snapshot.MarkForDisconnection()
		log.Debug("can't disconnect, snapshot is still in use")
		return
	}
	err := disconnectSnapshot(snapshot)
	if err != nil {
		log.Error("trie storage manager: disconnectSnapshot", "error", err.Error())
	}
}

func (tsm *trieStorageManager) removeSnapshot() {
	if len(tsm.snapshots) <= 0 {
		return
	}

	dbUniqueId := strconv.Itoa(tsm.snapshotId - len(tsm.snapshots))

	snapshot := tsm.snapshots[0]
	tsm.snapshots = tsm.snapshots[1:]
	removePath := path.Join(tsm.snapshotDbCfg.FilePath, dbUniqueId)

	if snapshot.IsInUse() {
		log.Debug("snapshot is still in use", "path", removePath)
		snapshot.MarkForRemoval()
		snapshot.SetPath(removePath)

		return
	}

	removeSnapshot(snapshot, removePath)
}

func disconnectSnapshot(db data.DBWriteCacher) error {
	return db.Close()
}

func removeSnapshot(db data.DBWriteCacher, path string) {
	err := disconnectSnapshot(db)
	if err != nil {
		log.Error("trie storage manager: disconnectSnapshot", "error", err.Error())
		return
	}

	log.Debug("remove trie snapshot db", "snapshot path", path)
	go removeDirectory(path)
}

func removeDirectory(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Error(err.Error())
	}
}

func newSnapshotNode(
	db data.DBWriteCacher,
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

func (tsm *trieStorageManager) newSnapshotDb() (storage.Persister, error) {
	snapshotPath := path.Join(tsm.snapshotDbCfg.FilePath, strconv.Itoa(tsm.snapshotId))
	for directoryExists(snapshotPath) {
		tsm.snapshotId++
		snapshotPath = path.Join(tsm.snapshotDbCfg.FilePath, strconv.Itoa(tsm.snapshotId))
	}

	log.Debug("create new trie snapshot db", "snapshot ID", tsm.snapshotId)
	arg := storageUnit.ArgDB{
		DBType:            storageUnit.DBType(tsm.snapshotDbCfg.Type),
		Path:              snapshotPath,
		BatchDelaySeconds: tsm.snapshotDbCfg.BatchDelaySeconds,
		MaxBatchSize:      tsm.snapshotDbCfg.MaxBatchSize,
		MaxOpenFiles:      tsm.snapshotDbCfg.MaxOpenFiles,
	}
	db, err := storageUnit.NewDB(arg)
	if err != nil {
		return nil, err
	}

	tsm.snapshotId++

	snapshot := &snapshotDb{
		DBWriteCacher: db,
	}
	tsm.snapshots = append(tsm.snapshots, snapshot)

	return db, nil
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
func (tsm *trieStorageManager) AddDirtyCheckpointHashes(rootHash []byte, hashes data.ModifiedHashes) bool {
	return tsm.checkpointHashesHolder.Put(rootHash, hashes)
}

// Remove removes the given hash form the storage and from the checkpoint hashes holder
func (tsm *trieStorageManager) Remove(hash []byte) error {
	tsm.checkpointHashesHolder.Remove(hash)
	return tsm.db.Remove(hash)
}

// Close - closes all underlying components
func (tsm *trieStorageManager) Close() error {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.cancelFunc()

	err := tsm.db.Close()

	for _, sdb := range tsm.snapshots {
		log.LogIfError(sdb.Close())
	}

	if err != nil {
		return fmt.Errorf("trieStorageManager close failed: %w", err)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsm *trieStorageManager) IsInterfaceNil() bool {
	return tsm == nil
}
