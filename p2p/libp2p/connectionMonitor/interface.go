package connectionMonitor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Sharder defines the eviction computing process of unwanted peers
type Sharder interface {
	ComputeEvictionList(pidList []peer.ID) []peer.ID
	Has(pid peer.ID, list []peer.ID) bool
	SetSeeders(addresses []string)
	IsSeeder(pid core.PeerID) bool
	IsInterfaceNil() bool
}
