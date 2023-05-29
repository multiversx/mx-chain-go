package syncer_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/stretchr/testify/require"
)

func getDefaultBaseAccSyncerArgs() syncer.ArgsNewBaseAccountsSyncer {
	return syncer.ArgsNewBaseAccountsSyncer{
		Hasher:                            &hashingMocks.HasherMock{},
		Marshalizer:                       testscommon.MarshalizerMock{},
		TrieStorageManager:                &storageManager.StorageManagerStub{},
		RequestHandler:                    &testscommon.RequestHandlerStub{},
		Timeout:                           time.Second,
		Cacher:                            testscommon.NewCacherMock(),
		UserAccountsSyncStatisticsHandler: &testscommon.SizeSyncStatisticsHandlerStub{},
		AppStatusHandler:                  &statusHandler.AppStatusHandlerStub{},
		MaxTrieLevelInMemory:              5,
		MaxHardCapForMissingNodes:         100,
		TrieSyncerVersion:                 3,
		CheckNodesOnDisk:                  false,
	}
}

func TestBaseAccountsSyncer_CheckArgs(t *testing.T) {
	t.Parallel()

	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.Hasher = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.Marshalizer = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilMarshalizer, err)
	})

	t.Run("nil trie storage manager", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.TrieStorageManager = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilStorageManager, err)
	})

	t.Run("nil requests handler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.RequestHandler = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilRequestHandler, err)
	})

	t.Run("nil cacher", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.Cacher = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilCacher, err)
	})

	t.Run("nil user accounts sync statistics handler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.UserAccountsSyncStatisticsHandler = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilSyncStatisticsHandler, err)
	})

	t.Run("nil app status handler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.AppStatusHandler = nil
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrNilAppStatusHandler, err)
	})

	t.Run("invalid max hard capacity for missing nodes", func(t *testing.T) {
		t.Parallel()

		args := getDefaultBaseAccSyncerArgs()
		args.MaxHardCapForMissingNodes = 0
		err := syncer.CheckBaseAccountsSyncerArgs(args)
		require.Equal(t, state.ErrInvalidMaxHardCapForMissingNodes, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, syncer.CheckBaseAccountsSyncerArgs(getDefaultBaseAccSyncerArgs()))
	})
}
