package factory

import (
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// NetworkComponentsMock -
type NetworkComponentsMock struct {
	Messenger               p2p.Messenger
	InputAntiFlood          factory.P2PAntifloodHandler
	OutputAntiFlood         factory.P2PAntifloodHandler
	PeerBlackList           process.PeerBlackListCacher
	PreferredPeersHolder    factory.PreferredPeersHolderHandler
	PeersRatingHandlerField p2p.PeersRatingHandler
	PeersRatingMonitorField p2p.PeersRatingMonitor
}

// PubKeyCacher -
func (ncm *NetworkComponentsMock) PubKeyCacher() process.TimeCacher {
	panic("implement me")
}

// PeerHonestyHandler -
func (ncm *NetworkComponentsMock) PeerHonestyHandler() factory.PeerHonestyHandler {
	panic("implement me")
}

// Create -
func (ncm *NetworkComponentsMock) Create() error {
	return nil
}

// Close -
func (ncm *NetworkComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (ncm *NetworkComponentsMock) CheckSubcomponents() error {
	return nil
}

// NetworkMessenger -
func (ncm *NetworkComponentsMock) NetworkMessenger() p2p.Messenger {
	return ncm.Messenger
}

// InputAntiFloodHandler -
func (ncm *NetworkComponentsMock) InputAntiFloodHandler() factory.P2PAntifloodHandler {
	return ncm.InputAntiFlood
}

// OutputAntiFloodHandler -
func (ncm *NetworkComponentsMock) OutputAntiFloodHandler() factory.P2PAntifloodHandler {
	return ncm.OutputAntiFlood
}

// PeerBlackListHandler -
func (ncm *NetworkComponentsMock) PeerBlackListHandler() process.PeerBlackListCacher {
	return ncm.PeerBlackList
}

// PreferredPeersHolderHandler -
func (ncm *NetworkComponentsMock) PreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	return ncm.PreferredPeersHolder
}

// PeersRatingHandler -
func (ncm *NetworkComponentsMock) PeersRatingHandler() p2p.PeersRatingHandler {
	return ncm.PeersRatingHandlerField
}

// PeersRatingMonitor -
func (ncm *NetworkComponentsMock) PeersRatingMonitor() p2p.PeersRatingMonitor {
	return ncm.PeersRatingMonitorField
}

// String -
func (ncm *NetworkComponentsMock) String() string {
	return "NetworkComponentsMock"
}

// IsInterfaceNil -
func (ncm *NetworkComponentsMock) IsInterfaceNil() bool {
	return ncm == nil
}
