package factory

import (
	"context"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
)

const typeLegacy = "legacy"
const typeOptimized = "optimized"
const defaultSeedersReconnectionInterval = time.Minute * 5

var log = logger.GetOrCreate("p2p/discovery/factory")

// NewPeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func NewPeerDiscoverer(
	context context.Context,
	host discovery.ConnectableHost,
	sharder p2p.Sharder,
	p2pConfig config.P2PConfig,
) (p2p.PeerDiscoverer, error) {
	if p2pConfig.KadDhtPeerDiscovery.Enabled {
		return createKadDhtPeerDiscoverer(context, host, sharder, p2pConfig)
	}

	log.Debug("using nil discoverer")
	return discovery.NewNilDiscoverer(), nil
}

func createKadDhtPeerDiscoverer(
	context context.Context,
	host discovery.ConnectableHost,
	sharder p2p.Sharder,
	p2pConfig config.P2PConfig,
) (p2p.PeerDiscoverer, error) {
	arg := discovery.ArgKadDht{
		Context:                     context,
		Host:                        host,
		KddSharder:                  sharder,
		PeersRefreshInterval:        time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		SeedersReconnectionInterval: defaultSeedersReconnectionInterval,
		ProtocolID:                  p2pConfig.KadDhtPeerDiscovery.ProtocolID,
		InitialPeersList:            p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:                  p2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:         time.Second * time.Duration(p2pConfig.KadDhtPeerDiscovery.RoutingTableRefreshIntervalInSec),
	}

	switch p2pConfig.Sharding.Type {
	case p2p.ListsSharder, p2p.OneListSharder, p2p.NilListSharder:
		return createKadDhtDiscoverer(p2pConfig, arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on "+
			"selected sharder: unknown sharder '%s'", p2p.ErrInvalidValue, p2pConfig.Sharding.Type)
	}
}

func createKadDhtDiscoverer(p2pConfig config.P2PConfig, arg discovery.ArgKadDht) (p2p.PeerDiscoverer, error) {
	switch p2pConfig.KadDhtPeerDiscovery.Type {
	case typeLegacy:
		log.Debug("using continuous (legacy) kad dht discoverer")
		return discovery.NewContinuousKadDhtDiscoverer(arg)
	case typeOptimized:
		log.Debug("using optimized kad dht discoverer")
		return discovery.NewOptimizedKadDhtDiscoverer(arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on type '%s'",
			p2p.ErrInvalidValue, p2pConfig.KadDhtPeerDiscovery.Type)
	}
}
