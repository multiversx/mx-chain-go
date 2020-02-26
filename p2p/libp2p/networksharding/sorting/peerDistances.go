package sorting

import (
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
)

// PeerDistance is a composite struct on top of a peer ID that also contains the kad distance measured
// against the current peer and held as a big.Int
type PeerDistance struct {
	peer.ID
	Distance *big.Int
}

// PeerDistances represents a sortable peerDistance slice
type PeerDistances []*PeerDistance

// Len returns the length of this slice
func (pd PeerDistances) Len() int {
	return len(pd)
}

// Less is used in sorting and returns if i-th element is less than j-th element
func (pd PeerDistances) Less(i, j int) bool {
	return pd[i].Distance.Cmp(pd[j].Distance) < 0
}

// Swap is used in sorting and swaps the values between the i-th position with the one found on j-th position
func (pd PeerDistances) Swap(i, j int) {
	pd[i], pd[j] = pd[j], pd[i]
}
