package factory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery/factory"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerDiscoverer_NoDiscoveryEnabledShouldRetNullDiscoverer(t *testing.T) {
	t.Parallel()

	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled: false,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(
		context.Background(),
		&mock.ConnectableHostStub{},
		&mock.SharderStub{},
		p2pConfig,
	)
	_, ok := pDiscoverer.(*discovery.NilDiscoverer)

	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestNewPeerDiscoverer_KadInvalidIntervalShouldErr(t *testing.T) {
	t.Parallel()

	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 0,
		},
		Sharding: config.ShardingConfig{
			Type: p2p.SharderVariantPrioBits,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(
		context.Background(),
		&mock.ConnectableHostStub{},
		&mock.SharderStub{},
		p2pConfig,
	)

	assert.Nil(t, pDiscoverer)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewPeerDiscoverer_KadPrioBitsShouldWork(t *testing.T) {
	t.Parallel()

	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
		},
		Sharding: config.ShardingConfig{
			Type: p2p.SharderVariantPrioBits,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(
		context.Background(),
		&mock.ConnectableHostStub{},
		&mock.SharderStub{},
		p2pConfig,
	)
	_, ok := pDiscoverer.(*discovery.KadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestNewPeerDiscoverer_KadListShouldWork(t *testing.T) {
	t.Parallel()

	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
		},
		Sharding: config.ShardingConfig{
			Type: p2p.SharderVariantWithLists,
		},
	}

	pDiscoverer, err := factory.NewPeerDiscoverer(
		context.Background(),
		&mock.ConnectableHostStub{},
		&mock.SharderStub{},
		p2pConfig,
	)
	_, ok := pDiscoverer.(*discovery.ContinuousKadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}
