package networksharding

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"math/big"
)

// NoSharder default sharder, only uses Kademlia distance in sorting
type NoSharder struct {
}

// GetShard always 0
func (ns *NoSharder) GetShard(id peer.ID) uint32 {
	return 0
}

// GetDistance Kademlia XOR distance
func (ns *NoSharder) GetDistance(a, b sortingID) *big.Int {
	c := make([]byte, len(a.key))
	for i := 0; i < len(a.key); i++ {
		c[i] = a.key[i] ^ b.key[i]
	}

	ret := big.NewInt(0).SetBytes(c)
	return ret
}

// SortList do sort
func (ns *NoSharder) SortList(peers []peer.ID, ref peer.ID) []peer.ID {
	return sortList(ns, peers, ref)
}
