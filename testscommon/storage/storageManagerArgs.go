package storage

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
)

// GetStorageManagerArgsAndOptions returns mock args for trie storage manager creation
func GetStorageManagerArgsAndOptions() (trie.NewTrieStorageManagerArgs, trie.StorageManagerOptions) {
	storageManagerArgs := trie.NewTrieStorageManagerArgs{
		MainStorer:        genericMocks.NewStorerMock(),
		CheckpointsStorer: genericMocks.NewStorerMock(),
		Marshalizer:       &mock.MarshalizerMock{},
		Hasher:            &hashingMocks.HasherMock{},
		GeneralConfig: config.TrieStorageManagerConfig{
			SnapshotsGoroutineNum: 2,
		},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	options := trie.StorageManagerOptions{
		PruningEnabled:     true,
		SnapshotsEnabled:   true,
		CheckpointsEnabled: true,
	}

	return storageManagerArgs, options
}
