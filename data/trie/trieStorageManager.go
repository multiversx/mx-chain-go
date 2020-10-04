package trie

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type pruningOperation byte

const (
	cancelPrune pruningOperation = 0
	prune       pruningOperation = 1
)

// trieStorageManager manages all the storage operations of the trie (commit, snapshot, checkpoint, pruning)
type trieStorageManager struct {
	db data.DBWriteCacher

	snapshots          []data.SnapshotDbHandler
	snapshotId         int
	snapshotDbCfg      config.DBConfig
	snapshotReq        chan *snapshotsQueueEntry
	pruningBuffer      atomicBuffer
	snapshotInProgress uint32
	maxSnapshots       uint8
	cancelFunc         context.CancelFunc

	dbEvictionWaitingList data.DBRemoveCacher
	storageOperationMutex sync.RWMutex
}

type snapshotsQueueEntry struct {
	rootHash []byte
	newDb    bool
}

// NewTrieStorageManager creates a new instance of trieStorageManager
func NewTrieStorageManager(
	db data.DBWriteCacher,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	snapshotDbCfg config.DBConfig,
	ewl data.DBRemoveCacher,
	generalConfig config.TrieStorageManagerConfig,
) (*trieStorageManager, error) {
	if check.IfNil(db) {
		return nil, ErrNilDatabase
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(ewl) {
		return nil, ErrNilEvictionWaitingList
	}

	snapshots, snapshotId, err := getSnapshotsAndSnapshotId(snapshotDbCfg)
	if err != nil {
		log.Debug("get snapshot", "error", err.Error())
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	tsm := &trieStorageManager{
		db:                    db,
		snapshots:             snapshots,
		snapshotId:            snapshotId,
		snapshotDbCfg:         snapshotDbCfg,
		pruningBuffer:         newPruningBuffer(generalConfig.PruningBufferLen),
		dbEvictionWaitingList: ewl,
		snapshotReq:           make(chan *snapshotsQueueEntry, generalConfig.SnapshotsBufferLen),
		snapshotInProgress:    0,
		maxSnapshots:          generalConfig.MaxSnapshots,
		cancelFunc:            cancelFunc,
	}

	go tsm.storageProcessLoop(ctx, marshalizer, hasher)
	return tsm, nil
}

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

// EnterSnapshotMode sets the snapshot mode on
func (tsm *trieStorageManager) EnterSnapshotMode() {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	tsm.snapshotInProgress++

	log.Trace("enter snapshot mode", "snapshots in progress", tsm.snapshotInProgress)
}

// ExitSnapshotMode sets the snapshot mode off
func (tsm *trieStorageManager) ExitSnapshotMode() {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	if tsm.snapshotInProgress < 1 {
		log.Error("ExitSnapshotMode called too many times")
	}

	if tsm.snapshotInProgress > 0 {
		tsm.snapshotInProgress--
	}

	log.Trace("exit snapshot mode", "snapshots in progress", tsm.snapshotInProgress)
}

// Prune removes the given hash from db
func (tsm *trieStorageManager) Prune(rootHash []byte, identifier data.TriePruningIdentifier) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	rootHash = append(rootHash, byte(identifier))

	if tsm.snapshotInProgress > 0 {
		if identifier == data.NewRoot {
			tsm.cancelPrune(rootHash)
			return
		}

		rootHash = append(rootHash, byte(prune))
		tsm.pruningBuffer.add(rootHash)

		return
	}

	oldHashes := tsm.pruningBuffer.removeAll()
	tsm.resolveBufferedHashes(oldHashes)
	tsm.prune(rootHash)
}

func (tsm *trieStorageManager) resolveBufferedHashes(oldHashes [][]byte) {
	for _, rootHash := range oldHashes {
		lastBytePos := len(rootHash) - 1
		if lastBytePos < 0 {
			continue
		}

		pruneOperation := pruningOperation(rootHash[lastBytePos])
		rootHash = rootHash[:lastBytePos]

		switch pruneOperation {
		case prune:
			tsm.prune(rootHash)
		case cancelPrune:
			tsm.cancelPrune(rootHash)
		default:
			log.Error("invalid pruning operation", "operation id", pruneOperation)
		}
	}
}

func (tsm *trieStorageManager) prune(rootHash []byte) {
	log.Trace("trie storage manager prune", "root", rootHash)

	err := tsm.removeFromDb(rootHash)
	if err != nil {
		log.Error("trie storage manager remove from db", "error", err, "rootHash", hex.EncodeToString(rootHash))
	}
}

// CancelPrune removes the given hash from the eviction waiting list
func (tsm *trieStorageManager) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	rootHash = append(rootHash, byte(identifier))

	if tsm.snapshotInProgress > 0 || tsm.pruningBuffer.len() != 0 {
		rootHash = append(rootHash, byte(cancelPrune))
		tsm.pruningBuffer.add(rootHash)

		return
	}

	tsm.cancelPrune(rootHash)
}

func (tsm *trieStorageManager) cancelPrune(rootHash []byte) {
	log.Trace("trie storage manager cancel prune", "root", rootHash)
	_, _ = tsm.dbEvictionWaitingList.Evict(rootHash)
}

func (tsm *trieStorageManager) removeFromDb(rootHash []byte) error {
	hashes, err := tsm.dbEvictionWaitingList.Evict(rootHash)
	if err != nil {
		return err
	}

	log.Debug("trie removeFromDb", "rootHash", rootHash)

	lastBytePos := len(rootHash) - 1
	if lastBytePos < 0 {
		return ErrInvalidIdentifier
	}
	identifier := data.TriePruningIdentifier(rootHash[lastBytePos])

	var hash []byte
	var shouldKeepHash bool
	for key := range hashes {
		shouldKeepHash, err = tsm.dbEvictionWaitingList.ShouldKeepHash(key, identifier)
		if err != nil {
			return err
		}
		if shouldKeepHash {
			continue
		}

		hash, err = hex.DecodeString(key)
		if err != nil {
			return err
		}

		log.Trace("remove hash from trie db", "hash", hex.EncodeToString(hash))
		err = tsm.db.Remove(hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// MarkForEviction adds the given hashes in the eviction waiting list at the provided key
func (tsm *trieStorageManager) MarkForEviction(root []byte, hashes data.ModifiedHashes) error {
	log.Trace("trie storage manager: mark for eviction", "root", root)

	return tsm.dbEvictionWaitingList.Put(root, hashes)
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
func (tsm *trieStorageManager) TakeSnapshot(rootHash []byte) {
	tsm.EnterSnapshotMode()

	snapshotEntry := &snapshotsQueueEntry{rootHash: rootHash, newDb: true}
	tsm.writeOnChan(snapshotEntry)
}

// SetCheckpoint creates a new checkpoint, or if there is another snapshot or checkpoint in progress,
// it adds this checkpoint in the queue. The checkpoint operation creates a new snapshot file
// only if there was no snapshot done prior to this
func (tsm *trieStorageManager) SetCheckpoint(rootHash []byte) {
	tsm.EnterSnapshotMode()

	checkpointEntry := &snapshotsQueueEntry{rootHash: rootHash, newDb: false}
	tsm.writeOnChan(checkpointEntry)
}

func (tsm *trieStorageManager) writeOnChan(entry *snapshotsQueueEntry) {
	select {
	case tsm.snapshotReq <- entry:
		return
	}
}

func (tsm *trieStorageManager) takeSnapshot(snapshot *snapshotsQueueEntry, msh marshal.Marshalizer, hsh hashing.Hasher) {
	defer tsm.ExitSnapshotMode()

	if tsm.isPresentInLastSnapshotDb(snapshot.rootHash) {
		log.Trace("snapshot for rootHash already taken", "rootHash", snapshot.rootHash)
		return
	}

	log.Trace("trie snapshot started", "rootHash", snapshot.rootHash, "newDB", snapshot.newDb)

	newRoot, err := newSnapshotNode(tsm.db, msh, hsh, snapshot.rootHash)
	if err != nil {
		log.Error("trie storage manager: newSnapshotTrie", "error", err.Error())
		return
	}
	db := tsm.getSnapshotDb(snapshot.newDb)
	if check.IfNil(db) {
		return
	}

	maxTrieLevelInMemory := uint(5)
	err = newRoot.commit(true, 0, maxTrieLevelInMemory, tsm.db, db)
	if err != nil {
		log.Error("trie storage manager: commit", "error", err.Error())
		return
	}

	log.Trace("trie snapshot finished", "rootHash", snapshot.rootHash)
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

	if uint8(len(tsm.snapshots)) > tsm.maxSnapshots {
		tsm.removeSnapshot()
	}

	return db
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

func removeSnapshot(db data.DBWriteCacher, path string) {
	err := db.Close()
	if err != nil {
		log.Error("trie storage manager: removeSnapshot", "error", err.Error())
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

	trieStorage := &trieStorageManager{
		db: db,
	}

	snapshotPmt := &patriciaMerkleTrie{
		root:        newRoot,
		trieStorage: trieStorage,
		marshalizer: msh,
		hasher:      hsh,
	}

	return snapshotPmt.root, nil
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

// GetSnapshotDbBatchDelay returns the batch write delay in seconds
func (tsm *trieStorageManager) GetSnapshotDbBatchDelay() int {
	return tsm.snapshotDbCfg.BatchDelaySeconds
}

// Close - closes all underlying components
func (tsm *trieStorageManager) Close() error {
	tsm.cancelFunc()

	err1 := tsm.db.Close()
	err2 := tsm.dbEvictionWaitingList.Close()

	for _, sdb := range tsm.snapshots {
		log.LogIfError(sdb.Close())
	}

	if err1 != nil || err2 != nil {
		errorStr := ""
		if err2 != nil {
			errorStr = err2.Error()
		}
		return fmt.Errorf("trieStorageManager close failed: %w , %s", err1, errorStr)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsm *trieStorageManager) IsInterfaceNil() bool {
	return tsm == nil
}
