package networksharding

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

const crtShardId = uint32(0)

var crtPid = peer.ID(fmt.Sprintf("%d pid", crtShardId))

func createStringPeersShardResolver() *mock.PeerShardResolverStub {
	return &mock.PeerShardResolverStub{
		ByIDCalled: func(pid p2p.PeerID) uint32 {
			strPid := string(pid)
			if strings.Contains(strPid, fmt.Sprintf("%d", crtShardId)) {
				return crtShardId
			}
			if strings.Contains(strPid, "u") {
				return sharding.UnknownShardId
			}

			return crtShardId + 1
		},
	}
}

func TestNewListKadSharder_NilPeerShardResolverShouldErr(t *testing.T) {
	t.Parallel()

	lks, err := NewListKadSharder(
		nil,
		"",
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)

	assert.True(t, check.IfNil(lks))
	assert.True(t, errors.Is(err, p2p.ErrNilPeerShardResolver))
}

func TestNewListKadSharder_InvalidMaxPeerCountShouldErr(t *testing.T) {
	t.Parallel()

	lks, err := NewListKadSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeers-1,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)

	assert.True(t, check.IfNil(lks))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListKadSharder_InvalidMaxIntraShardShouldErr(t *testing.T) {
	t.Parallel()

	lks, err := NewListKadSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeers,
		minAllowedPeersOnList-1,
		minAllowedPeersOnList,
	)

	assert.True(t, check.IfNil(lks))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListKadSharder_InvalidMaxCrossShardShouldErr(t *testing.T) {
	t.Parallel()

	lks, err := NewListKadSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList-1,
	)

	assert.True(t, check.IfNil(lks))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListKadSharder_ShouldWork(t *testing.T) {
	t.Parallel()

	lks, err := NewListKadSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)

	assert.False(t, check.IfNil(lks))
	assert.Nil(t, err)
}

//------- ComputeEvictList

func TestListKadSharder_ComputeEvictListAlreadyContainsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)
	pid := peer.ID("pid")
	connected := []peer.ID{pid}

	evictList := lks.ComputeEvictList(pid, connected)

	assert.Equal(t, 0, len(evictList))
}

func TestListKadSharder_ComputeEvictListNotReachedIntraShardShouldRetEmpty(t *testing.T) {
	t.Parallel()

	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)
	pidCrtShard := peer.ID(fmt.Sprintf("%d new pid", crtShardId))
	pidCrossShard := peer.ID(fmt.Sprintf("%d cross", crtShardId+1))
	connected := []peer.ID{pidCrossShard}

	evictList := lks.ComputeEvictList(pidCrtShard, connected)

	assert.Equal(t, 0, len(evictList))
}

func TestListKadSharder_ComputeEvictListNotReachedCrossShardShouldRetEmpty(t *testing.T) {
	t.Parallel()

	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)
	pidCrtShard := peer.ID(fmt.Sprintf("%d new pid", crtShardId))
	pidCrossShard := peer.ID(fmt.Sprintf("%d cross", crtShardId+1))
	connected := []peer.ID{pidCrtShard}

	evictList := lks.ComputeEvictList(pidCrossShard, connected)

	assert.Equal(t, 0, len(evictList))
}

func TestListKadSharder_ComputeEvictListReachedIntraShardShouldSortAndEvict(t *testing.T) {
	t.Parallel()

	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)
	pidCrtShard1 := peer.ID(fmt.Sprintf("%d - 1 - new pid", crtShardId))
	pidCrtShard2 := peer.ID(fmt.Sprintf("%d - 2 - new pid", crtShardId))
	connected := []peer.ID{pidCrtShard2}

	evictList := lks.ComputeEvictList(pidCrtShard1, connected)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, pidCrtShard2, evictList[0])
}

func TestListKadSharder_ComputeEvictListUnknownPeersShouldFillTheGap(t *testing.T) {
	t.Parallel()

	maxPeerCount := 4
	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		maxPeerCount,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)

	unknownPids := make([]peer.ID, maxPeerCount+1)
	for i := 0; i < maxPeerCount+1; i++ {
		unknownPids[i] = "u b pid"
	}
	newUnknownPid := peer.ID("u a pid")

	evictList := lks.ComputeEvictList(newUnknownPid, unknownPids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, unknownPids[0], evictList[0])
}

//------- computeDistance

func TestComputeDistance(t *testing.T) {
	t.Parallel()

	assert.Equal(t, uint64(0), computeDistance("", "").Uint64())
	assert.Equal(t, uint64(0), computeDistance("a", "").Uint64())
	assert.Equal(t, uint64(0), computeDistance("a", "a").Uint64())
	assert.Equal(t, uint64(1), computeDistance(peer.ID([]byte{0}), peer.ID([]byte{1})).Uint64())
	assert.Equal(t, uint64(255), computeDistance(peer.ID([]byte{0}), peer.ID([]byte{255})).Uint64())
	expectedResult := big.NewInt(0).SetBytes([]byte{255, 127})
	assert.Equal(t, expectedResult.Uint64(), computeDistance(peer.ID([]byte{0, 128}), peer.ID([]byte{255, 255})).Uint64())
}
