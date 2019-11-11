package networksharding

import (
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
)

// NoSharder default sharder, only uses Kademlia distance in sorting
type noSharder struct {
}

// GetShard always 0
func (ns *noSharder) GetShard(id peer.ID) uint32 {
	return 0
}

// GetDistance Kademlia XOR distance
func (ns *noSharder) GetDistance(a, b sortingID) *big.Int {
	c := make([]byte, len(a.key))
	for i := 0; i < len(a.key); i++ {
		c[i] = a.key[i] ^ b.key[i]
	}

	ret := big.NewInt(0).SetBytes(c)
	return ret
}

// SortList sort the list
func (ns *noSharder) SortList(peers []peer.ID, ref peer.ID) []peer.ID {
	return sortList(ns, peers, ref)
}
