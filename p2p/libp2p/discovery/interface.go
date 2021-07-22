package discovery

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

// ConnectableHost is an enhanced Host interface that has the ability to connect to a string address
type ConnectableHost interface {
	host.Host
	ConnectToPeer(ctx context.Context, address string) error
	AddressToPeerInfo(address string) (*peer.AddrInfo, error)
	IsInterfaceNil() bool
}

// Sharder defines the eviction computing process of unwanted peers
type Sharder interface {
	ComputeEvictionList(pidList []peer.ID) []peer.ID
	Has(pid peer.ID, list []peer.ID) bool
	SetSeeders(addresses []string)
	IsSeeder(pid core.PeerID) bool
	IsInterfaceNil() bool
}

// KadDhtHandler defines the behavior of a component that can find new peers in a p2p network through kad dht mechanism
type KadDhtHandler interface {
	Bootstrap(ctx context.Context) error
}
