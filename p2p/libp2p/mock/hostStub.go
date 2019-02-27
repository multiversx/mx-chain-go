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

type UpgradedHostStub struct {
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
	ConnectToPeerCalled         func(ctx context.Context, address string) error
}

func (hs *UpgradedHostStub) ConnectToPeer(ctx context.Context, address string) error {
	return hs.ConnectToPeerCalled(ctx, address)
}

func (hs *UpgradedHostStub) ID() peer.ID {
	return hs.IDCalled()
}

func (hs *UpgradedHostStub) Peerstore() peerstore.Peerstore {
	return hs.PeerstoreCalled()
}

func (hs *UpgradedHostStub) Addrs() []multiaddr.Multiaddr {
	return hs.AddrsCalled()
}

func (hs *UpgradedHostStub) Network() net.Network {
	return hs.NetworkCalled()
}

func (hs *UpgradedHostStub) Mux() *multistream.MultistreamMuxer {
	return hs.MuxCalled()
}

func (hs *UpgradedHostStub) Connect(ctx context.Context, pi peerstore.PeerInfo) error {
	return hs.ConnectCalled(ctx, pi)
}

func (hs *UpgradedHostStub) SetStreamHandler(pid protocol.ID, handler net.StreamHandler) {
	hs.SetStreamHandlerCalled(pid, handler)
}

func (hs *UpgradedHostStub) SetStreamHandlerMatch(pid protocol.ID, handler func(string) bool, streamHandler net.StreamHandler) {
	hs.SetStreamHandlerMatchCalled(pid, handler, streamHandler)
}

func (hs *UpgradedHostStub) RemoveStreamHandler(pid protocol.ID) {
	hs.RemoveStreamHandlerCalled(pid)
}

func (hs *UpgradedHostStub) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (net.Stream, error) {
	return hs.NewStreamCalled(ctx, p, pids...)
}

func (hs *UpgradedHostStub) Close() error {
	return hs.CloseCalled()
}

func (hs *UpgradedHostStub) ConnManager() ifconnmgr.ConnManager {
	return hs.ConnManagerCalled()
}
