package factory

import (
	"context"
	"fmt"
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

	return discovery.NewNilDiscoverer(), nil
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
		ProtocolID:           p2pConfig.KadDhtPeerDiscovery.ProtocolID,
		InitialPeersList:     p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:           p2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:  time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RoutingTableRefreshIntervalInSec),
	}

	switch p2pConfig.Sharding.Type {
	case p2p.ListsSharder, p2p.OneListSharder, p2p.NilListSharder:
		return discovery.NewContinuousKadDhtDiscoverer(arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on "+
			"selected sharder: unknown sharder '%s'", p2p.ErrInvalidValue, p2pConfig.Sharding.Type)
	}
}
