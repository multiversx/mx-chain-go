package storageManagerMock

import (
<<<<<<< HEAD:testscommon/storageManager/storageManagerArgs.go
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
=======
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
>>>>>>> rc/v1.5.0:testscommon/storage/storageManagerArgs.go
)

// GetStorageManagerArgsAndOptions returns mock args for trie storage manager creation
func GetStorageManagerArgsAndOptions() (trie.NewTrieStorageManagerArgs, trie.StorageManagerOptions) {
	storageManagerArgs := trie.NewTrieStorageManagerArgs{
		MainStorer:        genericMocks.NewStorerMock(),
		CheckpointsStorer: genericMocks.NewStorerMock(),
		Marshalizer:       &testscommon.MarshalizerMock{},
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
