package libp2p

import (
	"context"

	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
)

// PeerInfoHandler is the signature of the handler that gets called whenever an action for a peerInfo is triggered
type PeerInfoHandler func(pInfo peerstore.PeerInfo)

// UpgradedHost is an enhanced Host interface
type UpgradedHost interface {
	host.Host
	ConnectToPeer(ctx context.Context, address string) error
}

type upgradedHost struct {
	host.Host
}

func NewUpgradedHost(h host.Host) *upgradedHost {
	return &upgradedHost{
		Host: h,
	}
}

func (updHost *upgradedHost) ConnectToPeer(ctx context.Context, address string) error {
	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return err
	}

	pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
	if err != nil {
		return err
	}

	return updHost.Connect(ctx, *pInfo)
}
