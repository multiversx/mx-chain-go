package mock

import (
	"context"

	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multistream"
)

type HostStub struct {
	IDCalled                    func() peer.ID
	PeerstoreCalled             func() peerstore.Peerstore
	AddrsCalled                 func() []multiaddr.Multiaddr
	NetworkCalled               func() net.Network
	MuxCalled                   func() *multistream.MultistreamMuxer
	ConnectCalled               func(ctx context.Context, pi peerstore.PeerInfo) error
	SetStreamHandlerCalled      func(pid protocol.ID, handler net.StreamHandler)
	SetStreamHandlerMatchCalled func(protocol.ID, func(string) bool, net.StreamHandler)
	RemoveStreamHandlerCalled   func(pid protocol.ID)
	NewStreamCalled             func(ctx context.Context, p peer.ID, pids ...protocol.ID) (net.Stream, error)
	CloseCalled                 func() error
	ConnManagerCalled           func() ifconnmgr.ConnManager
}

func (hs *HostStub) ID() peer.ID {
	return hs.IDCalled()
}

func (hs *HostStub) Peerstore() peerstore.Peerstore {
	return hs.PeerstoreCalled()
}

func (hs *HostStub) Addrs() []multiaddr.Multiaddr {
	return hs.AddrsCalled()
}

func (hs *HostStub) Network() net.Network {
	return hs.NetworkCalled()
}

func (hs *HostStub) Mux() *multistream.MultistreamMuxer {
	return hs.MuxCalled()
}

func (hs *HostStub) Connect(ctx context.Context, pi peerstore.PeerInfo) error {
	return hs.ConnectCalled(ctx, pi)
}

func (hs *HostStub) SetStreamHandler(pid protocol.ID, handler net.StreamHandler) {
	hs.SetStreamHandlerCalled(pid, handler)
}

func (hs *HostStub) SetStreamHandlerMatch(pid protocol.ID, handler func(string) bool, streamHandler net.StreamHandler) {
	hs.SetStreamHandlerMatchCalled(pid, handler, streamHandler)
}

func (hs *HostStub) RemoveStreamHandler(pid protocol.ID) {
	hs.RemoveStreamHandlerCalled(pid)
}

func (hs *HostStub) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (net.Stream, error) {
	return hs.NewStreamCalled(ctx, p, pids...)
}

func (hs *HostStub) Close() error {
	return hs.CloseCalled()
}

func (hs *HostStub) ConnManager() ifconnmgr.ConnManager {
	return hs.ConnManagerCalled()
}
