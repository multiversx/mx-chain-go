package storage

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
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
