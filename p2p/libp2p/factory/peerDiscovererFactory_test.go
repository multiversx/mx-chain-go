package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/factory"
	"github.com/stretchr/testify/assert"
)

func TestPeerDiscovererFactory_CreatePeerDiscovererNoDiscoveryEnabledShouldRetNullDiscoverer(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled: false,
		},
	}

	f := factory.NewPeerDiscovererFactory(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	_, ok := pDiscoverer.(*discovery.NullDiscoverer)

	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPeerDiscovererCreator_CreatePeerDiscovererKadOkValsShouldWork(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:                          true,
			RefreshIntervalInSec:             1,
			RoutingTableRefreshIntervalInSec: 1,
			BucketSize:                       1,
		},
	}

	f := factory.NewPeerDiscovererFactory(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	_, ok := pDiscoverer.(*discovery.KadDhtDiscoverer)

	assert.False(t, check.IfNil(pDiscoverer))
	assert.True(t, ok)
	assert.Nil(t, err)
}
