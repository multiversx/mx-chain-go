package networksharding

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/sorting"
	"github.com/libp2p/go-libp2p-core/peer"
)

// SimplePrioBitsSharder only uses Kademlia distance in sorting
type SimplePrioBitsSharder struct {
}

// GetShard always 0
func (spbs *SimplePrioBitsSharder) GetShard(_ peer.ID) uint32 {
	return 0
}

// GetDistance Kademlia XOR distance
func (spbs *SimplePrioBitsSharder) GetDistance(a, b sorting.SortingID) *big.Int {
	c := make([]byte, len(a.Key))
	for i := 0; i < len(a.Key); i++ {
		c[i] = a.Key[i] ^ b.Key[i]
	}

	ret := big.NewInt(0).SetBytes(c)
	return ret
}

// SortList sort the list
func (spbs *SimplePrioBitsSharder) SortList(peers []peer.ID, ref peer.ID) ([]peer.ID, bool) {
	return sortList(spbs, peers, ref), true
}

// SetPeerShardResolver will do nothing
func (spbs *SimplePrioBitsSharder) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (spbs *SimplePrioBitsSharder) IsInterfaceNil() bool {
	return spbs == nil
}
