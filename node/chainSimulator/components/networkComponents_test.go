package components

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateNetworkComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateNetworkComponents(NewSyncedBroadcastNetwork())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("nil network should error", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateNetworkComponents(nil)
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func TestNetworkComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *networkComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateNetworkComponents(NewSyncedBroadcastNetwork())
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestNetworkComponentsHolder_Getters(t *testing.T) {
	t.Parallel()

	comp, err := CreateNetworkComponents(NewSyncedBroadcastNetwork())
	require.NoError(t, err)

	require.NotNil(t, comp.NetworkMessenger())
	require.NotNil(t, comp.InputAntiFloodHandler())
	require.NotNil(t, comp.OutputAntiFloodHandler())
	require.NotNil(t, comp.PubKeyCacher())
	require.NotNil(t, comp.PeerBlackListHandler())
	require.NotNil(t, comp.PeerHonestyHandler())
	require.NotNil(t, comp.PreferredPeersHolderHandler())
	require.NotNil(t, comp.PeersRatingHandler())
	require.NotNil(t, comp.PeersRatingMonitor())
	require.NotNil(t, comp.FullArchiveNetworkMessenger())
	require.NotNil(t, comp.FullArchivePreferredPeersHolderHandler())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())
	require.Nil(t, comp.Close())
}
