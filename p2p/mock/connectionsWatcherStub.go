package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// ConnectionsWatcherStub -
type ConnectionsWatcherStub struct {
	ConnectedCalled    func(pid core.PeerID, connection string)
	DisconnectedCalled func(pid core.PeerID)
	CloseCalled        func() error
}

// Connected -
func (stub *ConnectionsWatcherStub) Connected(pid core.PeerID, connection string) {
	if stub.ConnectedCalled != nil {
		stub.ConnectedCalled(pid, connection)
	}
}

// Disconnected -
func (stub *ConnectionsWatcherStub) Disconnected(pid core.PeerID) {
	if stub.DisconnectedCalled != nil {
		stub.DisconnectedCalled(pid)
	}
}

// Close -
func (stub *ConnectionsWatcherStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *ConnectionsWatcherStub) IsInterfaceNil() bool {
	return stub == nil
}
