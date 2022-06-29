package mock

import (
	"context"

	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ConnStub -
type ConnStub struct {
	IDCalled              func() string
	CloseCalled           func() error
	LocalPeerCalled       func() peer.ID
	LocalPrivateKeyCalled func() libp2pCrypto.PrivKey
	RemotePeerCalled      func() peer.ID
	RemotePublicKeyCalled func() libp2pCrypto.PubKey
	LocalMultiaddrCalled  func() multiaddr.Multiaddr
	RemoteMultiaddrCalled func() multiaddr.Multiaddr
	NewStreamCalled       func(ctx context.Context) (network.Stream, error)
	GetStreamsCalled      func() []network.Stream
	StatCalled            func() network.ConnStats
	ScopeCalled           func() network.ConnScope
}

// ID -
func (cs *ConnStub) ID() string {
	if cs.IDCalled != nil {
		return cs.IDCalled()
	}

	return ""
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
	if cs.RemoteMultiaddrCalled != nil {
		return cs.RemoteMultiaddrCalled()
	}

	ma, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk")
	return ma
}

// NewStream -
func (cs *ConnStub) NewStream(ctx context.Context) (network.Stream, error) {
	return cs.NewStreamCalled(ctx)
}

// GetStreams -
func (cs *ConnStub) GetStreams() []network.Stream {
	return cs.GetStreamsCalled()
}

// Stat -
func (cs *ConnStub) Stat() network.ConnStats {
	if cs.StatCalled != nil {
		return cs.StatCalled()
	}

	return network.ConnStats{}
}

// Scope -
func (cs *ConnStub) Scope() network.ConnScope {
	if cs.ScopeCalled != nil {
		cs.ScopeCalled()
	}

	return network.NullScope
}
