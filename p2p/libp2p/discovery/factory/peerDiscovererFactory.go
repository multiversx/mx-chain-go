package factory

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
)

// NewPeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func NewPeerDiscoverer(p2pConfig config.P2PConfig) (p2p.PeerDiscoverer, error) {
	if p2pConfig.KadDhtPeerDiscovery.Enabled {
		return createKadDhtPeerDiscoverer(p2pConfig)
	}

	return discovery.NewNullDiscoverer(), nil
}

func createKadDhtPeerDiscoverer(p2pConfig config.P2PConfig) (p2p.PeerDiscoverer, error) {
	arg := discovery.ArgKadDht{
		PeersRefreshInterval: time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		RandezVous:           p2pConfig.KadDhtPeerDiscovery.RandezVous,
		InitialPeersList:     p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:           p2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:  time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
	}

	switch p2pConfig.KadDhtPeerDiscovery.Type {
	case config.KadDhtVariantPrioBits:
		return discovery.NewKadDhtPeerDiscoverer(arg)
	case config.KadDhtVariantWithLists:
		return discovery.NewContinuousKadDhtDiscoverer(arg)
	default:
		return nil, fmt.Errorf("%w unknown kad dht type: %s", p2p.ErrInvalidValue, p2pConfig.KadDhtPeerDiscovery.Type)
	}
}
