package trie

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie/hashesHolder"
)

func (ts *trieSyncer) trieNodeIntercepted(hash []byte, val interface{}) {
	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	log.Trace("trie node intercepted", "hash", hash)

	n, ok := ts.nodesForTrie[string(hash)]
	if !ok || n.received {
		return
	}

	interceptedNode, err := trieNode(val, marshalizer, hasherMock)
	if err != nil {
		return
	}

	ts.nodesForTrie[string(hash)] = trieNodeInfo{
		trieNode: interceptedNode,
		received: true,
	}
}

// PruningBlockingOperations -
func (tsm *trieStorageManagerWithoutCheckpoints) PruningBlockingOperations() uint32 {
	ts, _ := tsm.StorageManager.(*trieStorageManager)
	return ts.pruningBlockingOps
}

// WaitForOperationToComplete -
func WaitForOperationToComplete(tsm common.StorageManager) {
	for tsm.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}
}

// GetFromCheckpoint -
func (tsm *trieStorageManager) GetFromCheckpoint(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	return tsm.checkpointsStorer.Get(key)
}

// CreateSmallTestTrieAndStorageManager -
func CreateSmallTestTrieAndStorageManager() (*patriciaMerkleTrie, *trieStorageManager) {
	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()

	return tr, trieStorage
}

// GetDirtyHashes -
func GetDirtyHashes(tr common.Trie) common.ModifiedHashes {
	hashes, _ := tr.GetAllHashes()
	dirtyHashes := make(common.ModifiedHashes)
	for _, hash := range hashes {
		dirtyHashes[string(hash)] = struct{}{}
	}

	return dirtyHashes
}

// WriteInChanNonBlocking -
func WriteInChanNonBlocking(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}

type StorageManagerExtensionStub struct {
	*storageManager.StorageManagerStub
}

// IsBaseTrieStorageManager -
func IsBaseTrieStorageManager(tsm common.StorageManager) bool {
	_, ok := tsm.(*trieStorageManager)
	return ok
}

// IsInEpochTrieStorageManager -
func IsTrieStorageManagerInEpoch(tsm common.StorageManager) bool {
	_, ok := tsm.(*trieStorageManagerInEpoch)
	return ok
}

// NewBaseIterator -
func NewBaseIterator(trie common.Trie) (*baseIterator, error) {
	return newBaseIterator(trie)
}

// GetDefaultTrieStorageManagerParameters -
func GetDefaultTrieStorageManagerParameters() NewTrieStorageManagerArgs {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}

	return NewTrieStorageManagerArgs{
		MainStorer:             testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer:      testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:            &marshal.GogoProtoMarshalizer{},
		Hasher:                 &testscommon.KeccakMock{},
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
		Identifier:             dataRetriever.UserAccountsUnit.String(),
		StatsCollector:         statistics.NewStateStatistics(),
	}
}
