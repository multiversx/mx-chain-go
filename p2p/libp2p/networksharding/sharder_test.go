package networksharding

import (
	"math/big"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

var (
	nodeA = peer.ID("NODE A")
	nodeB = peer.ID("NODE B")
	nodeC = peer.ID("NODE C")
	nodeD = peer.ID("NODE D")
	nodeE = peer.ID("NODE E")
	nodeF = peer.ID("NODE F")
)

func makeSortingID(id peer.ID) sortingID {
	return sortingID{
		id:       id,
		key:      keyFromID(id),
		shard:    0,
		distance: big.NewInt(0),
	}
}

func checkDistance(s Sharder, t *testing.T) {
	sidA := makeSortingID(nodeA)
	r := s.GetDistance(sidA, sidA)
	assert.Equal(t, r.Cmp(big.NewInt(0)), 0, "Distance from a to a should be 0")

	sidB := makeSortingID(nodeB)
	rab := s.GetDistance(sidA, sidB)
	rba := s.GetDistance(sidB, sidA)
	assert.Equal(t, rab.Cmp(rba), 0, "Distance from a to b should be equal to b to a")
}

func checkOrdering(s Sharder, t *testing.T) {
	l1 := s.SortList([]peer.ID{nodeB, nodeC, nodeD, nodeE, nodeF}, nodeA)
	l2 := s.SortList([]peer.ID{nodeB, nodeE, nodeF, nodeD, nodeC}, nodeA)

	assert.Equal(t, len(l1), len(l2), "The two lists should have the seame size")
	assert.Equal(t, l1, l2, "The two lists should be the same")

	l3 := s.SortList(l1, nodeA)

	assert.Equal(t, len(l1), len(l3), "The two lists should have the seame size")
	assert.Equal(t, l1, l3, "The two lists should be the same")
}

func TestSetterGetter(t *testing.T) {

	_, isNoSharder := Get().(*noSharder)
	assert.True(t, isNoSharder, "Init should have * NoSharder type")

	k1, _ := NewKadSharder(0, fs0)
	k2, _ := NewKadSharder(1, fs0)
	k3, _ := NewKadSharder(2, fs0)

	err := Set(k1)
	assert.NotNil(t, err)
	_, isNoSharder = Get().(*noSharder)
	assert.True(t, isNoSharder, "Init should have * NoSharder type")

	err = Set(k2)
	assert.Nil(t, err)
	c, isKadSharder := Get().(*kadSharder)
	assert.True(t, isKadSharder, "Current sharder should be *KadSharder")
	assert.Equal(t, c, k2)

	err = Set(k3)
	assert.NotNil(t, err)
	c, isKadSharder = Get().(*kadSharder)
	assert.True(t, isKadSharder, "Current sharder should be *KadSharder")
	assert.Equal(t, c, k2)
}
