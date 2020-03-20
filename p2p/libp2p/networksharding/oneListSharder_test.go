package networksharding

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewOneListSharder_InvalidMaxPeerCountShouldErr(t *testing.T) {
	t.Parallel()

	ols, err := NewOneListSharder(
		"",
		minAllowedConnectedPeersOneSharder-1,
	)

	assert.True(t, check.IfNil(ols))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewOneListSharder_ShouldWork(t *testing.T) {
	t.Parallel()

	ols, err := NewOneListSharder(
		"",
		minAllowedConnectedPeersOneSharder,
	)

	assert.False(t, check.IfNil(ols))
	assert.Nil(t, err)
}

//------- ComputeEvictionList

func TestOneListSharder_ComputeEvictionListNotReachedShouldRetEmpty(t *testing.T) {
	t.Parallel()

	ols, _ := NewOneListSharder(
		crtPid,
		minAllowedConnectedPeersOneSharder,
	)
	pid1 := peer.ID("pid1")
	pid2 := peer.ID("pid2")
	pids := []peer.ID{pid1, pid2}

	evictList := ols.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestOneListSharder_ComputeEvictionListReachedIntraShardShouldSortAndEvict(t *testing.T) {
	t.Parallel()

	ols, _ := NewOneListSharder(
		crtPid,
		minAllowedConnectedPeersOneSharder,
	)
	pid1 := peer.ID("pid1")
	pid2 := peer.ID("pid2")
	pid3 := peer.ID("pid3")
	pid4 := peer.ID("pid4")
	pids := []peer.ID{pid1, pid2, pid3, pid4}

	evictList := ols.ComputeEvictionList(pids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, pid3, evictList[0])
}

//------- Has

func TestOneListSharder_HasNotFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lnks := &oneListSharder{}

	assert.False(t, lnks.Has("pid4", list))
}

func TestOneListSharder_HasEmpty(t *testing.T) {
	t.Parallel()

	list := make([]peer.ID, 0)
	lnks := &oneListSharder{}

	assert.False(t, lnks.Has("pid4", list))
}

func TestOneListSharder_HasFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lnks := &oneListSharder{}

	assert.True(t, lnks.Has("pid2", list))
}

func TestOneListSharder_SetPeerShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	ols, _ := NewOneListSharder(
		"",
		minAllowedConnectedPeersOneSharder,
	)

	err := ols.SetPeerShardResolver(nil)

	assert.Nil(t, err)
}
