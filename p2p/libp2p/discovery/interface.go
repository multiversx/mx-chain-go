package discovery

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Sharder defines the eviction computing process of unwanted peers
type Sharder interface {
	ComputeEvictList(pidList []peer.ID) []peer.ID
	Has(pid peer.ID, list []peer.ID) bool
	PeerShardResolver() p2p.PeerShardResolver
	IsInterfaceNil() bool
}

// ConnectableHost is an enhanced Host interface that has the ability to connect to a string address
type ConnectableHost interface {
	host.Host
	ConnectToPeer(ctx context.Context, address string) error
	IsInterfaceNil() bool
}
