package state_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestNewManagedStateComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedStateComponents, err := stateComp.NewManagedStateComponents(nil)
		require.Equal(t, errorsMx.ErrNilStateComponentsFactory, err)
		require.Nil(t, managedStateComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)
		stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
		managedStateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedStateComponents)
	})
}

func TestManagedStateComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)
		stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
		managedStateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
		require.NoError(t, err)
		_ = args.Core.SetInternalMarshalizer(nil)
		err = managedStateComponents.Create()
		require.Error(t, err)
		require.Nil(t, managedStateComponents.AccountsAdapter())
		require.NoError(t, managedStateComponents.Close())
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)
		stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
		managedStateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
		require.NoError(t, err)
		require.Nil(t, managedStateComponents.AccountsAdapter())
		require.Nil(t, managedStateComponents.PeerAccounts())
		require.Nil(t, managedStateComponents.TriesContainer())
		require.Nil(t, managedStateComponents.TrieStorageManagers())
		require.Nil(t, managedStateComponents.AccountsAdapterAPI())
		require.Nil(t, managedStateComponents.AccountsRepository())
		require.Nil(t, managedStateComponents.MissingTrieNodesNotifier())

		err = managedStateComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedStateComponents.AccountsAdapter())
		require.NotNil(t, managedStateComponents.PeerAccounts())
		require.NotNil(t, managedStateComponents.TriesContainer())
		require.NotNil(t, managedStateComponents.TrieStorageManagers())
		require.NotNil(t, managedStateComponents.AccountsAdapterAPI())
		require.NotNil(t, managedStateComponents.AccountsRepository())
		require.NotNil(t, managedStateComponents.MissingTrieNodesNotifier())

		require.Equal(t, factory.StateComponentsName, managedStateComponents.String())
		require.NoError(t, managedStateComponents.Close())
	})
}

func TestManagedStateComponents_Close(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetStateFactoryArgs(coreComponents)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ := stateComp.NewManagedStateComponents(stateComponentsFactory)
	require.NoError(t, managedStateComponents.Close())
	err := managedStateComponents.Create()
	require.NoError(t, err)

	require.NoError(t, managedStateComponents.Close())
	require.Nil(t, managedStateComponents.AccountsAdapter())
}

func TestManagedStateComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetStateFactoryArgs(coreComponents)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ := stateComp.NewManagedStateComponents(stateComponentsFactory)
	err := managedStateComponents.CheckSubcomponents()
	require.Equal(t, errorsMx.ErrNilStateComponents, err)

	err = managedStateComponents.Create()
	require.NoError(t, err)

	err = managedStateComponents.CheckSubcomponents()
	require.NoError(t, err)

	require.NoError(t, managedStateComponents.Close())
}

func TestManagedStateComponents_Setters(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetStateFactoryArgs(coreComponents)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ := stateComp.NewManagedStateComponents(stateComponentsFactory)
	err := managedStateComponents.Create()
	require.NoError(t, err)

	triesContainer := &trieMock.TriesHolderStub{}
	triesStorageManagers := map[string]common.StorageManager{"a": &storageManager.StorageManagerStub{}}

	err = managedStateComponents.SetTriesContainer(nil)
	require.Equal(t, errorsMx.ErrNilTriesContainer, err)
	err = managedStateComponents.SetTriesContainer(triesContainer)
	require.NoError(t, err)

	err = managedStateComponents.SetTriesStorageManagers(nil)
	require.Equal(t, errorsMx.ErrNilTriesStorageManagers, err)
	err = managedStateComponents.SetTriesStorageManagers(triesStorageManagers)
	require.NoError(t, err)

	require.Equal(t, triesContainer, managedStateComponents.TriesContainer())
	require.Equal(t, triesStorageManagers, managedStateComponents.TrieStorageManagers())

	require.NoError(t, managedStateComponents.Close())
}

func TestManagedStateComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedStateComponents, _ := stateComp.NewManagedStateComponents(nil)
	require.True(t, managedStateComponents.IsInterfaceNil())

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetStateFactoryArgs(coreComponents)
	stateComponentsFactory, _ := stateComp.NewStateComponentsFactory(args)
	managedStateComponents, _ = stateComp.NewManagedStateComponents(stateComponentsFactory)
	require.False(t, managedStateComponents.IsInterfaceNil())
}
