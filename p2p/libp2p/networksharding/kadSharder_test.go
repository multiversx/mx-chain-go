package networksharding

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
	sha256 "github.com/minio/sha256-simd"
	"github.com/stretchr/testify/assert"
)

const (
	testNodesCount = 1000
)

func fakeShard0(id p2p.PeerID) uint32 {
	return 0
}

func fakeShardBit0Byte2(id p2p.PeerID) uint32 {
	ret := sha256.Sum256([]byte(id))
	return uint32(ret[2] & 1)
}

type testKadResolver struct {
	f func(p2p.PeerID) uint32
}

func (tkr *testKadResolver) ByID(peer p2p.PeerID) uint32 {
	return tkr.f(peer)
}

func (tkr *testKadResolver) IsBalanced() bool {
	return true
}

func (tkr *testKadResolver) IsInterfaceNil() bool {
	return tkr == nil
}

var (
	fs0   = &testKadResolver{fakeShard0}
	fs2_0 = &testKadResolver{fakeShardBit0Byte2}
)

func TestCutoOffBits(t *testing.T) {

	i := []byte{0xff, 0xff}[:]

	testData := []struct {
		l   uint32
		exp *big.Int
	}{
		{
			l:   1,
			exp: big.NewInt(0x7f<<8 | 0xff),
		},
		{
			l:   2,
			exp: big.NewInt(0x3f<<8 | 0xff),
		},
		{
			l:   3,
			exp: big.NewInt(0x1f<<8 | 0xff),
		},
		{
			l:   7,
			exp: big.NewInt(0x1<<8 | 0xff),
		},
		{
			l:   8,
			exp: big.NewInt(0xff),
		},

		{
			l:   9,
			exp: big.NewInt(0xff),
		},
	}

	for _, td := range testData {
		t.Run(fmt.Sprint(td.l, "_", td.exp), func(t *testing.T) {
			s, _ := NewKadSharder(td.l, fs0)
			k := s.(*kadSharder)
			r := k.resetDistanceBits(i)
			assert.Equal(t, big.NewInt(0).SetBytes(r), td.exp, "Should match")
		})

	}
}

func TestKadSharderDistance(t *testing.T) {
	s, _ := NewKadSharder(8, fs0)
	checkDistance(s, t)
}

func TestKadSharderOrdering2(t *testing.T) {
	s, _ := NewKadSharder(2, fs2_0)
	checkOrdering(s, t)
}

func TestKadSharderOrdering2_list(t *testing.T) {
	s, _ := NewKadSharder(4, fs2_0)

	peerList := make([]peer.ID, testNodesCount)
	for i := 0; i < testNodesCount; i++ {
		peerList[i] = peer.ID(fmt.Sprintf("NODE %d", i))
	}
	l1, _ := s.SortList(peerList, nodeA)

	refShardID := fakeShardBit0Byte2(p2p.PeerID(nodeA))
	sameShardScore := uint64(0)
	sameShardCount := uint64(0)
	otherShardScore := uint64(0)
	retLen := uint64(len(l1))
	t.Logf("[ref] %s , sha %x, shard %d\n", string(nodeA), sha256.Sum256([]byte(nodeA)), refShardID)
	for i, id := range l1 {
		shardID := fakeShardBit0Byte2(p2p.PeerID(id))

		if shardID == refShardID {
			sameShardScore += retLen - uint64(i)
			sameShardCount++
		} else {
			otherShardScore += retLen - uint64(i)
		}
	}

	avgSame := sameShardScore / sameShardCount
	avgOther := otherShardScore / (retLen - sameShardCount)
	t.Logf("Same shard avg score %d, Other shard avg score %d\n", avgSame, avgOther)

	assert.True(t, avgSame > avgOther)
}
