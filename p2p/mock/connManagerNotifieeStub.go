package mock

import (
	"context"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type ConnManagerNotifieeStub struct {
	UpsertTagCalled     func(p peer.ID, tag string, upsert func(int) int)
	ProtectCalled       func(id peer.ID, tag string)
	UnprotectCalled     func(id peer.ID, tag string) (protected bool)
	CloseCalled         func() error
	TagPeerCalled       func(p peer.ID, tag string, val int)
	UntagPeerCalled     func(p peer.ID, tag string)
	GetTagInfoCalled    func(p peer.ID) *connmgr.TagInfo
	TrimOpenConnsCalled func(ctx context.Context)
	ListenCalled        func(netw network.Network, ma multiaddr.Multiaddr)
	ListenCloseCalled   func(netw network.Network, ma multiaddr.Multiaddr)
	ConnectedCalled     func(netw network.Network, conn network.Conn)
	DisconnectedCalled  func(netw network.Network, conn network.Conn)
	OpenedStreamCalled  func(netw network.Network, stream network.Stream)
	ClosedStreamCalled  func(netw network.Network, stream network.Stream)
}

func (cmns *ConnManagerNotifieeStub) UpsertTag(p peer.ID, tag string, upsert func(int) int) {
	cmns.UpsertTagCalled(p, tag, upsert)
}

func (cmns *ConnManagerNotifieeStub) Protect(id peer.ID, tag string) {
	cmns.ProtectCalled(id, tag)
}

func (cmns *ConnManagerNotifieeStub) Unprotect(id peer.ID, tag string) (protected bool) {
	return cmns.UnprotectCalled(id, tag)
}

func (cmns *ConnManagerNotifieeStub) Close() error {
	return cmns.CloseCalled()
}

func (cmns *ConnManagerNotifieeStub) TagPeer(p peer.ID, tag string, val int) {
	cmns.TagPeerCalled(p, tag, val)
}

func (cmns *ConnManagerNotifieeStub) UntagPeer(p peer.ID, tag string) {
	cmns.UntagPeerCalled(p, tag)
}

func (cmns *ConnManagerNotifieeStub) GetTagInfo(p peer.ID) *connmgr.TagInfo {
	return cmns.GetTagInfoCalled(p)
}

func (cmns *ConnManagerNotifieeStub) TrimOpenConns(ctx context.Context) {
	cmns.TrimOpenConnsCalled(ctx)
}

func (cmns *ConnManagerNotifieeStub) Notifee() network.Notifiee {
	return cmns
}

func (cmns *ConnManagerNotifieeStub) Listen(netw network.Network, ma multiaddr.Multiaddr) {
	cmns.ListenCalled(netw, ma)
}

func (cmns *ConnManagerNotifieeStub) ListenClose(netw network.Network, ma multiaddr.Multiaddr) {
	cmns.ListenCloseCalled(netw, ma)
}

func (cmns *ConnManagerNotifieeStub) Connected(netw network.Network, conn network.Conn) {
	cmns.ConnectedCalled(netw, conn)
}

func (cmns *ConnManagerNotifieeStub) Disconnected(netw network.Network, conn network.Conn) {
	cmns.DisconnectedCalled(netw, conn)
}

func (cmns *ConnManagerNotifieeStub) OpenedStream(netw network.Network, stream network.Stream) {
	cmns.OpenedStreamCalled(netw, stream)
}

func (cmns *ConnManagerNotifieeStub) ClosedStream(netw network.Network, stream network.Stream) {
	cmns.ClosedStreamCalled(netw, stream)
}
