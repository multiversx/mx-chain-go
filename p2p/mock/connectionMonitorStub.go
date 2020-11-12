package mock

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// ConnectionMonitorStub -
type ConnectionMonitorStub struct {
	ListenCalled                        func(netw network.Network, ma multiaddr.Multiaddr)
	ListenCloseCalled                   func(netw network.Network, ma multiaddr.Multiaddr)
	ConnectedCalled                     func(netw network.Network, conn network.Conn)
	DisconnectedCalled                  func(netw network.Network, conn network.Conn)
	OpenedStreamCalled                  func(netw network.Network, stream network.Stream)
	ClosedStreamCalled                  func(netw network.Network, stream network.Stream)
	IsConnectedToTheNetworkCalled       func(netw network.Network) bool
	SetThresholdMinConnectedPeersCalled func(thresholdMinConnectedPeers int, netw network.Network)
	ThresholdMinConnectedPeersCalled    func() int
}

// Listen -
func (cms *ConnectionMonitorStub) Listen(netw network.Network, ma multiaddr.Multiaddr) {
	if cms.ListenCalled != nil {
		cms.ListenCalled(netw, ma)
	}
}

// ListenClose -
func (cms *ConnectionMonitorStub) ListenClose(netw network.Network, ma multiaddr.Multiaddr) {
	if cms.ListenCloseCalled != nil {
		cms.ListenCloseCalled(netw, ma)
	}
}

// Connected -
func (cms *ConnectionMonitorStub) Connected(netw network.Network, conn network.Conn) {
	if cms.ConnectedCalled != nil {
		cms.ConnectedCalled(netw, conn)
	}
}

// Disconnected -
func (cms *ConnectionMonitorStub) Disconnected(netw network.Network, conn network.Conn) {
	if cms.DisconnectedCalled != nil {
		cms.DisconnectedCalled(netw, conn)
	}
}

// OpenedStream -
func (cms *ConnectionMonitorStub) OpenedStream(netw network.Network, stream network.Stream) {
	if cms.OpenedStreamCalled != nil {
		cms.OpenedStreamCalled(netw, stream)
	}
}

// ClosedStream -
func (cms *ConnectionMonitorStub) ClosedStream(netw network.Network, stream network.Stream) {
	if cms.ClosedStreamCalled != nil {
		cms.ClosedStreamCalled(netw, stream)
	}
}

// IsConnectedToTheNetwork -
func (cms *ConnectionMonitorStub) IsConnectedToTheNetwork(netw network.Network) bool {
	if cms.IsConnectedToTheNetworkCalled != nil {
		return cms.IsConnectedToTheNetworkCalled(netw)
	}

	return false
}

// SetThresholdMinConnectedPeers -
func (cms *ConnectionMonitorStub) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network) {
	if cms.SetThresholdMinConnectedPeersCalled != nil {
		cms.SetThresholdMinConnectedPeersCalled(thresholdMinConnectedPeers, netw)
	}
}

// ThresholdMinConnectedPeers -
func (cms *ConnectionMonitorStub) ThresholdMinConnectedPeers() int {
	if cms.ThresholdMinConnectedPeersCalled != nil {
		return cms.ThresholdMinConnectedPeersCalled()
	}

	return 0
}

// Close -
func (cms *ConnectionMonitorStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (cms *ConnectionMonitorStub) IsInterfaceNil() bool {
	return cms == nil
}
