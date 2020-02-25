package networksharding

import (
	"crypto/sha256"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/sorting"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Sharder - Main sharder interface
type Sharder interface {
	// GetShard get the shard id of the peer
	GetShard(id peer.ID) uint32
	// GetDistance get the distance between a and b
	GetDistance(a, b sorting.SortedID) *big.Int
	// SortList sort the provided peers list
	SortList(peers []peer.ID, ref peer.ID) ([]peer.ID, bool)
	IsInterfaceNil() bool
}

func keyFromID(id peer.ID) []byte {
	key := sha256.Sum256([]byte(id))
	return key[:]
}

func getSortingList(s Sharder, peers []peer.ID, ref peer.ID) *sorting.SortedList {
	sl := sorting.SortedList{
		Ref: sorting.SortedID{
			ID:       ref,
			Key:      keyFromID(ref),
			Shard:    s.GetShard(ref),
			Distance: big.NewInt(0),
		},
		Peers: make([]sorting.SortedID, len(peers)),
	}

	for i, id := range peers {
		sl.Peers[i] = sorting.SortedID{
			ID:       id,
			Key:      keyFromID(id),
			Shard:    s.GetShard(id),
			Distance: big.NewInt(0),
		}
		sl.Peers[i].Distance = s.GetDistance(sl.Peers[i], sl.Ref)
	}
	return &sl
}

func sortList(s Sharder, peers []peer.ID, ref peer.ID) []peer.ID {
	sl := getSortingList(s, peers, ref)
	return sl.SortedPeers()
}
