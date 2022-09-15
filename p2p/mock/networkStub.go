package mock

import (
	"context"
	"errors"

	"github.com/jbenet/goprocess"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

// NetworkStub -
type NetworkStub struct {
	ConnsToPeerCalled     func(p peer.ID) []network.Conn
	ConnsCalled           func() []network.Conn
	ConnectednessCalled   func(peer.ID) network.Connectedness
	NotifyCalled          func(network.Notifiee)
	StopNotifyCalled      func(network.Notifiee)
	PeersCall             func() []peer.ID
	ClosePeerCall         func(peer.ID) error
	ResourceManagerCalled func() network.ResourceManager
}

// ResourceManager -
func (ns *NetworkStub) ResourceManager() network.ResourceManager {
	if ns.ResourceManagerCalled != nil {
		return ns.ResourceManagerCalled()
	}

	return nil
}

// Peerstore -
func (ns *NetworkStub) Peerstore() peerstore.Peerstore {
	return nil
}

// LocalPeer -
func (ns *NetworkStub) LocalPeer() peer.ID {
	return "not a peer"
}

// DialPeer -
func (ns *NetworkStub) DialPeer(_ context.Context, _ peer.ID) (network.Conn, error) {
	return nil, errors.New("dial error")
}

// ClosePeer -
func (ns *NetworkStub) ClosePeer(pid peer.ID) error {
	if ns.ClosePeerCall != nil {
		return ns.ClosePeerCall(pid)
	}

	return nil
}

// Connectedness -
func (ns *NetworkStub) Connectedness(pid peer.ID) network.Connectedness {
	if ns.ConnectednessCalled != nil {
		return ns.ConnectednessCalled(pid)
	}

	return network.NotConnected
}

// Peers -
func (ns *NetworkStub) Peers() []peer.ID {
	if ns.PeersCall != nil {
		return ns.PeersCall()
	}

	return make([]peer.ID, 0)
}

// Conns -
func (ns *NetworkStub) Conns() []network.Conn {
	if ns.ConnsCalled != nil {
		return ns.ConnsCalled()
	}

	return make([]network.Conn, 0)
}

// ConnsToPeer -
func (ns *NetworkStub) ConnsToPeer(p peer.ID) []network.Conn {
	if ns.ConnsToPeerCalled != nil {
		return ns.ConnsToPeerCalled(p)
	}

	return make([]network.Conn, 0)
}

// Notify -
func (ns *NetworkStub) Notify(notifee network.Notifiee) {
	if ns.NotifyCalled != nil {
		ns.NotifyCalled(notifee)
	}
}

// StopNotify -
func (ns *NetworkStub) StopNotify(notifee network.Notifiee) {
	if ns.StopNotifyCalled != nil {
		ns.StopNotifyCalled(notifee)
	}
}

// Close -
func (ns *NetworkStub) Close() error {
	return nil
}

// SetStreamHandler -
func (ns *NetworkStub) SetStreamHandler(network.StreamHandler) {}

// NewStream -
func (ns *NetworkStub) NewStream(context.Context, peer.ID) (network.Stream, error) {
	return nil, errors.New("new stream error")
}

// Listen -
func (ns *NetworkStub) Listen(...multiaddr.Multiaddr) error {
	return nil
}

// ListenAddresses -
func (ns *NetworkStub) ListenAddresses() []multiaddr.Multiaddr {
	return make([]multiaddr.Multiaddr, 0)
}

// InterfaceListenAddresses -
func (ns *NetworkStub) InterfaceListenAddresses() ([]multiaddr.Multiaddr, error) {
	return make([]multiaddr.Multiaddr, 0), nil
}

// Process -
func (ns *NetworkStub) Process() goprocess.Process {
	return nil
}
