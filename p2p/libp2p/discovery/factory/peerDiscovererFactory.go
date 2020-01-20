package factory

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
)

type peerDiscovererFactory struct {
	p2pConfig config.P2PConfig
}

// NewPeerDiscovererFactory creates a new instance of peer discovery factory
//TODO(iulian) refactor this: create a factory function, remove the struct, use a specialized argument and refactor the switch
func NewPeerDiscovererFactory(pConfig config.P2PConfig) *peerDiscovererFactory {
	return &peerDiscovererFactory{
		p2pConfig: pConfig,
	}
}

// CreatePeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func (pdc *peerDiscovererFactory) CreatePeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdc.p2pConfig.KadDhtPeerDiscovery.Enabled {
		return pdc.createKadDhtPeerDiscoverer()
	}

	return discovery.NewNullDiscoverer(), nil
}

func (pdc *peerDiscovererFactory) createKadDhtPeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec <= 0 {
		return nil, p2p.ErrNegativeOrZeroPeersRefreshInterval
	}

	switch pdc.p2pConfig.KadDhtPeerDiscovery.Type {
	case config.KadDhtVariantPrioBits:
		return discovery.NewKadDhtPeerDiscoverer(
			time.Second*time.Duration(pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
			pdc.p2pConfig.KadDhtPeerDiscovery.RandezVous,
			pdc.p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		), nil
	case config.KadDhtVariantWithLists:
		return discovery.NewContinousKadDhtDiscoverer(
			time.Second*time.Duration(pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
			pdc.p2pConfig.KadDhtPeerDiscovery.RandezVous,
			pdc.p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		), nil
	default:
		return nil, fmt.Errorf("%w unknown kad dht type: %s", p2p.ErrInvalidValue, pdc.p2pConfig.KadDhtPeerDiscovery.Type)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pdc *peerDiscovererFactory) IsInterfaceNil() bool {
	return pdc == nil
}
