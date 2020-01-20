package factory_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery/factory"
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

func TestPeerDiscovererFactory_CreatePeerDiscovererKadIntervalLessThanZeroShouldErr(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: -1,
		},
	}

	f := factory.NewPeerDiscovererFactory(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	assert.Nil(t, pDiscoverer)
	assert.Equal(t, p2p.ErrNegativeOrZeroPeersRefreshInterval, err)
}

func TestPeerDiscovererFactory_CreatePeerDiscovererKadPrioBitsShouldWork(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			Type:                 config.KadDhtVariantPrioBits,
		},
	}

	f := factory.NewPeerDiscovererFactory(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	_, ok := pDiscoverer.(*discovery.KadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPeerDiscovererFactory_CreatePeerDiscovererKadListShouldWork(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			Type:                 config.KadDhtVariantWithLists,
		},
	}

	f := factory.NewPeerDiscovererFactory(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	_, ok := pDiscoverer.(*discovery.ContinousKadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPeerDiscovererFactory_CreatePeerDiscovererKadUnknownShouldErr(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			Type:                 "unknown kad dht implementation",
		},
	}

	f := factory.NewPeerDiscovererFactory(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	assert.Nil(t, pDiscoverer)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}
