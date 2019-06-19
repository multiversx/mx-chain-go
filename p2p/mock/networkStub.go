package mock

import (
	"context"

	"github.com/jbenet/goprocess"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

type NetworkStub struct {
	ConnsToPeerCalled   func(p peer.ID) []network.Conn
	ConnsCalled         func() []network.Conn
	ConnectednessCalled func(peer.ID) network.Connectedness
	NotifyCalled        func(network.Notifiee)
}

func (ns *NetworkStub) Peerstore() peerstore.Peerstore {
	panic("implement me")
}

func (ns *NetworkStub) LocalPeer() peer.ID {
	panic("implement me")
}

func (ns *NetworkStub) DialPeer(ctx context.Context, pid peer.ID) (network.Conn, error) {
	panic("implement me")
}

func (ns *NetworkStub) ClosePeer(pid peer.ID) error {
	panic("implement me")
}

func (ns *NetworkStub) Connectedness(pid peer.ID) network.Connectedness {
	return ns.ConnectednessCalled(pid)
}

func (ns *NetworkStub) Peers() []peer.ID {
	panic("implement me")
}

func (ns *NetworkStub) Conns() []network.Conn {
	return ns.ConnsCalled()
}

func (ns *NetworkStub) ConnsToPeer(p peer.ID) []network.Conn {
	return ns.ConnsToPeerCalled(p)
}

func (ns *NetworkStub) Notify(notifee network.Notifiee) {
	ns.NotifyCalled(notifee)
}

func (ns *NetworkStub) StopNotify(network.Notifiee) {
	panic("implement me")
}

func (ns *NetworkStub) Close() error {
	panic("implement me")
}

func (ns *NetworkStub) SetStreamHandler(network.StreamHandler) {
	panic("implement me")
}

func (ns *NetworkStub) SetConnHandler(network.ConnHandler) {
	panic("implement me")
}

func (ns *NetworkStub) NewStream(context.Context, peer.ID) (network.Stream, error) {
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
