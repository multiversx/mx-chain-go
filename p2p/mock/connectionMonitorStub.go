package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type ConnectionMonitorStub struct {
	HandleConnectedPeerCalled      func(pid p2p.PeerID) error
	HandleDisconnectedPeerCalled   func(pid p2p.PeerID) error
	DoReconnectionBlockingCalled   func()
	CheckConnectionsBlockingCalled func()
}

func (cms *ConnectionMonitorStub) HandleConnectedPeer(pid p2p.PeerID) error {
	return cms.HandleConnectedPeerCalled(pid)
}

func (cms *ConnectionMonitorStub) HandleDisconnectedPeer(pid p2p.PeerID) error {
	return cms.HandleDisconnectedPeerCalled(pid)
}

func (cms *ConnectionMonitorStub) DoReconnectionBlocking() {
	cms.DoReconnectionBlockingCalled()
}

func (cms *ConnectionMonitorStub) CheckConnectionsBlocking() {
	cms.CheckConnectionsBlockingCalled()
}

func (cms *ConnectionMonitorStub) IsInterfaceNil() bool {
	return cms == nil
}
