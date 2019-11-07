package networksharding

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	sha256 "github.com/minio/sha256-simd"
	"math/big"
	"sort"
	"sync"
)

var (
	currentSharder Sharder = &NoSharder{}
	onceSetter     sync.Once
)

// Get the sharder
func Get() Sharder {
	return currentSharder
}

// Set the sharder, can only be done once
func Set(s Sharder) error {
	if s == nil {
		return fmt.Errorf("The sharder cannot be nil")
	}

	ret := fmt.Errorf("Already set")
	onceSetter.Do(func() {
		currentSharder = s
		ret = nil
	})
	return ret
}

// Sharder - Main sharder interface
type Sharder interface {
	// Get the shard id of the peer
	GetShard(id peer.ID) uint32
	// Get the distance between a and b
	GetDistance(a, b sortingID) *big.Int

	// Sort a list of peers
	SortList(peers []peer.ID, ref peer.ID) []peer.ID
}

type sortingID struct {
	id       peer.ID
	key      []byte
	shard    uint32
	distance *big.Int
}

type sortingList struct {
	ref   sortingID
	peers []sortingID
}

// Len is the number of elements in the collection.
func (sl *sortingList) Len() int {
	return len(sl.peers)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (sl *sortingList) Less(i int, j int) bool {
	return sl.peers[i].distance.Cmp(sl.peers[j].distance) < 0
}

// Swap swaps the elements with indexes i and j.
func (sl *sortingList) Swap(i int, j int) {
	sl.peers[i], sl.peers[j] = sl.peers[j], sl.peers[i]
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
