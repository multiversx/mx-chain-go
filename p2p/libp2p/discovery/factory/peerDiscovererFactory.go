package factory

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
)

// NewPeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func NewPeerDiscoverer(
	context context.Context,
	host discovery.ConnectableHost,
	sharder p2p.CommonSharder,
	p2pConfig config.P2PConfig,
) (p2p.PeerDiscoverer, error) {
	if p2pConfig.KadDhtPeerDiscovery.Enabled {
		return createKadDhtPeerDiscoverer(context, host, sharder, p2pConfig)
	}

	return discovery.NewNullDiscoverer(), nil
}

func createKadDhtPeerDiscoverer(
	context context.Context,
	host discovery.ConnectableHost,
	sharder p2p.CommonSharder,
	p2pConfig config.P2PConfig,
) (p2p.PeerDiscoverer, error) {
	arg := discovery.ArgKadDht{
		Context:              context,
		Host:                 host,
		KddSharder:           sharder,
		PeersRefreshInterval: time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		RandezVous:           p2pConfig.KadDhtPeerDiscovery.RandezVous,
		InitialPeersList:     p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:           p2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:  time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
	}

	switch p2pConfig.Sharding.Type {
	case p2p.SharderVariantPrioBits:
		return discovery.NewKadDhtPeerDiscoverer(arg)
	default:
		return discovery.NewContinuousKadDhtDiscoverer(arg)
	}
}
