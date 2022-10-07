package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// PeerEventsHandlerStub -
type PeerEventsHandlerStub struct {
	ConnectedCalled    func(pid core.PeerID, connection string)
	DisconnectedCalled func(pid core.PeerID)
}

// Connected -
func (stub *PeerEventsHandlerStub) Connected(pid core.PeerID, connection string) {
	if stub.ConnectedCalled != nil {
		stub.ConnectedCalled(pid, connection)
	}
}

// Disconnected -
func (stub *PeerEventsHandlerStub) Disconnected(pid core.PeerID) {
	if stub.DisconnectedCalled != nil {
		stub.DisconnectedCalled(pid)
	}
}

// IsInterfaceNil -
func (stub *PeerEventsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
