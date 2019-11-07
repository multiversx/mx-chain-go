package networksharding

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	sha256 "github.com/minio/sha256-simd"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestCutoOffBits(t *testing.T) {

	i := []byte{0xff, 0xff}[:]

	s, _ := NewKadSharder(1, fakeShard0)
	k := s.(*KadSharder)
	r := k.resetDistanceBits(i)
	assert.Equal(t, big.NewInt(0).SetBytes(r), big.NewInt(0x7f<<8|0xff), "Shouls match")

	s, _ = NewKadSharder(2, fakeShard0)
	k = s.(*KadSharder)
	r = k.resetDistanceBits(i)
	assert.Equal(t, big.NewInt(0).SetBytes(r), big.NewInt(0x3f<<8|0xff), "Shouls match")

	s, _ = NewKadSharder(3, fakeShard0)
	k = s.(*KadSharder)
	r = k.resetDistanceBits(i)
	assert.Equal(t, big.NewInt(0).SetBytes(r), big.NewInt(0x1f<<8|0xff), "Shouls match")

	s, _ = NewKadSharder(7, fakeShard0)
	k = s.(*KadSharder)
	r = k.resetDistanceBits(i)
	assert.Equal(t, big.NewInt(0).SetBytes(r), big.NewInt(0x1<<8|0xff), "Shouls match")

	s, _ = NewKadSharder(8, fakeShard0)
	k = s.(*KadSharder)
	r = k.resetDistanceBits(i)
	assert.Equal(t, big.NewInt(0).SetBytes(r), big.NewInt(0xff), "Shouls match")

	s, _ = NewKadSharder(9, fakeShard0)
	k = s.(*KadSharder)
	r = k.resetDistanceBits(i)
	assert.Equal(t, big.NewInt(0).SetBytes(r), big.NewInt(0xff), "Shouls match")
}

func fakeShard0(id peer.ID) uint32 {
	return 0
}

func fakeShardB2_1b(id peer.ID) uint32 {
	ret := sha256.Sum256([]byte(id))

	return uint32(ret[2] & 1)
}

func TestKadSharderDistance(t *testing.T) {
	s, _ := NewKadSharder(8, fakeShard0)
	checkDistance(s, t)
}

func TestKadSharderOrdering2(t *testing.T) {
	s, _ := NewKadSharder(2, fakeShardB2_1b)
	checkOrdering(s, t)
}

const (
	testNodesCount = 1000
)

func TestKadSharderOrdering2_list(t *testing.T) {
	s, _ := NewKadSharder(4, fakeShardB2_1b)

	peerList := make([]peer.ID, testNodesCount)
	for i := 0; i < testNodesCount; i++ {
		peerList[i] = peer.ID(fmt.Sprintf("NODE %d", i))
	}
	l1 := s.SortList(peerList, nodeA)

	refShardID := fakeShardB2_1b(nodeA)
	sameShardScore := uint64(0)
	sameShardCount := uint64(0)
	otherShardScore := uint64(0)
	retLen := uint64(len(l1))
	t.Logf("[ref] %s , sha %x, shard %d\n", string(nodeA), sha256.Sum256([]byte(nodeA)), refShardID)
	for i, id := range l1 {
		// keep this
		//sid := makeSortingID(id)
		shardID := fakeShardB2_1b(id)

		if shardID == refShardID {
			sameShardScore += retLen - uint64(i)
			sameShardCount++
		} else {
			otherShardScore += retLen - uint64(i)
		}
		// keep this
		// t.Logf("[%d] %s , sha %x, shard %d\n", i, string(id), sid.key, fakeShardB2_1b(id))
	}

	avgSame := sameShardScore / sameShardCount
	avgOther := otherShardScore / (retLen - sameShardCount)
	t.Logf("Same shard avg score %d, Other shard avg score %d\n", avgSame, avgOther)

	assert.True(t, avgSame > avgOther)
}
