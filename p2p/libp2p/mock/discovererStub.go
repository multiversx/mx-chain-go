package mock

import (
	"context"

	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-peerstore"
)

type DiscovererStub struct {
	FindPeersCalled func(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peerstore.PeerInfo, error)
}

func (ds *DiscovererStub) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peerstore.PeerInfo, error) {
	return ds.FindPeersCalled(ctx, ns, opts...)
}
