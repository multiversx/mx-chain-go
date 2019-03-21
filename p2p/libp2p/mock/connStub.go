package mock

import (
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type ConnStub struct {
	CloseCalled           func() error
	LocalPeerCalled       func() peer.ID
	LocalPrivateKeyCalled func() crypto.PrivKey
	RemotePeerCalled      func() peer.ID
	RemotePublicKeyCalled func() crypto.PubKey
	LocalMultiaddrCalled  func() multiaddr.Multiaddr
	RemoteMultiaddrCalled func() multiaddr.Multiaddr
	NewStreamCalled       func() (net.Stream, error)
	GetStreamsCalled      func() []net.Stream
	StatCalled            func() net.Stat
}

func (cs *ConnStub) Close() error {
	return cs.CloseCalled()
}

func (cs *ConnStub) LocalPeer() peer.ID {
	return cs.LocalPeerCalled()
}

func (cs *ConnStub) LocalPrivateKey() crypto.PrivKey {
	return cs.LocalPrivateKeyCalled()
}

func (cs *ConnStub) RemotePeer() peer.ID {
	return cs.RemotePeerCalled()
}

func (cs *ConnStub) RemotePublicKey() crypto.PubKey {
	return cs.RemotePublicKeyCalled()
}

func (cs *ConnStub) LocalMultiaddr() multiaddr.Multiaddr {
	return cs.LocalMultiaddrCalled()
}

func (cs *ConnStub) RemoteMultiaddr() multiaddr.Multiaddr {
	return cs.RemoteMultiaddrCalled()
}

func (cs *ConnStub) NewStream() (net.Stream, error) {
	return cs.NewStreamCalled()
}

func (cs *ConnStub) GetStreams() []net.Stream {
	return cs.GetStreamsCalled()
}

func (cs *ConnStub) Stat() net.Stat {
	return cs.StatCalled()
}
