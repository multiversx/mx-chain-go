package networksharding

import (
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
)

type sortingID struct {
	id       peer.ID
	key      []byte
	shard    uint32
	distance *big.Int
}

type sortingList struct {
	ref   sortingID
	peers []sortingID
}

// Len is the number of elements in the collection.
func (sl *sortingList) Len() int {
	return len(sl.peers)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (sl *sortingList) Less(i int, j int) bool {
	return sl.peers[i].distance.Cmp(sl.peers[j].distance) < 0
}

// Swap swaps the elements with indexes i and j.
func (sl *sortingList) Swap(i int, j int) {
	sl.peers[i], sl.peers[j] = sl.peers[j], sl.peers[i]
}
