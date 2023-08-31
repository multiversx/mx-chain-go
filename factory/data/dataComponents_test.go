package data_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewDataComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.ShardCoordinator = nil

		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.Nil(t, dcf)
		require.Equal(t, errorsMx.ErrNilShardCoordinator, err)
	})
	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(nil, runTypeComponents, shardCoordinator)
		args.Core = nil

		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.Nil(t, dcf)
		require.Equal(t, errorsMx.ErrNilCoreComponents, err)
	})
	t.Run("nil status core components should error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.StatusCore = nil

		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.Nil(t, dcf)
		require.Equal(t, errorsMx.ErrNilStatusCoreComponents, err)
	})
	t.Run("nil crypto components should error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Crypto = nil

		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.Nil(t, dcf)
		require.Equal(t, errorsMx.ErrNilCryptoComponents, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, dcf)
	})
}

func TestDataComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("NewBlockChain returns error for shard", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.StatusCore = &factory.StatusCoreComponentsStub{
			AppStatusHandlerField: nil,
		}
		args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.Error(t, err)
		require.Nil(t, dc)
	})
	t.Run("NewBlockChain returns error for meta", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = core.MetachainShardId
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.StatusCore = &factory.StatusCoreComponentsStub{
			AppStatusHandlerField: nil,
		}
		args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.Error(t, err)
		require.Nil(t, dc)
	})
	t.Run("createBlockChainFromConfig returns error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = 12345
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.Equal(t, errorsMx.ErrBlockchainCreation, err)
		require.Nil(t, dc)
	})
	t.Run("NewStorageServiceFactory returns error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Config.StoragePruning.NumActivePersisters = 0
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.Error(t, err)
		require.Nil(t, dc)
	})
	t.Run("createDataStoreFromConfig fails for shard due to bad config", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.Error(t, err)
		require.Nil(t, dc)
	})
	t.Run("createDataStoreFromConfig fails, invalid shard", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		cnt := 0
		shardCoordinator.SelfIDCalled = func() uint32 {
			cnt++
			if cnt > 1 {
				return 12345
			}
			return 0
		}
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.Equal(t, errorsMx.ErrDataStoreCreation, err)
		require.Nil(t, dc)
	})
	t.Run("NewDataPoolFromConfig fails should error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		coreComponents := componentsMock.GetCoreComponents()
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Config.TxBlockBodyDataPool.Type = "invalid"
		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)

		dc, err := dcf.Create()
		require.True(t, errors.Is(err, errorsMx.ErrDataPoolCreation))
		require.Nil(t, dc)
	})
	t.Run("should work for shard", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		dcf, err := dataComp.NewDataComponentsFactory(args)

		require.NoError(t, err)
		dc, err := dcf.Create()
		require.NoError(t, err)
		require.NotNil(t, dc)
	})
	t.Run("should work for meta", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = core.MetachainShardId
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)

		dcf, err := dataComp.NewDataComponentsFactory(args)
		require.NoError(t, err)
		dc, err := dcf.Create()
		require.NoError(t, err)
		require.NotNil(t, dc)
	})
}

func TestManagedDataComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	runTypeComponents := componentsMock.GetRunTypeComponents()
	args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
	dcf, _ := dataComp.NewDataComponentsFactory(args)

	dc, _ := dcf.Create()

	err := dc.Close()
	require.NoError(t, err)
}
