package data_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewManagedDataComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedDataComponents, err := dataComp.NewManagedDataComponents(nil)
		require.Equal(t, errorsMx.ErrNilDataComponentsFactory, err)
		require.Nil(t, managedDataComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		runTypeComponents := componentsMock.GetRunTypeComponents()
		coreComponents := componentsMock.GetCoreComponents()
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
		managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
		require.Nil(t, err)
		require.NotNil(t, managedDataComponents)
	})
}

func TestManagedDataComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
		dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
		managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
		require.NoError(t, err)
		err = managedDataComponents.Create()
		require.Error(t, err)
		require.Nil(t, managedDataComponents.Blockchain())
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		runTypeComponents := componentsMock.GetRunTypeComponents()
		args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
		dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
		managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
		require.NoError(t, err)
		require.Nil(t, managedDataComponents.Blockchain())
		require.Nil(t, managedDataComponents.StorageService())
		require.Nil(t, managedDataComponents.Datapool())
		require.Nil(t, managedDataComponents.MiniBlocksProvider())

		err = managedDataComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedDataComponents.Blockchain())
		require.NotNil(t, managedDataComponents.StorageService())
		require.NotNil(t, managedDataComponents.Datapool())
		require.NotNil(t, managedDataComponents.MiniBlocksProvider())

		require.Equal(t, factory.DataComponentsName, managedDataComponents.String())
	})
}

func TestManagedDataComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	runTypeComponents := componentsMock.GetRunTypeComponents()
	args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
	require.NoError(t, err)
	require.Equal(t, errorsMx.ErrNilDataComponents, managedDataComponents.CheckSubcomponents())

	err = managedDataComponents.Create()
	require.NoError(t, err)
	require.Nil(t, managedDataComponents.CheckSubcomponents())
}

func TestManagedDataComponents_SetBlockchain(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	runTypeComponents := componentsMock.GetRunTypeComponents()
	args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, _ := dataComp.NewManagedDataComponents(dataComponentsFactory)

	_ = managedDataComponents.Create()

	require.Equal(t, errorsMx.ErrNilBlockChainHandler, managedDataComponents.SetBlockchain(nil))
	require.Nil(t, managedDataComponents.SetBlockchain(&testscommon.ChainHandlerMock{}))
}

func TestManagedDataComponents_Close(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	runTypeComponents := componentsMock.GetRunTypeComponents()
	args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, _ := dataComp.NewManagedDataComponents(dataComponentsFactory)
	err := managedDataComponents.Create()
	require.NoError(t, err)

	err = managedDataComponents.Close()
	require.NoError(t, err)

	err = managedDataComponents.Close()
	require.NoError(t, err)
}

func TestManagedDataComponents_Clone(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	runTypeComponents := componentsMock.GetRunTypeComponents()
	args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, _ := dataComp.NewManagedDataComponents(dataComponentsFactory)

	clonedBeforeCreate := managedDataComponents.Clone()
	require.Equal(t, managedDataComponents, clonedBeforeCreate)

	_ = managedDataComponents.Create()
	clonedAfterCreate := managedDataComponents.Clone()
	require.Equal(t, managedDataComponents, clonedAfterCreate)

	_ = managedDataComponents.Close()
	clonedAfterClose := managedDataComponents.Clone()
	require.Equal(t, managedDataComponents, clonedAfterClose)
}

func TestManagedDataComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedDataComponents, _ := dataComp.NewManagedDataComponents(nil)
	require.True(t, managedDataComponents.IsInterfaceNil())

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	runTypeComponents := componentsMock.GetRunTypeComponents()
	args := componentsMock.GetDataArgs(coreComponents, runTypeComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, _ = dataComp.NewManagedDataComponents(dataComponentsFactory)
	require.False(t, managedDataComponents.IsInterfaceNil())
}
