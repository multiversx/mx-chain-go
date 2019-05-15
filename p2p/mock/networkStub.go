package mock

import (
	"context"

	"github.com/jbenet/goprocess"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
)

type NetworkStub struct {
	ConnsToPeerCalled   func(p peer.ID) []net.Conn
	ConnsCalled         func() []net.Conn
	ConnectednessCalled func(peer.ID) net.Connectedness
	NotifyCalled        func(net.Notifiee)
}

func (ns *NetworkStub) Peerstore() peerstore.Peerstore {
	panic("implement me")
}

func (ns *NetworkStub) LocalPeer() peer.ID {
	panic("implement me")
}

func (ns *NetworkStub) DialPeer(ctx context.Context, pid peer.ID) (net.Conn, error) {
	panic("implement me")
}

func (ns *NetworkStub) ClosePeer(pid peer.ID) error {
	panic("implement me")
}

func (ns *NetworkStub) Connectedness(pid peer.ID) net.Connectedness {
	return ns.ConnectednessCalled(pid)
}

func (ns *NetworkStub) Peers() []peer.ID {
	panic("implement me")
}

func (ns *NetworkStub) Conns() []net.Conn {
	return ns.ConnsCalled()
}

func (ns *NetworkStub) ConnsToPeer(p peer.ID) []net.Conn {
	return ns.ConnsToPeerCalled(p)
}

func (ns *NetworkStub) Notify(notifee net.Notifiee) {
	ns.NotifyCalled(notifee)
}

func (ns *NetworkStub) StopNotify(net.Notifiee) {
	panic("implement me")
}

func (ns *NetworkStub) Close() error {
	panic("implement me")
}

func (ns *NetworkStub) SetStreamHandler(net.StreamHandler) {
	panic("implement me")
}

func (ns *NetworkStub) SetConnHandler(net.ConnHandler) {
	panic("implement me")
}

func (ns *NetworkStub) NewStream(context.Context, peer.ID) (net.Stream, error) {
	panic("implement me")
}

func (ns *NetworkStub) Listen(...multiaddr.Multiaddr) error {
	panic("implement me")
}

func (ns *NetworkStub) ListenAddresses() []multiaddr.Multiaddr {
	panic("implement me")
}

func (ns *NetworkStub) InterfaceListenAddresses() ([]multiaddr.Multiaddr, error) {
	panic("implement me")
}

func (ns *NetworkStub) Process() goprocess.Process {
	panic("implement me")
}
