package storage

import (
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
)

// GetStorageManagerArgs returns mock args for trie storage manager creation
func GetStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:        testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer: testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:       &mock.MarshalizerMock{},
		Hasher:            &hashingMocks.HasherMock{},
		GeneralConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 2,
		},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
		Identifier:             dataRetriever.UserAccountsUnit.String(),
		StatsCollector:         disabled.NewStateStatistics(),
	}
}

// GetStorageManagerOptions returns default options for trie storage manager creation
func GetStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:     true,
		SnapshotsEnabled:   true,
		CheckpointsEnabled: true,
	}
}
