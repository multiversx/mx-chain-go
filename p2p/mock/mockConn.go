package mock

import (
	cr "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type MockConn struct {
	LocalP  peer.ID
	RemoteP peer.ID
}

func (mc *MockConn) Close() error {
	return nil
}

func (mc *MockConn) LocalPeer() peer.ID {
	return mc.LocalP
}

func (mc *MockConn) LocalPrivateKey() cr.PrivKey {
	panic("implement me")
}

func (mc *MockConn) RemotePeer() peer.ID {
	return mc.RemoteP
}

func (mc MockConn) RemotePublicKey() cr.PubKey {
	panic("implement me")
}

func (mc MockConn) LocalMultiaddr() multiaddr.Multiaddr {
	panic("implement me")
}

func (mc MockConn) RemoteMultiaddr() multiaddr.Multiaddr {
	panic("implement me")
}

func (mc MockConn) NewStream() (net.Stream, error) {
	panic("implement me")
}

func (mc MockConn) GetStreams() []net.Stream {
	panic("implement me")
}

func (mc MockConn) Stat() net.Stat {
	panic("implement me")
}
