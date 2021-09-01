package trie

import (
	"path"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

func (ts *trieSyncer) trieNodeIntercepted(hash []byte, val interface{}) {
	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	log.Trace("trie node intercepted", "hash", hash)

	n, ok := ts.nodesForTrie[string(hash)]
	if !ok || n.received {
		return
	}

	interceptedNode, err := trieNode(val, marshalizer, hasher)
	if err != nil {
		return
	}

	ts.nodesForTrie[string(hash)] = trieNodeInfo{
		trieNode: interceptedNode,
		received: true,
	}
}

func (tsm *trieStorageManagerWithoutCheckpoints) PruningBlockingOperations() uint32 {
	return tsm.pruningBlockingOps
}

func WaitForOperationToComplete(tsm common.StorageManager) {
	for tsm.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}
}

// GetSnapshotThatContainsHash returns the snapshot that contains the given hash
func (tsm *trieStorageManager) GetSnapshotThatContainsHash(rootHash []byte) common.SnapshotDbHandler {
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

func (tsm *trieStorageManager) NewSnapshotDb() (storage.Persister, error) {
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

func (tsm *trieStorageManager) GetSnapshots() []common.SnapshotDbHandler {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	return tsm.snapshots
}

func (tsm *trieStorageManager) SnapshotId() int {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	return tsm.snapshotId
}

func (tsm *trieStorageManager) GetFromCheckpoint(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	return tsm.checkpointsStorer.Get(key)
}

func CreateSmallTestTrieAndStorageManager() (*patriciaMerkleTrie, *trieStorageManager) {
	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()

	return tr, trieStorage
}

func GetDirtyHashes(tr common.Trie) common.ModifiedHashes {
	hashes, _ := tr.GetAllHashes()
	dirtyHashes := make(common.ModifiedHashes)
	for _, hash := range hashes {
		dirtyHashes[string(hash)] = struct{}{}
	}

	return dirtyHashes
}
