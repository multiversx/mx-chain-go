package network_test

import (
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	networkComp "github.com/multiversx/mx-chain-go/factory/network"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewManagedNetworkComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(nil)
		require.Equal(t, errorsMx.ErrNilNetworkComponentsFactory, err)
		require.Nil(t, managedNetworkComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(componentsMock.GetNetworkFactoryArgs())
		managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedNetworkComponents)
	})
}

func TestManagedNetworkComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		networkArgs := componentsMock.GetNetworkFactoryArgs()
		networkArgs.MainP2pConfig.Node.Port = "invalid"
		networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
		managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
		require.NoError(t, err)
		err = managedNetworkComponents.Create()
		require.Error(t, err)
		require.Nil(t, managedNetworkComponents.NetworkMessenger())
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		networkArgs := componentsMock.GetNetworkFactoryArgs()
		networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
		managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedNetworkComponents)
		require.Nil(t, managedNetworkComponents.NetworkMessenger())
		require.Nil(t, managedNetworkComponents.InputAntiFloodHandler())
		require.Nil(t, managedNetworkComponents.OutputAntiFloodHandler())
		require.Nil(t, managedNetworkComponents.PeerBlackListHandler())
		require.Nil(t, managedNetworkComponents.PubKeyCacher())
		require.Nil(t, managedNetworkComponents.PreferredPeersHolderHandler())
		require.Nil(t, managedNetworkComponents.PeerHonestyHandler())
		require.Nil(t, managedNetworkComponents.PeersRatingHandler())
		require.Nil(t, managedNetworkComponents.FullArchiveNetworkMessenger())
		require.Nil(t, managedNetworkComponents.FullArchivePeersRatingHandler())
		require.Nil(t, managedNetworkComponents.FullArchivePeersRatingMonitor())
		require.Nil(t, managedNetworkComponents.FullArchivePreferredPeersHolderHandler())

		err = managedNetworkComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedNetworkComponents.NetworkMessenger())
		require.NotNil(t, managedNetworkComponents.InputAntiFloodHandler())
		require.NotNil(t, managedNetworkComponents.OutputAntiFloodHandler())
		require.NotNil(t, managedNetworkComponents.PeerBlackListHandler())
		require.NotNil(t, managedNetworkComponents.PubKeyCacher())
		require.NotNil(t, managedNetworkComponents.PreferredPeersHolderHandler())
		require.NotNil(t, managedNetworkComponents.PeerHonestyHandler())
		require.NotNil(t, managedNetworkComponents.PeersRatingHandler())
		require.NotNil(t, managedNetworkComponents.FullArchiveNetworkMessenger())
		require.NotNil(t, managedNetworkComponents.FullArchivePeersRatingHandler())
		require.NotNil(t, managedNetworkComponents.FullArchivePeersRatingMonitor())
		require.NotNil(t, managedNetworkComponents.FullArchivePreferredPeersHolderHandler())

		require.Equal(t, factory.NetworkComponentsName, managedNetworkComponents.String())
	})
}

func TestManagedNetworkComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	networkArgs := componentsMock.GetNetworkFactoryArgs()
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	require.NoError(t, err)

	require.Error(t, managedNetworkComponents.CheckSubcomponents())

	_ = managedNetworkComponents.Create()

	require.NoError(t, managedNetworkComponents.CheckSubcomponents())
}

func TestManagedNetworkComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	networkArgs := componentsMock.GetNetworkFactoryArgs()
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	managedNetworkComponents, _ := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	err := managedNetworkComponents.Close()
	require.NoError(t, err)

	err = managedNetworkComponents.Create()
	require.NoError(t, err)

	err = managedNetworkComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedNetworkComponents.NetworkMessenger())
}

func TestManagedNetworkComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedNetworkComponents, _ := networkComp.NewManagedNetworkComponents(nil)
	require.True(t, managedNetworkComponents.IsInterfaceNil())

	networkArgs := componentsMock.GetNetworkFactoryArgs()
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	managedNetworkComponents, _ = networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	require.False(t, managedNetworkComponents.IsInterfaceNil())
}
