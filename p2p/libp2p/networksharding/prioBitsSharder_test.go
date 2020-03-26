package networksharding

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

const (
	testNodesCount = 1000
)

func fakeShard0(_ p2p.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{
		ShardID: 0,
	}
}

func fakeShardBit0Byte0(id p2p.PeerID) core.P2PPeerInfo {
	ret := sha256.Sum256([]byte(id))

	return core.P2PPeerInfo{
		ShardID: uint32(ret[0] & 1),
	}
}

type testKadResolver struct {
	f func(p2p.PeerID) core.P2PPeerInfo
}

func (tkr *testKadResolver) GetPeerInfo(pid p2p.PeerID) core.P2PPeerInfo {
	return tkr.f(pid)
}

func (tkr *testKadResolver) NumShards() uint32 {
	return 3
}

func (tkr *testKadResolver) IsInterfaceNil() bool {
	return tkr == nil
}

var (
	fs0  = &testKadResolver{fakeShard0}
	fs20 = &testKadResolver{fakeShardBit0Byte0}
)

func TestNewPrioBitsSharder_ZeroPrioBitsShouldErr(t *testing.T) {
	t.Parallel()

	pbs, err := NewPrioBitsSharder(0, &mock.PeerShardResolverStub{})

	assert.True(t, check.IfNil(pbs))
	assert.True(t, errors.Is(err, ErrBadParams))
}

func TestNewPrioBitsSharder_NilPeerShardResolverShouldErr(t *testing.T) {
	t.Parallel()

	pbs, err := NewPrioBitsSharder(1, nil)

	assert.True(t, check.IfNil(pbs))
	assert.True(t, errors.Is(err, p2p.ErrNilPeerShardResolver))
}

func TestNewPrioBitsSharder_ShouldWork(t *testing.T) {
	t.Parallel()

	pbs, err := NewPrioBitsSharder(1, &mock.PeerShardResolverStub{})

	assert.False(t, check.IfNil(pbs))
	assert.Nil(t, err)
}

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
		tdCopy := td
		t.Run(fmt.Sprint(tdCopy.l, "_", tdCopy.exp), func(t *testing.T) {
			pbs, _ := NewPrioBitsSharder(tdCopy.l, fs0)
			r := pbs.resetDistanceBits(i)
			assert.Equal(t, big.NewInt(0).SetBytes(r), tdCopy.exp, "Should match")
		})
	}
}

func TestPrioBitsSharderDistance(t *testing.T) {
	pbs, _ := NewPrioBitsSharder(8, fs0)
	checkDistance(pbs, t)
}

func TestKadSharderOrdering2(t *testing.T) {
	pbs, _ := NewPrioBitsSharder(2, fs20)
	checkOrdering(pbs, t)
}

func TestPrioBitsSharderOrdering2_list(t *testing.T) {
	s, _ := NewPrioBitsSharder(4, fs20)

	peerList := make([]peer.ID, testNodesCount)
	for i := 0; i < testNodesCount; i++ {
		peerList[i] = peer.ID(fmt.Sprintf("NODE %d", i))
	}
	l1, _ := s.SortList(peerList, nodeA)

	refPeerInfo := fakeShardBit0Byte0(p2p.PeerID(nodeA))
	sameShardScore := uint64(0)
	sameShardCount := uint64(0)
	otherShardScore := uint64(0)
	retLen := uint64(len(l1))
	fmt.Printf("[ref] %s , sha %x, shard %d\n", string(nodeA), sha256.Sum256([]byte(nodeA)), refPeerInfo.ShardID)
	for i, id := range l1 {
		peerInfo := fakeShardBit0Byte0(p2p.PeerID(id))

		if peerInfo.ShardID == refPeerInfo.ShardID {
			sameShardScore += retLen - uint64(i)
			sameShardCount++
		} else {
			otherShardScore += retLen - uint64(i)
		}
	}

	avgSame := sameShardScore / sameShardCount
	avgOther := otherShardScore / (retLen - sameShardCount)
	fmt.Printf("Same shard avg score %d, Other shard avg score %d\n", avgSame, avgOther)

	assert.True(t, avgSame > avgOther)
}

func TestPrioBitsSharder_SetPeerShardResolverNilShouldErr(t *testing.T) {
	t.Parallel()

	pbs, _ := NewPrioBitsSharder(1, &mock.PeerShardResolverStub{})

	err := pbs.SetPeerShardResolver(nil)

	assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
}

func TestKadSharder_SetPeerShardResolverShouldWork(t *testing.T) {
	t.Parallel()

	pbs, _ := NewPrioBitsSharder(1, &mock.PeerShardResolverStub{})
	newPeerShardResolver := &mock.PeerShardResolverStub{}
	err := pbs.SetPeerShardResolver(newPeerShardResolver)

	//pointer testing
	assert.True(t, pbs.resolver == newPeerShardResolver)
	assert.Nil(t, err)
}
