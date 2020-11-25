package mock

import (
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// NetworkComponentsMock -
type NetworkComponentsMock struct {
	Messenger       p2p.Messenger
	InputAntiFlood  factory.P2PAntifloodHandler
	OutputAntiFlood factory.P2PAntifloodHandler
	PeerBlackList   process.PeerBlackListCacher
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

// IsInterfaceNil -
func (ncm *NetworkComponentsMock) IsInterfaceNil() bool {
	return ncm == nil
}
