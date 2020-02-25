package networksharding

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewNoListKadSharder_InvalidMaxPeerCountShouldErr(t *testing.T) {
	t.Parallel()

	lnks, err := NewListNoKadSharder(
		"",
		minAllowedConnectedPeers-1,
	)

	assert.True(t, check.IfNil(lnks))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewNoListKadSharder_ShouldWork(t *testing.T) {
	t.Parallel()

	lnks, err := NewListNoKadSharder(
		"",
		minAllowedConnectedPeers,
	)

	assert.False(t, check.IfNil(lnks))
	assert.Nil(t, err)
}

//------- ComputeEvictionList

func TestNewListKadSharder_ComputeEvictionListNotReachedShouldRetEmpty(t *testing.T) {
	t.Parallel()

	lnks, _ := NewListNoKadSharder(
		crtPid,
		minAllowedConnectedPeers,
	)
	pid1 := peer.ID("pid1")
	pid2 := peer.ID("pid2")
	pids := []peer.ID{pid1, pid2}

	evictList := lnks.ComputeEvictionList(pids)

	assert.Equal(t, 0, len(evictList))
}

func TestListNoKadSharder_ComputeEvictionListReachedIntraShardShouldSortAndEvict(t *testing.T) {
	t.Parallel()

	lnks, _ := NewListNoKadSharder(
		crtPid,
		minAllowedConnectedPeers,
	)
	pid1 := peer.ID("pid1")
	pid2 := peer.ID("pid2")
	pid3 := peer.ID("pid3")
	pids := []peer.ID{pid1, pid2, pid3}

	evictList := lnks.ComputeEvictionList(pids)

	assert.Equal(t, 1, len(evictList))
	assert.Equal(t, pid3, evictList[0])
}

//------- Has

func TestListNoKadSharder_HasNotFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lnks := &listNoKadSharder{}

	assert.False(t, lnks.Has("pid4", list))
}

func TestListNoKadSharder_HasEmpty(t *testing.T) {
	t.Parallel()

	list := make([]peer.ID, 0)
	lnks := &listNoKadSharder{}

	assert.False(t, lnks.Has("pid4", list))
}

func TestListNoKadSharder_HasFound(t *testing.T) {
	t.Parallel()

	list := []peer.ID{"pid1", "pid2", "pid3"}
	lnks := &listNoKadSharder{}

	assert.True(t, lnks.Has("pid2", list))
}

func TestListNoKadSharder_SetPeerShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	lnks, _ := NewListNoKadSharder(
		"",
		minAllowedConnectedPeers,
	)

	err := lnks.SetPeerShardResolver(nil)

	assert.Nil(t, err)
}
