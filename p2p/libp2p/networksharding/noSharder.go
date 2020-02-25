package networksharding

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/sorting"
	"github.com/libp2p/go-libp2p-core/peer"
)

// NoSharder default sharder, only uses Kademlia distance in sorting
type NoSharder struct {
}

// GetShard always 0
func (ns *NoSharder) GetShard(_ peer.ID) uint32 {
	return 0
}

// GetDistance Kademlia XOR distance
func (ns *NoSharder) GetDistance(a, b sorting.SortingID) *big.Int {
	c := make([]byte, len(a.Key))
	for i := 0; i < len(a.Key); i++ {
		c[i] = a.Key[i] ^ b.Key[i]
	}

	ret := big.NewInt(0).SetBytes(c)
	return ret
}

// SortList sort the list
func (ns *NoSharder) SortList(peers []peer.ID, ref peer.ID) ([]peer.ID, bool) {
	return sortList(ns, peers, ref), true
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NoSharder) IsInterfaceNil() bool {
	return ns == nil
}
