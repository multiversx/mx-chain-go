package network_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	networkComp "github.com/ElrondNetwork/elrond-go/factory/network"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedNetworkComponents --------------------
func TestManagedNetworkComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	networkArgs := componentsMock.GetNetworkArgs()
	networkArgs.P2pConfig.Node.Port = "invalid"
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	require.NoError(t, err)
	err = managedNetworkComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedNetworkComponents.NetworkMessenger())
}

func TestManagedNetworkComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	networkArgs := componentsMock.GetNetworkArgs()
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	managedNetworkComponents, err := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	require.NoError(t, err)
	require.False(t, check.IfNil(managedNetworkComponents))
	require.Nil(t, managedNetworkComponents.NetworkMessenger())
	require.Nil(t, managedNetworkComponents.InputAntiFloodHandler())
	require.Nil(t, managedNetworkComponents.OutputAntiFloodHandler())
	require.Nil(t, managedNetworkComponents.PeerBlackListHandler())
	require.Nil(t, managedNetworkComponents.PubKeyCacher())
	require.Nil(t, managedNetworkComponents.PreferredPeersHolderHandler())
	require.Nil(t, managedNetworkComponents.PeerHonestyHandler())

	err = managedNetworkComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedNetworkComponents.NetworkMessenger())
	require.NotNil(t, managedNetworkComponents.InputAntiFloodHandler())
	require.NotNil(t, managedNetworkComponents.OutputAntiFloodHandler())
	require.NotNil(t, managedNetworkComponents.PeerBlackListHandler())
	require.NotNil(t, managedNetworkComponents.PubKeyCacher())
	require.NotNil(t, managedNetworkComponents.PreferredPeersHolderHandler())
	require.NotNil(t, managedNetworkComponents.PeerHonestyHandler())
}

func TestManagedNetworkComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	networkArgs := componentsMock.GetNetworkArgs()
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

	networkArgs := componentsMock.GetNetworkArgs()
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	managedNetworkComponents, _ := networkComp.NewManagedNetworkComponents(networkComponentsFactory)
	err := managedNetworkComponents.Create()
	require.NoError(t, err)

	err = managedNetworkComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedNetworkComponents.NetworkMessenger())
}
