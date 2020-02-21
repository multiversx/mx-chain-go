package networksharding

import (
	"fmt"
	"math/big"
	"math/bits"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/libp2p/go-libp2p-core/peer"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

const minAllowedConnectedPeers = 2
const minAllowedPeersOnList = 1

var leadingZerosCount = []int{
	8, 7, 6, 6, 5, 5, 5, 5,
	4, 4, 4, 4, 4, 4, 4, 4,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// this will fail if we have less than 256 values in the slice
var _ = leadingZerosCount[255]

// listKadSharder is the struct able to compute an eviction list of connected peers id according to the
// provided parameters. It basically splits all connected peers into 3 lists: intra shard peers, cross shard peers
// and unknown peers by the following rule: both intra shard and cross shard lists are upper bounded to provided
// maximum levels, unknown list is able to fill the gap until maximum peer count value is fulfilled.
type listKadSharder struct {
	peerShardResolver p2p.PeerShardResolver
	selfPeerId        peer.ID
	maxPeerCount      int
	maxIntraShard     int
	maxCrossShard     int
	computeDistance   func(src peer.ID, dest peer.ID) *big.Int
}

// NewListKadSharder creates a new kad list based kad sharder instance
func NewListKadSharder(
	resolver p2p.PeerShardResolver,
	selfPeerId peer.ID,
	maxPeerCount int,
	maxIntraShard int,
	maxCrossShard int,
) (*listKadSharder, error) {
	if check.IfNil(resolver) {
		return nil, p2p.ErrNilPeerShardResolver
	}
	if maxPeerCount < minAllowedConnectedPeers {
		return nil, fmt.Errorf("%w, maxPeerCount should be at least %d", p2p.ErrInvalidValue, minAllowedConnectedPeers)
	}
	if maxIntraShard < minAllowedPeersOnList {
		return nil, fmt.Errorf("%w, maxIntraShard should be at least %d", p2p.ErrInvalidValue, minAllowedPeersOnList)
	}
	if maxCrossShard < minAllowedPeersOnList {
		return nil, fmt.Errorf("%w, maxCrossShard should be at least %d", p2p.ErrInvalidValue, minAllowedPeersOnList)
	}

	return &listKadSharder{
		peerShardResolver: resolver,
		selfPeerId:        selfPeerId,
		maxPeerCount:      maxPeerCount,
		maxIntraShard:     maxIntraShard,
		maxCrossShard:     maxCrossShard,
		computeDistance:   computeDistanceByCountingBits,
	}, nil
}

// ComputeEvictionList returns the eviction list
func (lks *listKadSharder) ComputeEvictionList(pidList []peer.ID) []peer.ID {
	evictionProposed := make([]peer.ID, 0)
	intraShard, crossShard, unknownShard := lks.splitPeerIds(pidList)

	intraShard, e := lks.evict(intraShard, lks.maxIntraShard)
	evictionProposed = append(evictionProposed, e...)

	crossShard, e = lks.evict(crossShard, lks.maxCrossShard)
	evictionProposed = append(evictionProposed, e...)

	sum := len(intraShard) + len(crossShard) + len(unknownShard)
	if sum <= lks.maxPeerCount {
		return evictionProposed
	}
	remainingForUnknown := lks.maxPeerCount + 1 - len(intraShard) - len(crossShard)
	_, e = lks.evict(unknownShard, remainingForUnknown)

	return append(evictionProposed, e...)
}

// Has returns true if provided pid is among the provided list
func (lks *listKadSharder) Has(pid peer.ID, list []peer.ID) bool {
	for _, p := range list {
		if p == pid {
			return true
		}
	}

	return false
}

// PeerShardResolver returns the peer shard resolver used by this kad sharder
func (lks *listKadSharder) PeerShardResolver() p2p.PeerShardResolver {
	return lks.peerShardResolver
}

//TODO study if we need to hve a dedicated section for metanodes
func (lks *listKadSharder) splitPeerIds(peers []peer.ID) (peerDistances, peerDistances, peerDistances) {
	selfId := lks.peerShardResolver.ByID(p2p.PeerID(lks.selfPeerId))

	intraShard := peerDistances{}
	crossShard := peerDistances{}
	unknownShard := peerDistances{}

	for _, p := range peers {
		pd := peerDistance{
			ID:       p,
			distance: lks.computeDistance(p, lks.selfPeerId),
		}
		pid := p2p.PeerID(p)
		shardId := lks.peerShardResolver.ByID(pid)

		switch shardId {
		case sharding.UnknownShardId:
			unknownShard = append(unknownShard, pd)
		case selfId:
			intraShard = append(intraShard, pd)
		default:
			crossShard = append(crossShard, pd)
		}
	}

	return intraShard, crossShard, unknownShard
}

func (lks *listKadSharder) evict(distances peerDistances, numKeep int) (peerDistances, []peer.ID) {
	if numKeep < 0 {
		numKeep = 0
	}
	if numKeep >= len(distances) {
		return distances, make([]peer.ID, 0)
	}

	sort.Sort(distances)
	remaining := distances[:numKeep]
	evictedPD := distances[numKeep:]
	evictedPids := make([]peer.ID, len(evictedPD))
	for i, pd := range evictedPD {
		evictedPids[i] = pd.ID
	}

	return remaining, evictedPids
}

// IsInterfaceNil returns true if there is no value under the interface
func (lks *listKadSharder) IsInterfaceNil() bool {
	return lks == nil
}

// computes the kademlia distance between 2 provided peers by doing byte xor operations and counting the resulting bits
func computeDistanceByCountingBits(src peer.ID, dest peer.ID) *big.Int {
	srcBuff := kbucket.ConvertPeerID(src)
	destBuff := kbucket.ConvertPeerID(dest)

	cumulatedBits := 0
	for i := 0; i < len(srcBuff); i++ {
		result := srcBuff[i] ^ destBuff[i]
		cumulatedBits += bits.OnesCount8(result)
	}

	return big.NewInt(0).SetInt64(int64(cumulatedBits))
}

// computes the kademlia distance between 2 provided peers by doing byte xor operations and applying log2 on the result
func computeDistanceLog2Based(src peer.ID, dest peer.ID) *big.Int {
	srcBuff := kbucket.ConvertPeerID(src)
	destBuff := kbucket.ConvertPeerID(dest)

	val := 0
	for i := 0; i < len(srcBuff); i++ {
		result := srcBuff[i] ^ destBuff[i]
		val += leadingZerosCount[result]
		if result != 0 {
			break
		}
	}

	val = len(srcBuff)*8 - val

	return big.NewInt(0).SetInt64(int64(val))
}
