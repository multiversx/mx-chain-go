package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage/databaseremover/disabled"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/stretchr/testify/require"
)

func importPeerFactoryArgs(t *testing.T) StorageServiceFactoryArgs {
	args := createMockArgument(t)
	args.NodeProcessingMode = common.ImportDb
	args.ShardCoordinator = &mock.ShardCoordinatorMock{SelfShardId: core.MetachainShardId, NumShards: 3}
	args.Config.StateTriesConfig.ImportDBPeerSnapshotRoundsInterval = 100
	args.Config.StateTriesConfig.SnapshotsEnabled = true
	args.Config.StateTriesConfig.PeerStatePruningEnabled = false
	args.Config.PeerAccountsTrieStorage.DB.Type = "MemoryDB"
	return args
}

func TestImportDBPeerFactorySelectsProcessingMemoryOnly(t *testing.T) {
	args := importPeerFactoryArgs(t)
	factory, err := NewStorageServiceFactory(args)
	require.NoError(t, err)
	storer, err := factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	require.IsType(t, pruning.NewImportDBPeerPruningStorer(0), storer)
	require.NoError(t, storer.Close())
	args.StorageType = ImportDBStorageService
	factory, err = NewStorageServiceFactory(args)
	require.NoError(t, err)
	storer, err = factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	_, memory := storer.(interface {
		Rotate()
		SnapshotFinished()
	})
	require.False(t, memory, "import source trie storer must remain disabled")
}

func TestImportDBPeerFactoryRejectsIncompatibleSettings(t *testing.T) {
	for _, name := range []string{"normal mode", "shard", "disk", "snapshots disabled", "peer pruning", "archive"} {
		t.Run(name, func(t *testing.T) {
			args := importPeerFactoryArgs(t)
			switch name {
			case "normal mode":
				args.NodeProcessingMode = common.Normal
			case "shard":
				args.ShardCoordinator = &mock.ShardCoordinatorMock{NumShards: 3}
			case "disk":
				args.Config.PeerAccountsTrieStorage.DB.Type = "LvlDBSerial"
			case "snapshots disabled":
				args.Config.StateTriesConfig.SnapshotsEnabled = false
			case "peer pruning":
				args.Config.StateTriesConfig.PeerStatePruningEnabled = true
			case "archive":
				args.PrefsConfig.FullArchive = true
			}
			_, err := NewStorageServiceFactory(args)
			require.Error(t, err)
		})
	}
}
