package factory

import (
	"os"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/databaseremover/disabled"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
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
	require.IsType(t, pruning.NewImportDBPeerPruningStorer(0, nil), storer)
	require.NoError(t, storer.Close())
	args.StorageType = ImportDBStorageService
	factory, err = NewStorageServiceFactory(args)
	require.NoError(t, err)
	storer, err = factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	_, memory := storer.(storageMocks.ImportDBPeerStorer)
	require.False(t, memory, "import source trie storer must remain disabled")
}

func TestImportDBPeerFactoryReadsExistingDiskState(t *testing.T) {
	args := importPeerFactoryArgs(t)
	args.CurrentEpoch = 7
	args.Config.PeerAccountsTrieStorage.DB.Type = "LvlDBSerial"
	args.Config.StateTriesConfig.ImportDBPeerSnapshotRoundsInterval = 0
	factory, err := NewStorageServiceFactory(args)
	require.NoError(t, err)
	disk, err := factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	require.NoError(t, disk.Put([]byte("starting-root"), []byte("persisted-node")))
	require.NoError(t, disk.Close())

	args.Config.PeerAccountsTrieStorage.DB.Type = "MemoryDB"
	args.Config.StateTriesConfig.ImportDBPeerSnapshotRoundsInterval = 100
	factory, err = NewStorageServiceFactory(args)
	require.NoError(t, err)
	memory, err := factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, memory.Close()) })
	value, err := memory.Get([]byte("starting-root"))
	require.NoError(t, err, "startup must read the normal disk DB before its first snapshot")
	require.Equal(t, "persisted-node", string(value))
	require.NoError(t, memory.Put([]byte("memory-only"), []byte("new-node")))
	memory.(storageMocks.ImportDBPeerStorer).SnapshotFinished()

	// Closing the startup fallback must leave the normal DB intact and unlocked.
	args.Config.PeerAccountsTrieStorage.DB.Type = "LvlDBSerial"
	args.Config.StateTriesConfig.ImportDBPeerSnapshotRoundsInterval = 0
	factory, err = NewStorageServiceFactory(args)
	require.NoError(t, err)
	disk, err = factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	defer func() { require.NoError(t, disk.Close()) }()
	value, err = disk.Get([]byte("starting-root"))
	require.NoError(t, err)
	require.Equal(t, "persisted-node", string(value))
	_, err = disk.Get([]byte("memory-only"))
	require.Error(t, err, "replay writes must never reach the disk source")
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
			require.ErrorIs(t, err, storage.ErrInvalidConfig)
		})
	}
}

func TestImportDBPeerFactoryBoundsBootstrapEpochScan(t *testing.T) {
	args := importPeerFactoryArgs(t)
	args.CurrentEpoch = 12
	args.Config.DbLookupExtensions.Enabled = false
	args.Config.StoragePruning.NumActivePersisters = 4
	args.Config.StoragePruning.NumEpochsToKeep = 4
	args.Config.PeerAccountsTrieStorage.DB.Type = "LvlDBSerial"
	args.Config.PeerAccountsTrieStorage.DB.ShardIDProviderType = "BinarySplit"
	args.Config.PeerAccountsTrieStorage.DB.NumShards = 4
	args.Config.StateTriesConfig.ImportDBPeerSnapshotRoundsInterval = 0
	factory, err := NewStorageServiceFactory(args)
	require.NoError(t, err)
	disk, err := factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	require.NoError(t, disk.PutInEpoch([]byte("current-root"), []byte("current-value"), 12))
	require.NoError(t, disk.PutInEpoch([]byte("older-node"), []byte("older-value"), 10))
	require.NoError(t, disk.Close())
	untouched := args.PathManager.PathForEpoch(core.GetShardIDString(core.MetachainShardId), 8, args.Config.PeerAccountsTrieStorage.DB.FilePath)
	_, err = os.Stat(untouched)
	require.ErrorIs(t, err, os.ErrNotExist)
	args.Config.PeerAccountsTrieStorage.DB.Type = "MemoryDB"
	args.Config.StateTriesConfig.ImportDBPeerSnapshotRoundsInterval = 200
	factory, err = NewStorageServiceFactory(args)
	require.NoError(t, err)
	memory, err := factory.createPeerTrieStorer(disabled.NewDisabledCustomDatabaseRemover())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, memory.Close()) })
	for key, expected := range map[string]string{"current-root": "current-value", "older-node": "older-value"} {
		value, readErr := memory.Get([]byte(key))
		require.NoError(t, readErr)
		require.Equal(t, expected, string(value))
	}
	_, err = os.Stat(untouched)
	require.ErrorIs(t, err, os.ErrNotExist, "startup must not create historical peer DBs outside the configured window")
}
