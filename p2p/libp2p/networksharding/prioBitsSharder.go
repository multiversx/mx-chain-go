package networksharding

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/sorting"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	maxMaskBits  = 8
	fullMaskBits = 0xff

	minInShardConnRatio = 0.65 // the minimum in shard vs total connections ratio
	minOOSHardLimit     = 3    // the hard limit for minimum out of shard connections count
)

// prioBitsSharder KAD based sharder using prio bits
// Resets a number of MSb to decrease the distance between nodes from the same shard
type prioBitsSharder struct {
	prioBits    uint32
	mutResolver sync.RWMutex
	resolver    p2p.PeerShardResolver
}

// NewPrioBitsSharder kadSharder constructor
// prioBits - Number of bits to reset
// psp - peer shard resolver used to get the shard id for a given peer.ID
func NewPrioBitsSharder(prioBits uint32, psp p2p.PeerShardResolver) (*prioBitsSharder, error) {
	if prioBits == 0 {
		return nil, fmt.Errorf("%w, prioBits should be greater than 0", ErrBadParams)
	}
	if check.IfNil(psp) {
		return nil, p2p.ErrNilPeerShardResolver
	}
	k := &prioBitsSharder{
		resolver: psp,
		prioBits: prioBits,
	}

	if prioBits > maxMaskBits {
		k.prioBits = maxMaskBits
	}
	return k, nil
}

// GetShard get the shard id of the peer
func (pbs *prioBitsSharder) GetShard(id peer.ID) uint32 {
	pbs.mutResolver.RLock()
	defer pbs.mutResolver.RUnlock()

	return pbs.resolver.ByID(p2p.PeerID(id))
}

// Resets distance bits
func (pbs *prioBitsSharder) resetDistanceBits(d []byte) []byte {
	if pbs.prioBits == 0 {
		return d
	}
	mask := byte(((1 << (maxMaskBits - pbs.prioBits)) - 1) & fullMaskBits)
	b0 := d[0] & mask
	return append([]byte{b0}[:], d[1:]...)
}

// GetDistance get the distance between a and b
func (pbs *prioBitsSharder) GetDistance(a, b sorting.SortedID) *big.Int {
	c := make([]byte, len(a.Key))
	for i := 0; i < len(a.Key); i++ {
		c[i] = a.Key[i] ^ b.Key[i]
	}

	if a.Shard == b.Shard {
		c = pbs.resetDistanceBits(c)
	}

	ret := big.NewInt(0).SetBytes(c)
	return ret
}

// SortList sort the provided peers list
func (pbs *prioBitsSharder) SortList(peers []peer.ID, ref peer.ID) ([]peer.ID, bool) {
	sl := getSortingList(pbs, peers, ref)
	// for balance we should have between 1 and 20% connections outside of shard
	peerCnt := len(peers)
	inShardCnt := inShardCount(sl)
	balanced := getMinOOS(pbs.prioBits, peerCnt) <= (peerCnt - inShardCnt)

	if balanced {
		minInShard := int(math.Floor(float64(peerCnt) * minInShardConnRatio))
		balanced = inShardCnt >= minInShard
	}

	return sl.SortedPeers(), balanced
}

func inShardCount(sl *sorting.SortedList) int {
	cnt := 0
	for _, p := range sl.Peers {
		if p.Shard == sl.Ref.Shard {
			cnt++
		}
	}
	return cnt
}

func getMinOOS(bits uint32, conns int) int {
	t := conns / (1 << (bits + 1))
	if t < minOOSHardLimit {
		return minOOSHardLimit
	}
	return t
}

// SetPeerShardResolver sets the peer shard resolver for this sharder
func (pbs *prioBitsSharder) SetPeerShardResolver(psp p2p.PeerShardResolver) error {
	if check.IfNil(psp) {
		return p2p.ErrNilPeerShardResolver
	}

	pbs.mutResolver.Lock()
	pbs.resolver = psp
	pbs.mutResolver.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pbs *prioBitsSharder) IsInterfaceNil() bool {
	return pbs == nil
}
