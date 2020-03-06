package libp2p

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// PeerInfoHandler is the signature of the handler that gets called whenever an action for a peerInfo is triggered
type PeerInfoHandler func(pInfo peer.AddrInfo)

// ConnectableHost is an enhanced Host interface that has the ability to connect to a string address
type ConnectableHost interface {
	host.Host
	ConnectToPeer(ctx context.Context, address string) error
	IsInterfaceNil() bool
}

type connectableHost struct {
	host.Host
}

// NewConnectableHost creates a new connectable host implementation
func NewConnectableHost(h host.Host) *connectableHost {
	return &connectableHost{
		Host: h,
	}
}

// ConnectToPeer connects to a peer by knowing its string address
func (connHost *connectableHost) ConnectToPeer(ctx context.Context, address string) error {
	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return err
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		return err
	}

	return connHost.Connect(ctx, *pInfo)
}

// IsInterfaceNil returns true if there is no value under the interface
func (connHost *connectableHost) IsInterfaceNil() bool {
	return connHost == nil
}
