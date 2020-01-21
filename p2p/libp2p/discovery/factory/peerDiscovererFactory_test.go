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

func TestNewPeerDiscoverer_NoDiscoveryEnabledShouldRetNullDiscoverer(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled: false,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(p2pConfig)
	_, ok := pDiscoverer.(*discovery.NullDiscoverer)

	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestNewPeerDiscoverer_KadInvalidIntervalShouldErr(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			Type:                 config.KadDhtVariantPrioBits,
			RefreshIntervalInSec: 0,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(p2pConfig)

	assert.Nil(t, pDiscoverer)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewPeerDiscoverer_KadPrioBitsShouldWork(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			Type:                 config.KadDhtVariantPrioBits,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(p2pConfig)
	_, ok := pDiscoverer.(*discovery.KadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestNewPeerDiscoverer_KadListShouldWork(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			Type:                 config.KadDhtVariantWithLists,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(p2pConfig)
	_, ok := pDiscoverer.(*discovery.ContinuousKadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestNewPeerDiscoverer_KadUnknownShouldErr(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			Type:                 "unknown kad dht implementation",
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(p2pConfig)

	assert.Nil(t, pDiscoverer)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}
