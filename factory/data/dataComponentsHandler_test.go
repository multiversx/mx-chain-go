package data_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	dataComp "github.com/ElrondNetwork/elrond-go/factory/data"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedDataComponents --------------------
func TestManagedDataComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
	require.NoError(t, err)
	err = managedDataComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedDataComponents.Blockchain())
}

func TestManagedDataComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, err := dataComp.NewManagedDataComponents(dataComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedDataComponents.Blockchain())
	require.Nil(t, managedDataComponents.StorageService())
	require.Nil(t, managedDataComponents.Datapool())

	err = managedDataComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedDataComponents.Blockchain())
	require.NotNil(t, managedDataComponents.StorageService())
	require.NotNil(t, managedDataComponents.Datapool())
}

func TestManagedDataComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(args)
	managedDataComponents, _ := dataComp.NewManagedDataComponents(dataComponentsFactory)
	err := managedDataComponents.Create()
	require.NoError(t, err)

	err = managedDataComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedDataComponents.Blockchain())
}

func TestManagedDataComponents_Clone(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
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
