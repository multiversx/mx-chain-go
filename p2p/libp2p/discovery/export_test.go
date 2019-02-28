package discovery

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
)

func (mpd *MdnsPeerDiscoverer) RefreshInterval() time.Duration {
	return mpd.refreshInterval
}

func (mpd *MdnsPeerDiscoverer) ServiceTag() string {
	return mpd.serviceTag
}

func (mpd *MdnsPeerDiscoverer) ContextProvider() *libp2p.Libp2pContext {
	return mpd.contextProvider
}

func (kdd *KadDhtDiscoverer) RefreshInterval() time.Duration {
	return kdd.refreshInterval
}

func (kdd *KadDhtDiscoverer) InitialPeersList() []string {
	return kdd.initialPeersList
}

func (kdd *KadDhtDiscoverer) RandezVous() string {
	return kdd.randezVous
}

func (kdd *KadDhtDiscoverer) ContextProvider() *libp2p.Libp2pContext {
	return kdd.contextProvider
}

func (kdd *KadDhtDiscoverer) ConnectToOnePeerFromInitialPeersList(
	durationBetweenAttempts time.Duration,
	initialPeersList []string) <-chan struct{} {

	return kdd.connectToOnePeerFromInitialPeersList(durationBetweenAttempts, initialPeersList)
}
