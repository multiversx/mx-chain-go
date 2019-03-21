package mock

import (
	"context"

	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type ConnManagerNotifieeStub struct {
	TagPeerCalled       func(p peer.ID, tag string, val int)
	UntagPeerCalled     func(p peer.ID, tag string)
	GetTagInfoCalled    func(p peer.ID) *ifconnmgr.TagInfo
	TrimOpenConnsCalled func(ctx context.Context)

	ListenCalled       func(netw net.Network, ma multiaddr.Multiaddr)
	ListenCloseCalled  func(netw net.Network, ma multiaddr.Multiaddr)
	ConnectedCalled    func(netw net.Network, conn net.Conn)
	DisconnectedCalled func(netw net.Network, conn net.Conn)
	OpenedStreamCalled func(netw net.Network, stream net.Stream)
	ClosedStreamCalled func(netw net.Network, stream net.Stream)
}

func (cmns *ConnManagerNotifieeStub) TagPeer(p peer.ID, tag string, val int) {
	cmns.TagPeerCalled(p, tag, val)
}

func (cmns *ConnManagerNotifieeStub) UntagPeer(p peer.ID, tag string) {
	cmns.UntagPeerCalled(p, tag)
}

func (cmns *ConnManagerNotifieeStub) GetTagInfo(p peer.ID) *ifconnmgr.TagInfo {
	return cmns.GetTagInfoCalled(p)
}

func (cmns *ConnManagerNotifieeStub) TrimOpenConns(ctx context.Context) {
	cmns.TrimOpenConnsCalled(ctx)
}

func (cmns *ConnManagerNotifieeStub) Notifee() net.Notifiee {
	return cmns
}

func (cmns *ConnManagerNotifieeStub) Listen(netw net.Network, ma multiaddr.Multiaddr) {
	cmns.ListenCalled(netw, ma)
}

func (cmns *ConnManagerNotifieeStub) ListenClose(netw net.Network, ma multiaddr.Multiaddr) {
	cmns.ListenCloseCalled(netw, ma)
}

func (cmns *ConnManagerNotifieeStub) Connected(netw net.Network, conn net.Conn) {
	cmns.ConnectedCalled(netw, conn)
}

func (cmns *ConnManagerNotifieeStub) Disconnected(netw net.Network, conn net.Conn) {
	cmns.DisconnectedCalled(netw, conn)
}

func (cmns *ConnManagerNotifieeStub) OpenedStream(netw net.Network, stream net.Stream) {
	cmns.OpenedStreamCalled(netw, stream)
}

func (cmns *ConnManagerNotifieeStub) ClosedStream(netw net.Network, stream net.Stream) {
	cmns.ClosedStreamCalled(netw, stream)
}
