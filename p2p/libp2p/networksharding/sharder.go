package networksharding

import (
	"crypto/sha256"
	"math/big"
	"sort"

	"github.com/libp2p/go-libp2p-core/peer"
)

// Sharder - Main sharder interface
type Sharder interface {
	// GetShard get the shard id of the peer
	GetShard(id peer.ID) uint32
	// GetDistance get the distance between a and b
	GetDistance(a, b sortingID) *big.Int
	// SortList sort the provided peers list
	SortList(peers []peer.ID, ref peer.ID) []peer.ID
	IsInterfaceNil() bool
}

func keyFromID(id peer.ID) []byte {
	key := sha256.Sum256([]byte(id))
	return key[:]
}

func sortList(s Sharder, peers []peer.ID, ref peer.ID) []peer.ID {
	sl := sortingList{
		ref: sortingID{
			id:       ref,
			key:      keyFromID(ref),
			shard:    s.GetShard(ref),
			distance: big.NewInt(0),
		},
		peers: make([]sortingID, len(peers)),
	}

	for i, id := range peers {
		sl.peers[i] = sortingID{
			id:       id,
			key:      keyFromID(id),
			shard:    s.GetShard(id),
			distance: big.NewInt(0),
		}
		sl.peers[i].distance = s.GetDistance(sl.peers[i], sl.ref)
	}

	sort.Sort(&sl)

	ret := make([]peer.ID, len(peers))

	for i, id := range sl.peers {
		ret[i] = id.id
	}

	return ret
}
