package networksharding

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"math/big"
)

// KadGetShardID Callback returning the shard id of the given peer.ID
type KadGetShardID func(peer.ID) uint32

// KadSharder KAD based sharder
//
// Resets a number of MSb in the distance to nods sharding the same shard
type KadSharder struct {
	cutOffMask byte
	gsFunc     KadGetShardID
}

// GetShard Check networksharding.Sharder
func (ks *KadSharder) GetShard(id peer.ID) uint32 {
	return ks.gsFunc(id)
}

// Resets distance bits
func (ks *KadSharder) resetDistanceBits(d []byte) []byte {
	return append([]byte{d[0] &^ ks.cutOffMask}[:], d[1:]...)
}

// GetDistance Check networksharding.Sharder
func (ks *KadSharder) GetDistance(a, b sortingID) *big.Int {
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

// SortList Check networksharding.Sharder
func (ks *KadSharder) SortList(peers []peer.ID, ref peer.ID) []peer.ID {
	return sortList(ks, peers, ref)
}

// NewKadSharder KadSharder constructor
// prioBits - Number of reseted bits.
// f - Callback used to get the shard id for a given peer.ID
func NewKadSharder(prioBits uint32, f KadGetShardID) (Sharder, error) {
	if prioBits == 0 || f == nil {
		return nil, fmt.Errorf("Bad params")
	}
	k := &KadSharder{
		cutOffMask: 0,
		gsFunc:     f,
	}

	if prioBits >= 8 {
		k.cutOffMask = 0xff
	} else {
		t := byte(0)
		for i := uint32(0); i < prioBits; i++ {
			t |= 1 << (7 - i)
		}
		k.cutOffMask = t
	}
	return k, nil
}
