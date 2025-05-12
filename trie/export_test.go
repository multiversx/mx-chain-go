package trie

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
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
func NewBaseIterator(trie common.Trie, rootHash []byte) (*baseIterator, error) {
	return newBaseIterator(trie, rootHash)
}

// GetDefaultTrieStorageManagerParameters -
func GetDefaultTrieStorageManagerParameters() NewTrieStorageManagerArgs {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}

	return NewTrieStorageManagerArgs{
		MainStorer:     testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:    &marshal.GogoProtoMarshalizer{},
		Hasher:         &testscommon.KeccakMock{},
		GeneralConfig:  generalCfg,
		IdleProvider:   &testscommon.ProcessStatusHandlerStub{},
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: statistics.NewStateStatistics(),
	}
}

// ExecuteUpdatesFromBatch -
func ExecuteUpdatesFromBatch(tr common.Trie) {
	pmt, _ := tr.(*patriciaMerkleTrie)
	_ = pmt.updateTrie()
}

// GetBatchManager -
func GetBatchManager(tr common.Trie) common.TrieBatchManager {
	return tr.(*patriciaMerkleTrie).batchManager
}

// SetGoRoutinesManager -
func SetGoRoutinesManager(tr common.Trie, gm common.TrieGoroutinesManager) {
	pmt, _ := tr.(*patriciaMerkleTrie)
	pmt.goRoutinesManager = gm
}
