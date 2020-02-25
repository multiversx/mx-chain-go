package sorting

import (
	"math/big"
	"sort"

	"github.com/libp2p/go-libp2p-core/peer"
)

// SortingID contains the peer data
type SortingID struct {
	ID       peer.ID
	Key      []byte
	Shard    uint32
	Distance *big.Int
}

// SortingList holds a sorted list of elements in respect with the reference value
type SortingList struct {
	Ref   SortingID
	Peers []SortingID
}

// Len is the number of elements in the collection.
func (sl *SortingList) Len() int {
	return len(sl.Peers)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (sl *SortingList) Less(i int, j int) bool {
	return sl.Peers[i].Distance.Cmp(sl.Peers[j].Distance) < 0
}

// Swap swaps the elements with indexes i and j.
func (sl *SortingList) Swap(i int, j int) {
	sl.Peers[i], sl.Peers[j] = sl.Peers[j], sl.Peers[i]
}

// SortedPeers get the orted list of peers
func (sl *SortingList) SortedPeers() []peer.ID {
	sort.Sort(sl)
	ret := make([]peer.ID, len(sl.Peers))

	for i, id := range sl.Peers {
		ret[i] = id.ID
	}
	return ret
}
