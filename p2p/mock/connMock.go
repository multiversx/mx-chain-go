package mock

import (
	cr "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

// ConnMock is used for testing
type ConnMock struct {
	LocalP  peer.ID
	RemoteP peer.ID
	Status  net.Stat

	CloseCalled func(*ConnMock) error
}

// Closes the connection, dummy
func (cm *ConnMock) Close() error {
	if cm.CloseCalled != nil {
		return cm.CloseCalled(cm)
	}

	return nil
}

// LocalPeer returns the current peer
func (cm *ConnMock) LocalPeer() peer.ID {
	return cm.LocalP
}

// LocalPrivateKey returns current private key used in P2P
func (cm *ConnMock) LocalPrivateKey() cr.PrivKey {
	panic("implement me")
}

// RemotePeer returns the other side peer
func (cm *ConnMock) RemotePeer() peer.ID {
	return cm.RemoteP
}

// RemotePublicKey returns the public key of the other side peer, dummy, will panic
func (cm ConnMock) RemotePublicKey() cr.PubKey {
	panic("implement me")
}

// LocalMultiaddr is dummy, will panic
func (cm ConnMock) LocalMultiaddr() multiaddr.Multiaddr {
	panic("implement me")
}

// RemoteMultiaddr is dummy, will panic
func (cm ConnMock) RemoteMultiaddr() multiaddr.Multiaddr {
	panic("implement me")
}

// NewStream is dummy, will panic
func (cm ConnMock) NewStream() (net.Stream, error) {
	panic("implement me")
}

// GetStreams is dummy, will panic
func (cm ConnMock) GetStreams() []net.Stream {
	panic("implement me")
}

// Stat is dummy, will panic
func (cm ConnMock) Stat() net.Stat {
	return cm.Status
}
