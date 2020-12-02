package networksharding

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ p2p.Sharder = (*nilListSharder)(nil)

// nilListSharder will not cause connections trimming
type nilListSharder struct{}

// NewNilListSharder returns a disabled sharder implementation
func NewNilListSharder() *nilListSharder {
	return &nilListSharder{}
}

// ComputeEvictionList will always output an empty list as to not cause connection trimming
func (nls *nilListSharder) ComputeEvictionList(_ []peer.ID) []peer.ID {
	return make([]peer.ID, 0)
}

// Has will output false, causing all peers to connect to each other
func (nls *nilListSharder) Has(_ peer.ID, _ []peer.ID) bool {
	return false
}

// SetPeerShardResolver will do nothing
func (nls *nilListSharder) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// SetSeeders does nothing
func (nls *nilListSharder) SetSeeders(_ []string) {
}

// IsSeeder returns false
func (nls *nilListSharder) IsSeeder(_ core.PeerID) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (nls *nilListSharder) IsInterfaceNil() bool {
	return nls == nil
}
