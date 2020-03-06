package mock

import (
	"context"

	"github.com/jbenet/goprocess"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

// NetworkStub -
type NetworkStub struct {
	ConnsToPeerCalled   func(p peer.ID) []network.Conn
	ConnsCalled         func() []network.Conn
	ConnectednessCalled func(peer.ID) network.Connectedness
	NotifyCalled        func(network.Notifiee)
	PeersCall           func() []peer.ID
	ClosePeerCall       func(peer.ID) error
}

// Peerstore -
func (ns *NetworkStub) Peerstore() peerstore.Peerstore {
	panic("implement me")
}

// LocalPeer -
func (ns *NetworkStub) LocalPeer() peer.ID {
	return "not a peer"
}

// DialPeer -
func (ns *NetworkStub) DialPeer(ctx context.Context, pid peer.ID) (network.Conn, error) {
	panic("implement me")
}

// ClosePeer -
func (ns *NetworkStub) ClosePeer(pid peer.ID) error {
	return ns.ClosePeerCall(pid)
}

// Connectedness -
func (ns *NetworkStub) Connectedness(pid peer.ID) network.Connectedness {
	return ns.ConnectednessCalled(pid)
}

// Peers -
func (ns *NetworkStub) Peers() []peer.ID {
	return ns.PeersCall()
}

// Conns -
func (ns *NetworkStub) Conns() []network.Conn {
	return ns.ConnsCalled()
}

// ConnsToPeer -
func (ns *NetworkStub) ConnsToPeer(p peer.ID) []network.Conn {
	return ns.ConnsToPeerCalled(p)
}

// Notify -
func (ns *NetworkStub) Notify(notifee network.Notifiee) {
	ns.NotifyCalled(notifee)
}

// StopNotify -
func (ns *NetworkStub) StopNotify(network.Notifiee) {
	panic("implement me")
}

// Close -
func (ns *NetworkStub) Close() error {
	panic("implement me")
}

// SetStreamHandler -
func (ns *NetworkStub) SetStreamHandler(network.StreamHandler) {
	panic("implement me")
}

// SetConnHandler -
func (ns *NetworkStub) SetConnHandler(network.ConnHandler) {
	panic("implement me")
}

// NewStream -
func (ns *NetworkStub) NewStream(context.Context, peer.ID) (network.Stream, error) {
	panic("implement me")
}

// Listen -
func (ns *NetworkStub) Listen(...multiaddr.Multiaddr) error {
	panic("implement me")
}

// ListenAddresses -
func (ns *NetworkStub) ListenAddresses() []multiaddr.Multiaddr {
	panic("implement me")
}

// InterfaceListenAddresses -
func (ns *NetworkStub) InterfaceListenAddresses() ([]multiaddr.Multiaddr, error) {
	panic("implement me")
}

// Process -
func (ns *NetworkStub) Process() goprocess.Process {
	panic("implement me")
}
