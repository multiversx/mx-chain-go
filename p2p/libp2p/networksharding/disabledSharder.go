package networksharding

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

// disabledSharder will not cause connections trimming
type disabledSharder struct{}

// NewDisabledSharder returns a disabled sharder implementation
func NewDisabledSharder() *disabledSharder {
	return &disabledSharder{}
}

// ComputeEvictionList will always output an empty list as to not cause connection trimming
func (ds *disabledSharder) ComputeEvictionList(_ []peer.ID) []peer.ID {
	return make([]peer.ID, 0)
}

// Has will output false, causing all peers to connect to each other
func (ds *disabledSharder) Has(_ peer.ID, _ []peer.ID) bool {
	return false
}

// SetPeerShardResolver will do nothing
func (ds *disabledSharder) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ds *disabledSharder) IsInterfaceNil() bool {
	return ds == nil
}
