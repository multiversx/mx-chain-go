package networksharding

import (
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
)

// peerDistance is a composite struct on top of a peer ID that also contains the kad distance measured
// against the current peer and held as a big.Int
type peerDistance struct {
	peer.ID
	distance *big.Int
}

// peerDistances represents a sortable peerDistance slice
type peerDistances []peerDistance

// Len returns the length of this slice
func (pd peerDistances) Len() int {
	return len(pd)
}

// Less is used in sorting and returns if i-th element is less than j-th element
func (pd peerDistances) Less(i, j int) bool {
	return pd[i].distance.Cmp(pd[j].distance) < 0
}

// Swap is used in sorting and swaps the values between the i-th position with the one found on j-th position
func (pd peerDistances) Swap(i, j int) {
	pd[i], pd[j] = pd[j], pd[i]
}
