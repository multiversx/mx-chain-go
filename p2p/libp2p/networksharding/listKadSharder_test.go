package networksharding

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
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
				return core.UnknownShardId
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

//------- ComputeEvictionList

func TestListKadSharder_ComputeEvictionListNotReachedIntraShardShouldRetEmpty(t *testing.T) {
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
	pids := []peer.ID{pidCrtShard, pidCrossShard}

	evictList := lks.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListKadSharder_ComputeEvictionListNotReachedCrossShardShouldRetEmpty(t *testing.T) {
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
	pids := []peer.ID{pidCrtShard, pidCrossShard}

	evictList := lks.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListKadSharder_ComputeEvictionListReachedIntraShardShouldSortAndEvict(t *testing.T) {
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
	pids := []peer.ID{pidCrtShard2, pidCrtShard1}

	evictList := lks.ComputeEvictionList(pids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, pidCrtShard1, evictList[0])
}

func TestListKadSharder_ComputeEvictionListUnknownPeersShouldFillTheGap(t *testing.T) {
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
	unknownPids = append(unknownPids, newUnknownPid)

	evictList := lks.ComputeEvictionList(unknownPids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, unknownPids[0], evictList[0])
}

//------- Has

func TestListKadSharder_HasNotFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lks := &listKadSharder{}

	assert.False(t, lks.Has("pid4", list))
}

func TestListKadSharder_HasEmpty(t *testing.T) {
	t.Parallel()

	list := make([]peer.ID, 0)
	lks := &listKadSharder{}

	assert.False(t, lks.Has("pid4", list))
}

func TestListKadSharder_HasFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lks := &listKadSharder{}

	assert.True(t, lks.Has("pid2", list))
}

//------- computeDistance

func TestComputeDistanceByCountingBits(t *testing.T) {
	t.Parallel()

	//compute will be done on hashes. Impossible to predict the outcome in this test
	assert.Equal(t, uint64(0), computeDistanceByCountingBits("", "").Uint64())
	assert.Equal(t, uint64(0), computeDistanceByCountingBits("a", "a").Uint64())
	assert.Equal(t, uint64(139), computeDistanceByCountingBits(peer.ID([]byte{0}), peer.ID([]byte{1})).Uint64())
	assert.Equal(t, uint64(130), computeDistanceByCountingBits(peer.ID([]byte{0}), peer.ID([]byte{255})).Uint64())
	assert.Equal(t, uint64(117), computeDistanceByCountingBits(peer.ID([]byte{0, 128}), peer.ID([]byte{255, 255})).Uint64())
}

func TestComputeDistanceLog2Based(t *testing.T) {
	t.Parallel()

	//compute will be done on hashes. Impossible to predict the outcome in this test
	assert.Equal(t, uint64(0), computeDistanceLog2Based("", "").Uint64())
	assert.Equal(t, uint64(0), computeDistanceLog2Based("a", "a").Uint64())
	assert.Equal(t, uint64(254), computeDistanceLog2Based(peer.ID([]byte{0}), peer.ID([]byte{1})).Uint64())
	assert.Equal(t, uint64(250), computeDistanceLog2Based(peer.ID([]byte{254}), peer.ID([]byte{255})).Uint64())
	assert.Equal(t, uint64(256), computeDistanceLog2Based(peer.ID([]byte{0, 128}), peer.ID([]byte{255, 255})).Uint64())
}

func TestListKadSharder_SetPeerShardResolverNilShouldErr(t *testing.T) {
	t.Parallel()

	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)

	err := lks.SetPeerShardResolver(nil)

	assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
}

func TestListKadSharder_SetPeerShardResolverShouldWork(t *testing.T) {
	t.Parallel()

	lks, _ := NewListKadSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeers,
		minAllowedPeersOnList,
		minAllowedPeersOnList,
	)
	newPeerShardResolver := &mock.PeerShardResolverStub{}
	err := lks.SetPeerShardResolver(newPeerShardResolver)

	//pointer testing
	assert.True(t, lks.peerShardResolver == newPeerShardResolver)
	assert.Nil(t, err)
}
