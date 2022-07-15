package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	stateComp "github.com/ElrondNetwork/elrond-go/factory/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedStateComponents --------------------
func TestManagedStateComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateFactoryArgs(coreComponents, shardCoordinator)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
	require.NoError(t, err)
	_ = args.Core.SetInternalMarshalizer(nil)
	err = managedStateComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedStateComponents.AccountsAdapter())
}

func TestManagedStateComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateFactoryArgs(coreComponents, shardCoordinator)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedStateComponents.AccountsAdapter())
	require.Nil(t, managedStateComponents.PeerAccounts())
	require.Nil(t, managedStateComponents.TriesContainer())
	require.Nil(t, managedStateComponents.TrieStorageManagers())

	err = managedStateComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedStateComponents.AccountsAdapter())
	require.NotNil(t, managedStateComponents.PeerAccounts())
	require.NotNil(t, managedStateComponents.TriesContainer())
	require.NotNil(t, managedStateComponents.TrieStorageManagers())
}

func TestManagedStateComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateFactoryArgs(coreComponents, shardCoordinator)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ := stateComp.NewManagedStateComponents(stateComponentsFactory)
	err := managedStateComponents.Create()
	require.NoError(t, err)

	err = managedStateComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedStateComponents.AccountsAdapter())
}

func TestManagedStateComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateFactoryArgs(coreComponents, shardCoordinator)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ := stateComp.NewManagedStateComponents(stateComponentsFactory)
	err := managedStateComponents.Create()
	require.NoError(t, err)

	err = managedStateComponents.CheckSubcomponents()
	require.NoError(t, err)
}

func TestManagedStateComponents_Setters(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateFactoryArgs(coreComponents, shardCoordinator)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ := stateComp.NewManagedStateComponents(stateComponentsFactory)
	err := managedStateComponents.Create()
	require.NoError(t, err)

	triesContainer := &mock.TriesHolderStub{}
	triesStorageManagers := map[string]common.StorageManager{"a": &testscommon.StorageManagerStub{}}

	err = managedStateComponents.SetTriesContainer(triesContainer)
	require.NoError(t, err)

	err = managedStateComponents.SetTriesStorageManagers(triesStorageManagers)
	require.NoError(t, err)

	require.Equal(t, triesContainer, managedStateComponents.TriesContainer())
	require.Equal(t, triesStorageManagers, managedStateComponents.TrieStorageManagers())
}
