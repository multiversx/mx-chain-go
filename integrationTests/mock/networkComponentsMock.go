package mock

import (
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// NetworkComponentsStub -
type NetworkComponentsStub struct {
	Messenger       p2p.Messenger
	InputAntiFlood  factory.P2PAntifloodHandler
	OutputAntiFlood factory.P2PAntifloodHandler
	PeerBlackList   process.PeerBlackListCacher
	PeerHonesty     factory.PeerHonestyHandler
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

// IsInterfaceNil -
func (ncs *NetworkComponentsStub) IsInterfaceNil() bool {
	return ncs == nil
}
