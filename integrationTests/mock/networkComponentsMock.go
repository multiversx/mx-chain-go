package mock

import (
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// NetworkComponentsStub -
type NetworkComponentsStub struct {
	Messenger                        p2p.Messenger
	MessengerCalled                  func() p2p.Messenger
	InputAntiFlood                   factory.P2PAntifloodHandler
	OutputAntiFlood                  factory.P2PAntifloodHandler
	PeerBlackList                    process.PeerBlackListCacher
	PeerHonesty                      factory.PeerHonestyHandler
	PreferredPeersHolder             factory.PreferredPeersHolderHandler
	PeersRatingHandlerField          p2p.PeersRatingHandler
	PeersRatingMonitorField          p2p.PeersRatingMonitor
	FullArchiveNetworkMessengerField p2p.Messenger
	FullArchivePreferredPeersHolder  factory.PreferredPeersHolderHandler
}

// PubKeyCacher -
func (ncs *NetworkComponentsStub) PubKeyCacher() process.TimeCacher {
	panic("implement me")
}

// PeerHonestyHandler -
func (ncs *NetworkComponentsStub) PeerHonestyHandler() factory.PeerHonestyHandler {
	return ncs.PeerHonesty
}

// Create -
func (ncs *NetworkComponentsStub) Create() error {
	return nil
}

// Close -
func (ncs *NetworkComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (ncs *NetworkComponentsStub) CheckSubcomponents() error {
	return nil
}

// NetworkMessenger -
func (ncs *NetworkComponentsStub) NetworkMessenger() p2p.Messenger {
	if ncs.MessengerCalled != nil {
		return ncs.MessengerCalled()
	}
	return ncs.Messenger
}

// InputAntiFloodHandler -
func (ncs *NetworkComponentsStub) InputAntiFloodHandler() factory.P2PAntifloodHandler {
	return ncs.InputAntiFlood
}

// OutputAntiFloodHandler -
func (ncs *NetworkComponentsStub) OutputAntiFloodHandler() factory.P2PAntifloodHandler {
	return ncs.OutputAntiFlood
}

// PeerBlackListHandler -
func (ncs *NetworkComponentsStub) PeerBlackListHandler() process.PeerBlackListCacher {
	return ncs.PeerBlackList
}

// PreferredPeersHolderHandler -
func (ncs *NetworkComponentsStub) PreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	return ncs.PreferredPeersHolder
}

// PeersRatingHandler -
func (ncs *NetworkComponentsStub) PeersRatingHandler() p2p.PeersRatingHandler {
	return ncs.PeersRatingHandlerField
}

// PeersRatingMonitor -
func (ncs *NetworkComponentsStub) PeersRatingMonitor() p2p.PeersRatingMonitor {
	return ncs.PeersRatingMonitorField
}

// FullArchiveNetworkMessenger -
func (ncs *NetworkComponentsStub) FullArchiveNetworkMessenger() p2p.Messenger {
	return ncs.FullArchiveNetworkMessengerField
}

// FullArchivePreferredPeersHolderHandler -
func (ncs *NetworkComponentsStub) FullArchivePreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	return ncs.FullArchivePreferredPeersHolder
}

// String -
func (ncs *NetworkComponentsStub) String() string {
	return factory.NetworkComponentsName
}

// IsInterfaceNil -
func (ncs *NetworkComponentsStub) IsInterfaceNil() bool {
	return ncs == nil
}
