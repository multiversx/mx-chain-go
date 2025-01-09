package state

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	disabled2 "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/trie"
)

// GetDefaultTrieParameters -
func GetDefaultTrieParameters() (common.StorageManager, marshal.Marshalizer, hashing.Hasher) {
	db := testscommon.NewMemDbMock()
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}

	tsmArgs := trie.NewTrieStorageManagerArgs{
		MainStorer:  db,
		Marshalizer: marshaller,
		Hasher:      hasher,
		GeneralConfig: config.TrieStorageManagerConfig{
			SnapshotsGoroutineNum: 5,
		},
		IdleProvider:   disabled2.NewProcessStatusHandler(),
		Identifier:     "identifier",
		StatsCollector: disabled.NewStateStatistics(),
	}
	tsm, _ := trie.NewTrieStorageManager(tsmArgs)
	return tsm, marshaller, hasher
}

// GetNewTrie -
func GetNewTrie() common.Trie {
	tsm, marshaller, hasher := GetDefaultTrieParameters()
	tr, _ := trie.NewTrie(tsm, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 5)

	return tr
}

// AddDataToTrie -
func AddDataToTrie(tr common.Trie, numLeaves int) {
	for i := 0; i < numLeaves; i++ {
		val := fmt.Sprintf("value%v", i)
		_ = tr.Update([]byte(val), []byte(val))
	}
	_ = tr.Commit()
}

// GetTrieWithData returns a trie with some data.
// The added data builds a rootNode that is a branch with 2 leaves and 1 extension node which will have 4 leaves when traversed;
// this way the size of the iterator will be highest when the extension node is reached but 2 leaves will
// have already been retrieved
func GetTrieWithData() common.Trie {
	tr := GetNewTrie()
	_ = tr.Update([]byte("key1"), []byte("value1"))
	_ = tr.Update([]byte("key2"), []byte("value2"))
	_ = tr.Update([]byte("key13"), []byte("value3"))
	_ = tr.Update([]byte("key23"), []byte("value4"))
	_ = tr.Update([]byte("key33"), []byte("value4"))
	_ = tr.Update([]byte("key43"), []byte("value4"))
	_ = tr.Commit()
	return tr
}
