package storage

import (
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
<<<<<<< HEAD
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
=======
>>>>>>> rc/v1.7.0
	"github.com/multiversx/mx-chain-go/trie"
)

// GetStorageManagerArgs returns mock args for trie storage manager creation
func GetStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
<<<<<<< HEAD
		MainStorer:        testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer: testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:       &marshallerMock.MarshalizerMock{},
		Hasher:            &hashingMocks.HasherMock{},
=======
		MainStorer:  testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
>>>>>>> rc/v1.7.0
		GeneralConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 2,
		},
		IdleProvider:   &testscommon.ProcessStatusHandlerStub{},
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: disabled.NewStateStatistics(),
	}
}

// GetStorageManagerOptions returns default options for trie storage manager creation
func GetStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:   true,
		SnapshotsEnabled: true,
	}
}
