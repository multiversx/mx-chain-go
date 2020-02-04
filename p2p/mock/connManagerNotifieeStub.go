package mock

import (
	"context"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ConnManagerNotifieeStub -
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

// UpsertTag -
func (cmns *ConnManagerNotifieeStub) UpsertTag(p peer.ID, tag string, upsert func(int) int) {
	cmns.UpsertTagCalled(p, tag, upsert)
}

// Protect -
func (cmns *ConnManagerNotifieeStub) Protect(id peer.ID, tag string) {
	cmns.ProtectCalled(id, tag)
}

// Unprotect -
func (cmns *ConnManagerNotifieeStub) Unprotect(id peer.ID, tag string) (protected bool) {
	return cmns.UnprotectCalled(id, tag)
}

// Close -
func (cmns *ConnManagerNotifieeStub) Close() error {
	return cmns.CloseCalled()
}

// TagPeer -
func (cmns *ConnManagerNotifieeStub) TagPeer(p peer.ID, tag string, val int) {
	cmns.TagPeerCalled(p, tag, val)
}

// UntagPeer -
func (cmns *ConnManagerNotifieeStub) UntagPeer(p peer.ID, tag string) {
	cmns.UntagPeerCalled(p, tag)
}

// GetTagInfo -
func (cmns *ConnManagerNotifieeStub) GetTagInfo(p peer.ID) *connmgr.TagInfo {
	return cmns.GetTagInfoCalled(p)
}

// TrimOpenConns -
func (cmns *ConnManagerNotifieeStub) TrimOpenConns(ctx context.Context) {
	cmns.TrimOpenConnsCalled(ctx)
}

// Notifee -
func (cmns *ConnManagerNotifieeStub) Notifee() network.Notifiee {
	return cmns
}

// Listen -
func (cmns *ConnManagerNotifieeStub) Listen(netw network.Network, ma multiaddr.Multiaddr) {
	cmns.ListenCalled(netw, ma)
}

// ListenClose -
func (cmns *ConnManagerNotifieeStub) ListenClose(netw network.Network, ma multiaddr.Multiaddr) {
	cmns.ListenCloseCalled(netw, ma)
}

// Connected -
func (cmns *ConnManagerNotifieeStub) Connected(netw network.Network, conn network.Conn) {
	cmns.ConnectedCalled(netw, conn)
}

// Disconnected -
func (cmns *ConnManagerNotifieeStub) Disconnected(netw network.Network, conn network.Conn) {
	cmns.DisconnectedCalled(netw, conn)
}

// OpenedStream -
func (cmns *ConnManagerNotifieeStub) OpenedStream(netw network.Network, stream network.Stream) {
	cmns.OpenedStreamCalled(netw, stream)
}

// ClosedStream -
func (cmns *ConnManagerNotifieeStub) ClosedStream(netw network.Network, stream network.Stream) {
	cmns.ClosedStreamCalled(netw, stream)
}
