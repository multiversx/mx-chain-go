package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
)

//TODO change this to method factory pattern, something like
// NewPeerDiscoverer(pConfig config.P2PConfig) (p2p.PeerDiscoverer, error)
type peerDiscovererFactory struct {
	p2pConfig config.P2PConfig
}

// NewPeerDiscovererFactory creates a new instance of peer discovery factory
func NewPeerDiscovererFactory(pConfig config.P2PConfig) *peerDiscovererFactory {
	return &peerDiscovererFactory{
		p2pConfig: pConfig,
	}
}

// CreatePeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func (pdf *peerDiscovererFactory) CreatePeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdf.p2pConfig.KadDhtPeerDiscovery.Enabled {
		return pdf.createKadDhtPeerDiscoverer()
	}

	return discovery.NewNullDiscoverer(), nil
}

func (pdf *peerDiscovererFactory) createKadDhtPeerDiscoverer() (p2p.PeerDiscoverer, error) {
	arg := discovery.ArgKadDht{
		PeersRefreshInterval: time.Second * time.Duration(pdf.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		RandezVous:           pdf.p2pConfig.KadDhtPeerDiscovery.RandezVous,
		InitialPeersList:     pdf.p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:           pdf.p2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:  time.Second * time.Duration(pdf.p2pConfig.KadDhtPeerDiscovery.RoutingTableRefreshIntervalInSec),
	}

	return discovery.NewKadDhtPeerDiscoverer(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pdf *peerDiscovererFactory) IsInterfaceNil() bool {
	return pdf == nil
}
