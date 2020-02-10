package mock

import (
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ConnStub -
type ConnStub struct {
	CloseCalled           func() error
	LocalPeerCalled       func() peer.ID
	LocalPrivateKeyCalled func() libp2pCrypto.PrivKey
	RemotePeerCalled      func() peer.ID
	RemotePublicKeyCalled func() libp2pCrypto.PubKey
	LocalMultiaddrCalled  func() multiaddr.Multiaddr
	RemoteMultiaddrCalled func() multiaddr.Multiaddr
	NewStreamCalled       func() (network.Stream, error)
	GetStreamsCalled      func() []network.Stream
	StatCalled            func() network.Stat
}

// Close -
func (cs *ConnStub) Close() error {
	return cs.CloseCalled()
}

// LocalPeer -
func (cs *ConnStub) LocalPeer() peer.ID {
	return cs.LocalPeerCalled()
}

// LocalPrivateKey -
func (cs *ConnStub) LocalPrivateKey() libp2pCrypto.PrivKey {
	return cs.LocalPrivateKeyCalled()
}

// RemotePeer -
func (cs *ConnStub) RemotePeer() peer.ID {
	return cs.RemotePeerCalled()
}

// RemotePublicKey -
func (cs *ConnStub) RemotePublicKey() libp2pCrypto.PubKey {
	return cs.RemotePublicKeyCalled()
}

// LocalMultiaddr -
func (cs *ConnStub) LocalMultiaddr() multiaddr.Multiaddr {
	return cs.LocalMultiaddrCalled()
}

// RemoteMultiaddr -
func (cs *ConnStub) RemoteMultiaddr() multiaddr.Multiaddr {
	return cs.RemoteMultiaddrCalled()
}

// NewStream -
func (cs *ConnStub) NewStream() (network.Stream, error) {
	return cs.NewStreamCalled()
}

// GetStreams -
func (cs *ConnStub) GetStreams() []network.Stream {
	return cs.GetStreamsCalled()
}

// Stat -
func (cs *ConnStub) Stat() network.Stat {
	return cs.StatCalled()
}
