package mock

import (
	"context"
	"errors"

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
	AddressToPeerInfoCalled     func(address string) (*peer.AddrInfo, error)
}

// EventBus -
func (hs *ConnectableHostStub) EventBus() event.Bus {
	if hs.EventBusCalled != nil {
		return hs.EventBusCalled()
	}

	return &EventBusStub{}
}

// ConnectToPeer -
func (hs *ConnectableHostStub) ConnectToPeer(ctx context.Context, address string) error {
	if hs.ConnectToPeerCalled != nil {
		return hs.ConnectToPeerCalled(ctx, address)
	}

	return nil
}

// ID -
func (hs *ConnectableHostStub) ID() peer.ID {
	if hs.IDCalled != nil {
		return hs.IDCalled()
	}

	return "mock pid"
}

// Peerstore -
func (hs *ConnectableHostStub) Peerstore() peerstore.Peerstore {
	if hs.PeerstoreCalled != nil {
		return hs.PeerstoreCalled()
	}

	return nil
}

// Addrs -
func (hs *ConnectableHostStub) Addrs() []multiaddr.Multiaddr {
	if hs.AddrsCalled != nil {
		return hs.AddrsCalled()
	}

	return make([]multiaddr.Multiaddr, 0)
}

// Network -
func (hs *ConnectableHostStub) Network() network.Network {
	if hs.NetworkCalled != nil {
		return hs.NetworkCalled()
	}

	return &NetworkStub{}
}

// Mux -
func (hs *ConnectableHostStub) Mux() protocol.Switch {
	if hs.MuxCalled != nil {
		return hs.MuxCalled()
	}

	return nil
}

// Connect -
func (hs *ConnectableHostStub) Connect(ctx context.Context, pi peer.AddrInfo) error {
	if hs.ConnectCalled != nil {
		return hs.ConnectCalled(ctx, pi)
	}

	return nil
}

// SetStreamHandler -
func (hs *ConnectableHostStub) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	if hs.SetStreamHandlerCalled != nil {
		hs.SetStreamHandlerCalled(pid, handler)
	}
}

// SetStreamHandlerMatch -
func (hs *ConnectableHostStub) SetStreamHandlerMatch(pid protocol.ID, handler func(string) bool, streamHandler network.StreamHandler) {
	if hs.SetStreamHandlerMatchCalled != nil {
		hs.SetStreamHandlerMatchCalled(pid, handler, streamHandler)
	}
}

// RemoveStreamHandler -
func (hs *ConnectableHostStub) RemoveStreamHandler(pid protocol.ID) {
	if hs.RemoveStreamHandlerCalled != nil {
		hs.RemoveStreamHandlerCalled(pid)
	}
}

// NewStream -
func (hs *ConnectableHostStub) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	if hs.NewStreamCalled != nil {
		return hs.NewStreamCalled(ctx, p, pids...)
	}

	return nil, errors.New("no stream")
}

// Close -
func (hs *ConnectableHostStub) Close() error {
	if hs.CloseCalled != nil {
		return hs.CloseCalled()
	}

	return nil
}

// ConnManager -
func (hs *ConnectableHostStub) ConnManager() connmgr.ConnManager {
	if hs.ConnManagerCalled != nil {
		return hs.ConnManagerCalled()
	}

	return nil
}

// AddressToPeerInfo -
func (hs *ConnectableHostStub) AddressToPeerInfo(address string) (*peer.AddrInfo, error) {
	if hs.AddressToPeerInfoCalled != nil {
		return hs.AddressToPeerInfoCalled(address)
	}

	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return nil, err
	}

	return peer.AddrInfoFromP2pAddr(multiAddr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hs *ConnectableHostStub) IsInterfaceNil() bool {
	return hs == nil
}
