package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// ConnectionsNotifieeStub -
type ConnectionsNotifieeStub struct {
	PeerConnectedCalled func(pid core.PeerID)
}

// PeerConnected -
func (stub *ConnectionsNotifieeStub) PeerConnected(pid core.PeerID) {
	if stub.PeerConnectedCalled != nil {
		stub.PeerConnectedCalled(pid)
	}
}

// IsInterfaceNil -
func (stub *ConnectionsNotifieeStub) IsInterfaceNil() bool {
	return stub == nil
}
