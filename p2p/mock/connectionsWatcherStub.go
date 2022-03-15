package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// ConnectionsWatcherStub -
type ConnectionsWatcherStub struct {
	NewKnownConnectionCalled func(pid core.PeerID, connection string)
	CloseCalled              func() error
	PeerConnectedCalled      func(pid core.PeerID)
}

// NewKnownConnection -
func (stub *ConnectionsWatcherStub) NewKnownConnection(pid core.PeerID, connection string) {
	if stub.NewKnownConnectionCalled != nil {
		stub.NewKnownConnectionCalled(pid, connection)
	}
}

// PeerConnected -
func (stub *ConnectionsWatcherStub) PeerConnected(pid core.PeerID) {
	if stub.PeerConnectedCalled != nil {
		stub.PeerConnectedCalled(pid)
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
