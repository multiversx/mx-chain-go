package connectionMonitor

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Sharder defines the eviction computing process of unwanted peers
type Sharder interface {
	ComputeEvictList(pid peer.ID, connected []peer.ID) []peer.ID
	Has(pid peer.ID, list []peer.ID) bool
	PeerShardResolver() p2p.PeerShardResolver
	IsInterfaceNil() bool
}
