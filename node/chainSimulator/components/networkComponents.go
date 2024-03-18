package components

import (
	disabledBootstrap "github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/factory"
	disabledFactory "github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
	disabledP2P "github.com/multiversx/mx-chain-go/p2p/disabled"
	"github.com/multiversx/mx-chain-go/process"
	disabledAntiflood "github.com/multiversx/mx-chain-go/process/throttle/antiflood/disabled"
)

type networkComponentsHolder struct {
	closeHandler                           *closeHandler
	networkMessenger                       p2p.Messenger
	inputAntiFloodHandler                  factory.P2PAntifloodHandler
	outputAntiFloodHandler                 factory.P2PAntifloodHandler
	pubKeyCacher                           process.TimeCacher
	peerBlackListHandler                   process.PeerBlackListCacher
	peerHonestyHandler                     factory.PeerHonestyHandler
	preferredPeersHolderHandler            factory.PreferredPeersHolderHandler
	peersRatingHandler                     p2p.PeersRatingHandler
	peersRatingMonitor                     p2p.PeersRatingMonitor
	fullArchiveNetworkMessenger            p2p.Messenger
	fullArchivePreferredPeersHolderHandler factory.PreferredPeersHolderHandler
}

// CreateNetworkComponents creates a new networkComponentsHolder instance
func CreateNetworkComponents(network SyncedBroadcastNetworkHandler) (*networkComponentsHolder, error) {
	messenger, err := NewSyncedMessenger(network)
	if err != nil {
		return nil, err
	}

	instance := &networkComponentsHolder{
		closeHandler:                           NewCloseHandler(),
		networkMessenger:                       messenger,
		inputAntiFloodHandler:                  disabled.NewAntiFlooder(),
		outputAntiFloodHandler:                 disabled.NewAntiFlooder(),
		pubKeyCacher:                           &disabledAntiflood.TimeCache{},
		peerBlackListHandler:                   &disabledAntiflood.PeerBlacklistCacher{},
		peerHonestyHandler:                     disabled.NewPeerHonesty(),
		preferredPeersHolderHandler:            disabledFactory.NewPreferredPeersHolder(),
		peersRatingHandler:                     disabledBootstrap.NewDisabledPeersRatingHandler(),
		peersRatingMonitor:                     disabled.NewPeersRatingMonitor(),
		fullArchiveNetworkMessenger:            disabledP2P.NewNetworkMessenger(),
		fullArchivePreferredPeersHolderHandler: disabledFactory.NewPreferredPeersHolder(),
	}

	instance.collectClosableComponents()

	return instance, nil
}

// NetworkMessenger returns the network messenger
func (holder *networkComponentsHolder) NetworkMessenger() p2p.Messenger {
	return holder.networkMessenger
}

// InputAntiFloodHandler returns the input antiflooder
func (holder *networkComponentsHolder) InputAntiFloodHandler() factory.P2PAntifloodHandler {
	return holder.inputAntiFloodHandler
}

// OutputAntiFloodHandler returns the output antiflooder
func (holder *networkComponentsHolder) OutputAntiFloodHandler() factory.P2PAntifloodHandler {
	return holder.outputAntiFloodHandler
}

// PubKeyCacher returns the public key cacher
func (holder *networkComponentsHolder) PubKeyCacher() process.TimeCacher {
	return holder.pubKeyCacher
}

// PeerBlackListHandler returns the peer blacklist handler
func (holder *networkComponentsHolder) PeerBlackListHandler() process.PeerBlackListCacher {
	return holder.peerBlackListHandler
}

// PeerHonestyHandler returns the peer honesty handler
func (holder *networkComponentsHolder) PeerHonestyHandler() factory.PeerHonestyHandler {
	return holder.peerHonestyHandler
}

// PreferredPeersHolderHandler returns the preferred peers holder
func (holder *networkComponentsHolder) PreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	return holder.preferredPeersHolderHandler
}

// PeersRatingHandler returns the peers rating handler
func (holder *networkComponentsHolder) PeersRatingHandler() p2p.PeersRatingHandler {
	return holder.peersRatingHandler
}

// PeersRatingMonitor returns the peers rating monitor
func (holder *networkComponentsHolder) PeersRatingMonitor() p2p.PeersRatingMonitor {
	return holder.peersRatingMonitor
}

// FullArchiveNetworkMessenger returns the full archive network messenger
func (holder *networkComponentsHolder) FullArchiveNetworkMessenger() p2p.Messenger {
	return holder.fullArchiveNetworkMessenger
}

// FullArchivePreferredPeersHolderHandler returns the full archive preferred peers holder
func (holder *networkComponentsHolder) FullArchivePreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	return holder.fullArchivePreferredPeersHolderHandler
}

func (holder *networkComponentsHolder) collectClosableComponents() {
	holder.closeHandler.AddComponent(holder.networkMessenger)
	holder.closeHandler.AddComponent(holder.inputAntiFloodHandler)
	holder.closeHandler.AddComponent(holder.outputAntiFloodHandler)
	holder.closeHandler.AddComponent(holder.peerHonestyHandler)
	holder.closeHandler.AddComponent(holder.fullArchiveNetworkMessenger)
}

// Close will call the Close methods on all inner components
func (holder *networkComponentsHolder) Close() error {
	return holder.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *networkComponentsHolder) IsInterfaceNil() bool {
	return holder == nil
}

// Create will do nothing
func (holder *networkComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (holder *networkComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (holder *networkComponentsHolder) String() string {
	return ""
}
