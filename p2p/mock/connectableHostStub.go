package mock

import (
	"context"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
)

// ConnectableHostStub -
type ConnectableHostStub struct {
	EventBusCalled              func() event.Bus
	IDCalled                    func() peer.ID
	PeerstoreCalled             func() peerstore.Peerstore
	AddrsCalled                 func() []multiaddr.Multiaddr
	NetworkCalled               func() network.Network
	MuxCalled                   func() protocol.Switch
	ConnectCalled               func(ctx context.Context, pi peer.AddrInfo) error
	SetStreamHandlerCalled      func(pid protocol.ID, handler network.StreamHandler)
	SetStreamHandlerMatchCalled func(protocol.ID, func(string) bool, network.StreamHandler)
	RemoveStreamHandlerCalled   func(pid protocol.ID)
	NewStreamCalled             func(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	CloseCalled                 func() error
	ConnManagerCalled           func() connmgr.ConnManager
	ConnectToPeerCalled         func(ctx context.Context, address string) error
}

// EventBus -
func (hs *ConnectableHostStub) EventBus() event.Bus {
	return hs.EventBusCalled()
}

// ConnectToPeer -
func (hs *ConnectableHostStub) ConnectToPeer(ctx context.Context, address string) error {
	return hs.ConnectToPeerCalled(ctx, address)
}

// ID -
func (hs *ConnectableHostStub) ID() peer.ID {
	return hs.IDCalled()
}

// Peerstore -
func (hs *ConnectableHostStub) Peerstore() peerstore.Peerstore {
	return hs.PeerstoreCalled()
}

// Addrs -
func (hs *ConnectableHostStub) Addrs() []multiaddr.Multiaddr {
	return hs.AddrsCalled()
}

// Network -
func (hs *ConnectableHostStub) Network() network.Network {
	return hs.NetworkCalled()
}

// Mux -
func (hs *ConnectableHostStub) Mux() protocol.Switch {
	return hs.MuxCalled()
}

// Connect -
func (hs *ConnectableHostStub) Connect(ctx context.Context, pi peer.AddrInfo) error {
	return hs.ConnectCalled(ctx, pi)
}

// SetStreamHandler -
func (hs *ConnectableHostStub) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	hs.SetStreamHandlerCalled(pid, handler)
}

// SetStreamHandlerMatch -
func (hs *ConnectableHostStub) SetStreamHandlerMatch(pid protocol.ID, handler func(string) bool, streamHandler network.StreamHandler) {
	hs.SetStreamHandlerMatchCalled(pid, handler, streamHandler)
}

// RemoveStreamHandler -
func (hs *ConnectableHostStub) RemoveStreamHandler(pid protocol.ID) {
	hs.RemoveStreamHandlerCalled(pid)
}

// NewStream -
func (hs *ConnectableHostStub) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	return hs.NewStreamCalled(ctx, p, pids...)
}

// Close -
func (hs *ConnectableHostStub) Close() error {
	return hs.CloseCalled()
}

// ConnManager -
func (hs *ConnectableHostStub) ConnManager() connmgr.ConnManager {
	return hs.ConnManagerCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (hs *ConnectableHostStub) IsInterfaceNil() bool {
	if hs == nil {
		return true
	}
	return false
}
