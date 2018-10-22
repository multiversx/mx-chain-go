package mock

import (
	cr "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type ConnMock struct {
	LocalP  peer.ID
	RemoteP peer.ID
}

func (cm *ConnMock) Close() error {
	return nil
}

func (cm *ConnMock) LocalPeer() peer.ID {
	return cm.LocalP
}

func (cm *ConnMock) LocalPrivateKey() cr.PrivKey {
	panic("implement me")
}

func (cm *ConnMock) RemotePeer() peer.ID {
	return cm.RemoteP
}

func (cm ConnMock) RemotePublicKey() cr.PubKey {
	panic("implement me")
}

func (cm ConnMock) LocalMultiaddr() multiaddr.Multiaddr {
	panic("implement me")
}

func (cm ConnMock) RemoteMultiaddr() multiaddr.Multiaddr {
	panic("implement me")
}

func (cm ConnMock) NewStream() (net.Stream, error) {
	panic("implement me")
}

func (cm ConnMock) GetStreams() []net.Stream {
	panic("implement me")
}

func (cm ConnMock) Stat() net.Stat {
	panic("implement me")
}
