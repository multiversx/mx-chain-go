package storageManagerMock

import (
	"github.com/multiversx/mx-chain-go/config"
<<<<<<< HEAD:testscommon/storageManager/storageManagerArgs.go
=======
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/genesis/mock"
>>>>>>> rc/v1.6.0:testscommon/storage/storageManagerArgs.go
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
)

<<<<<<< HEAD:testscommon/storageManager/storageManagerArgs.go
// GetStorageManagerArgsAndOptions returns mock args for trie storage manager creation
func GetStorageManagerArgsAndOptions() (trie.NewTrieStorageManagerArgs, trie.StorageManagerOptions) {
	storageManagerArgs := trie.NewTrieStorageManagerArgs{
		MainStorer:        genericMocks.NewStorerMock(),
		CheckpointsStorer: genericMocks.NewStorerMock(),
		Marshalizer:       &testscommon.MarshalizerMock{},
=======
// GetStorageManagerArgs returns mock args for trie storage manager creation
func GetStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:        testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer: testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:       &mock.MarshalizerMock{},
>>>>>>> rc/v1.6.0:testscommon/storage/storageManagerArgs.go
		Hasher:            &hashingMocks.HasherMock{},
		GeneralConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 2,
		},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
		Identifier:             dataRetriever.UserAccountsUnit.String(),
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
