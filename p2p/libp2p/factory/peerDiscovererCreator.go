package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
)

type peerDiscovererCreator struct {
	p2pConfig config.P2PConfig
}

// NewPeerDiscovererCreator creates a new instance of peer discovery factory
func NewPeerDiscovererCreator(pConfig config.P2PConfig) *peerDiscovererCreator {
	return &peerDiscovererCreator{
		p2pConfig: pConfig,
	}
}

// CreatePeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func (pdc *peerDiscovererCreator) CreatePeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdc.p2pConfig.KadDhtPeerDiscovery.Enabled {
		return pdc.createKadDhtPeerDiscoverer()
	}

	return discovery.NewNullDiscoverer(), nil
}

func (pdc *peerDiscovererCreator) createKadDhtPeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec <= 0 {
		return nil, p2p.ErrNegativeOrZeroPeersRefreshInterval
	}

	return discovery.NewKadDhtPeerDiscoverer(
		time.Second*time.Duration(pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		pdc.p2pConfig.KadDhtPeerDiscovery.RandezVous,
		pdc.p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
	), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pdc *peerDiscovererCreator) IsInterfaceNil() bool {
	if pdc == nil {
		return true
	}
	return false
}
