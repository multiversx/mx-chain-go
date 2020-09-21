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
const crossShardId = uint32(1)

const validatorMarker = "validator"
const observerMarker = "observer"
const unknownMarker = "unknown"

var crtPid = peer.ID(fmt.Sprintf("%d pid", crtShardId))

func createStringPeersShardResolver() *mock.PeerShardResolverStub {
	return &mock.PeerShardResolverStub{
		GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
			strPid := string(pid)
			pInfo := core.P2PPeerInfo{}

			if strings.Contains(strPid, fmt.Sprintf("%d", crtShardId)) {
				pInfo.ShardID = crtShardId
			} else {
				pInfo.ShardID = crossShardId
			}

			if strings.Contains(strPid, unknownMarker) {
				pInfo.PeerType = core.UnknownPeer
			}
			if strings.Contains(strPid, validatorMarker) {
				pInfo.PeerType = core.ValidatorPeer
			}
			if strings.Contains(strPid, observerMarker) {
				pInfo.PeerType = core.ObserverPeer
			}

			return pInfo
		},
	}
}

func countPeers(peers []peer.ID, shardID uint32, marker string) int {
	counter := 0
	for _, pid := range peers {
		if strings.Contains(string(pid), marker) &&
			strings.Contains(string(pid), fmt.Sprintf("%d", shardID)) {
			counter++
		}
	}

	return counter
}

func TestNewListsSharder_NilPeerShardResolverShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		nil,
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	assert.True(t, check.IfNil(ls))
	assert.True(t, errors.Is(err, p2p.ErrNilPeerShardResolver))
}

func TestNewListsSharder_InvalidIntraShardValidatorsShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators-1,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	assert.True(t, check.IfNil(ls))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListsSharder_InvalidCrossShardValidatorsShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators-1,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	assert.True(t, check.IfNil(ls))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListsSharder_InvalidIntraShardObserversShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers-1,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	assert.True(t, check.IfNil(ls))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListsSharder_InvalidCrossShardObserversShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers-1,
		minAllowedFullHistoryNodes,
	)

	assert.True(t, check.IfNil(ls))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListsSharder_NoRoomForUnknownShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers+1,
		minAllowedFullHistoryNodes,
	)

	assert.True(t, check.IfNil(ls))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewListsSharder_ShouldWork(t *testing.T) {
	t.Parallel()

	ls, err := NewListsSharder(
		&mock.PeerShardResolverStub{},
		"",
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	assert.False(t, check.IfNil(ls))
	assert.Nil(t, err)
}

//------- ComputeEvictionList

func TestListsSharder_ComputeEvictionListNotReachedValidatorsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)
	pidCrtShard := peer.ID(fmt.Sprintf("%d %s", crtShardId, validatorMarker))
	pidCrossShard := peer.ID(fmt.Sprintf("%d %s", crossShardId, validatorMarker))
	pids := []peer.ID{pidCrtShard, pidCrossShard}

	evictList := ls.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListsSharder_ComputeEvictionListNotReachedObserversShouldRetEmpty(t *testing.T) {
	t.Parallel()

	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)
	pidCrtShard := peer.ID(fmt.Sprintf("%d %s", crtShardId, observerMarker))
	pidCrossShard := peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker))
	pids := []peer.ID{pidCrtShard, pidCrossShard}

	evictList := ls.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListsSharder_ComputeEvictionListNotReachedUnknownShouldRetEmpty(t *testing.T) {
	t.Parallel()

	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)
	pidUnknown := peer.ID(fmt.Sprintf("0 %s", unknownMarker))
	pids := []peer.ID{pidUnknown}

	evictList := ls.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListsSharder_ComputeEvictionListReachedIntraShardShouldSortAndEvict(t *testing.T) {
	t.Parallel()

	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)
	pidCrtShard1 := peer.ID(fmt.Sprintf("%d - 1 - %s", crtShardId, validatorMarker))
	pidCrtShard2 := peer.ID(fmt.Sprintf("%d - 2 - %s", crtShardId, validatorMarker))
	pids := []peer.ID{pidCrtShard2, pidCrtShard1}

	evictList := ls.ComputeEvictionList(pids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, pidCrtShard1, evictList[0])
}

func TestListsSharder_ComputeEvictionListUnknownPeersShouldFillTheGap(t *testing.T) {
	t.Parallel()

	maxPeerCount := 5
	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		maxPeerCount,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	unknownPids := make([]peer.ID, maxPeerCount)
	for i := 0; i < maxPeerCount; i++ {
		unknownPids[i] = unknownMarker
	}
	newUnknownPid := peer.ID(unknownMarker)
	unknownPids = append(unknownPids, newUnknownPid)

	evictList := ls.ComputeEvictionList(unknownPids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, unknownPids[0], evictList[0])
}

func TestListsSharder_ComputeEvictionListCrossShouldFillTheGap(t *testing.T) {
	t.Parallel()

	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		5,
		1,
		1,
		1,
		1,
		0,
	)

	pids := []peer.ID{
		peer.ID(fmt.Sprintf("%d %s", crossShardId, validatorMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, validatorMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
	}

	evictList := ls.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListsSharder_ComputeEvictionListEvictFromAllShouldWork(t *testing.T) {
	t.Parallel()

	ls, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		5,
		1,
		1,
		1,
		1,
		0,
	)

	pids := []peer.ID{
		peer.ID(fmt.Sprintf("%d %s", crtShardId, validatorMarker)),
		peer.ID(fmt.Sprintf("%d %s", crtShardId, validatorMarker)),

		peer.ID(fmt.Sprintf("%d %s", crossShardId, validatorMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, validatorMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, validatorMarker)),

		peer.ID(fmt.Sprintf("%d %s", crtShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crtShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crtShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crtShardId, observerMarker)),

		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, observerMarker)),

		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
		peer.ID(fmt.Sprintf("%d %s", crossShardId, unknownMarker)),
	}

	evictList := ls.ComputeEvictionList(pids)

	assert.Equal(t, 15, len(evictList))
	assert.Equal(t, 1, countPeers(evictList, crtShardId, validatorMarker))
	assert.Equal(t, 2, countPeers(evictList, crossShardId, validatorMarker))
	assert.Equal(t, 3, countPeers(evictList, crtShardId, observerMarker))
	assert.Equal(t, 4, countPeers(evictList, crossShardId, observerMarker))
	assert.Equal(t, 5, countPeers(evictList, crossShardId, unknownMarker))
}

//------- Has

func TestListsSharder_HasNotFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	ls := &listsSharder{}

	assert.False(t, ls.Has("pid4", list))
}

func TestListsSharder_HasEmpty(t *testing.T) {
	t.Parallel()

	list := make([]peer.ID, 0)
	lks := &listsSharder{}

	assert.False(t, lks.Has("pid4", list))
}

func TestListsSharder_HasFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lks := &listsSharder{}

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

func TestListsSharder_SetPeerShardResolverNilShouldErr(t *testing.T) {
	t.Parallel()

	lks, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)

	err := lks.SetPeerShardResolver(nil)

	assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
}

func TestListsSharder_SetPeerShardResolverShouldWork(t *testing.T) {
	t.Parallel()

	lks, _ := NewListsSharder(
		createStringPeersShardResolver(),
		crtPid,
		minAllowedConnectedPeersListSharder,
		minAllowedValidators,
		minAllowedValidators,
		minAllowedObservers,
		minAllowedObservers,
		minAllowedFullHistoryNodes,
	)
	newPeerShardResolver := &mock.PeerShardResolverStub{}
	err := lks.SetPeerShardResolver(newPeerShardResolver)

	//pointer testing
	assert.True(t, lks.peerShardResolver == newPeerShardResolver)
	assert.Nil(t, err)
}
