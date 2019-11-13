package networksharding

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	maxMaskBits  = 8
	fullMaskBits = 0xff
)

// kadSharder KAD based sharder
//
// Resets a number of MSb to decrease the distance between nodes from the same shard
type kadSharder struct {
	prioBits uint32
	resolver p2p.PeerShardResolver
}

// NewKadSharder kadSharder constructor
// prioBits - Number of reseted bits.
// f - Callback used to get the shard id for a given peer.ID
func NewKadSharder(prioBits uint32, kgs p2p.PeerShardResolver) (Sharder, error) {
	if prioBits == 0 || kgs == nil || kgs.IsInterfaceNil() {
		return nil, ErrBadParams
	}
	k := &kadSharder{
		prioBits: 8,
		resolver: kgs,
	}

	if prioBits < maxMaskBits {
		k.prioBits = prioBits
	}
	return k, nil
}

// GetShard get the shard id of the peer
func (ks *kadSharder) GetShard(id peer.ID) uint32 {
	return ks.resolver.ByID(p2p.PeerID(id))
}

// Resets distance bits
func (ks *kadSharder) resetDistanceBits(d []byte) []byte {
	if ks.prioBits == 0 {
		return d
	}
	mask := byte(((1 << (maxMaskBits - ks.prioBits)) - 1) & fullMaskBits)
	b0 := d[0] & mask
	return append([]byte{b0}[:], d[1:]...)
}

// GetDistance get the distance between a and b
func (ks *kadSharder) GetDistance(a, b sortingID) *big.Int {
	c := make([]byte, len(a.key))
	for i := 0; i < len(a.key); i++ {
		c[i] = a.key[i] ^ b.key[i]
	}

	if a.shard == b.shard {
		c = ks.resetDistanceBits(c)
	}

	ret := big.NewInt(0).SetBytes(c)
	return ret
}

// SortList sort the provided peers list
func (ks *kadSharder) SortList(peers []peer.ID, ref peer.ID) []peer.ID {
	return sortList(ks, peers, ref)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ks *kadSharder) IsInterfaceNil() bool {
	return ks == nil
}
