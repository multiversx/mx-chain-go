package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkComponentsFactory_NilP2PConfigShouldErr(t *testing.T) {
	t.Parallel()

	ncf, err := NewNetworkComponentsFactory(nil, &config.Config{}, &mock.AppStatusHandlerFake{})
	require.Nil(t, ncf)
	require.Equal(t, ErrNilP2PConfiguration, err)
}

func TestNewNetworkComponentsFactory_NilMainConfigShouldErr(t *testing.T) {
	t.Parallel()

	ncf, err := NewNetworkComponentsFactory(&config.P2PConfig{}, nil, &mock.AppStatusHandlerFake{})
	require.Nil(t, ncf)
	require.Equal(t, ErrNilConfiguration, err)
}

func TestNewNetworkComponentsFactory_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	ncf, err := NewNetworkComponentsFactory(&config.P2PConfig{}, &config.Config{}, nil)
	require.Nil(t, ncf)
	require.Equal(t, ErrNilStatusHandler, err)
}

func TestNewNetworkComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	ncf, err := NewNetworkComponentsFactory(&config.P2PConfig{}, &config.Config{}, &mock.AppStatusHandlerFake{})
	require.NoError(t, err)
	require.NotNil(t, ncf)
}

func TestNetworkComponentsFactory_Create_ShouldErrDueToBadConfig(t *testing.T) {
	t.Parallel()

	ncf, _ := NewNetworkComponentsFactory(&config.P2PConfig{}, &config.Config{}, &mock.AppStatusHandlerFake{})

	nc, err := ncf.Create()
	require.Error(t, err)
	require.Nil(t, nc)
}

func TestNetworkComponentsFactory_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	p2pConfig := &config.P2PConfig{
		Node: config.NodeConfig{
			Seed: "seed",
		},
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:                          false,
			RefreshIntervalInSec:             10,
			RandezVous:                       "bien sur",
			InitialPeerList:                  []string{"peer0", "peer1"},
			BucketSize:                       10,
			RoutingTableRefreshIntervalInSec: 5,
		},
		Sharding: config.ShardingConfig{
			TargetPeerCount:         10,
			PrioBits:                10,
			MaxIntraShardValidators: 10,
			MaxCrossShardValidators: 10,
			MaxIntraShardObservers:  10,
			MaxCrossShardObservers:  10,
			Type:                    "NilListSharder",
		},
	}
	ncf, _ := NewNetworkComponentsFactory(p2pConfig, &config.Config{}, &mock.AppStatusHandlerFake{})

	ncf.SetListenAddress(libp2p.ListenLocalhostAddrWithIp4AndTcp)

	nc, err := ncf.Create()
	require.NoError(t, err)
	require.NotNil(t, nc)
}
