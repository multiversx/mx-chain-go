package network_test

import (
	"errors"
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	networkComp "github.com/multiversx/mx-chain-go/factory/network"
	"github.com/multiversx/mx-chain-go/p2p"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil StatusHandler should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.StatusHandler = nil
		ncf, err := networkComp.NewNetworkComponentsFactory(args)
		require.Nil(t, ncf)
		require.Equal(t, errorsMx.ErrNilStatusHandler, err)
	})
	t.Run("nil Marshalizer should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.Marshalizer = nil
		ncf, err := networkComp.NewNetworkComponentsFactory(args)
		require.Nil(t, ncf)
		require.True(t, errors.Is(err, errorsMx.ErrNilMarshalizer))
	})
	t.Run("nil Syncer should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.Syncer = nil
		ncf, err := networkComp.NewNetworkComponentsFactory(args)
		require.Nil(t, ncf)
		require.Equal(t, errorsMx.ErrNilSyncTimer, err)
	})
	t.Run("nil CryptoComponents should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.CryptoComponents = nil
		ncf, err := networkComp.NewNetworkComponentsFactory(args)
		require.Nil(t, ncf)
		require.Equal(t, errorsMx.ErrNilCryptoComponentsHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		ncf, err := networkComp.NewNetworkComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, ncf)
	})
}

func TestNetworkComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("NewPeersHolder fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.PreferredPeersSlices = []string{"invalid peer"}

		ncf, _ := networkComp.NewNetworkComponentsFactory(args)

		nc, err := ncf.Create()
		require.Error(t, err)
		require.Nil(t, nc)
	})
	t.Run("first NewLRUCache fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.MainConfig.PeersRatingConfig.BadRatedCacheCapacity = 0

		ncf, _ := networkComp.NewNetworkComponentsFactory(args)

		nc, err := ncf.Create()
		require.Error(t, err)
		require.Nil(t, nc)
	})
	t.Run("second NewLRUCache fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.MainConfig.PeersRatingConfig.TopRatedCacheCapacity = 0

		ncf, _ := networkComp.NewNetworkComponentsFactory(args)

		nc, err := ncf.Create()
		require.Error(t, err)
		require.Nil(t, nc)
	})
	t.Run("NewP2PAntiFloodComponents fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.MainConfig.Antiflood.Enabled = true
		args.MainConfig.Antiflood.SlowReacting.BlackList.NumFloodingRounds = 0 // NewP2PAntiFloodComponents fails

		ncf, _ := networkComp.NewNetworkComponentsFactory(args)

		nc, err := ncf.Create()
		require.Error(t, err)
		require.Nil(t, nc)
	})
	t.Run("NewAntifloodDebugger fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.MainConfig.Antiflood.Enabled = true
		args.MainConfig.Debug.Antiflood.CacheSize = 0 // NewAntifloodDebugger fails

		ncf, _ := networkComp.NewNetworkComponentsFactory(args)

		nc, err := ncf.Create()
		require.Error(t, err)
		require.Nil(t, nc)
	})
	t.Run("createPeerHonestyHandler fails should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		args.MainConfig.PeerHonesty.Type = "invalid" // createPeerHonestyHandler fails

		ncf, _ := networkComp.NewNetworkComponentsFactory(args)

		nc, err := ncf.Create()
		require.Error(t, err)
		require.Nil(t, nc)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetNetworkFactoryArgs()
		ncf, _ := networkComp.NewNetworkComponentsFactory(args)
		ncf.SetListenAddress(p2p.ListenLocalhostAddrWithIp4AndTcp)

		nc, err := ncf.Create()
		require.NoError(t, err)
		require.NotNil(t, nc)
		require.NoError(t, nc.Close())
	})
}

func TestNetworkComponents_Close(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, _ := networkComp.NewNetworkComponentsFactory(args)

	nc, _ := ncf.Create()

	err := nc.Close()
	require.NoError(t, err)
}
