package connectionMonitor

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const durationTimeoutWaiting = time.Second * 2
const durationStartGoRoutine = time.Second

func createMockArgsConnectionMonitorSimple() ArgsConnectionMonitorSimple {
	return ArgsConnectionMonitorSimple{
		Reconnecter:                &mock.ReconnecterStub{},
		ThresholdMinConnectedPeers: 3,
		Sharder:                    &mock.KadSharderStub{},
		PreferredPeersHolder:       &p2pmocks.PeersHolderStub{},
		ConnectionsWatcher:         &mock.ConnectionsWatcherStub{},
	}
}

func TestNewLibp2pConnectionMonitorSimple(t *testing.T) {
	t.Parallel()

	t.Run("nil reconnecter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.Reconnecter = nil
		lcms, err := NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilReconnecter, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil sharder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.Sharder = nil
		lcms, err := NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilSharder, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil preferred peers holder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.PreferredPeersHolder = nil
		lcms, err := NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilPreferredPeersHolder, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil connections watcher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.ConnectionsWatcher = nil
		lcms, err := NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilConnectionsWatcher, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		lcms, err := NewLibp2pConnectionMonitorSimple(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(lcms))
	})
}

func TestNewLibp2pConnectionMonitorSimple_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
	t.Parallel()

	chReconnectCalled := make(chan struct{}, 1)

	rs := &mock.ReconnecterStub{
		ReconnectToNetworkCalled: func(ctx context.Context) {
			chReconnectCalled <- struct{}{}
		},
	}

	ns := mock.NetworkStub{
		PeersCall: func() []peer.ID {
			// only one connection which is under the threshold
			return []peer.ID{"mock"}
		},
	}

	args := createMockArgsConnectionMonitorSimple()
	args.Reconnecter = rs
	lcms, _ := NewLibp2pConnectionMonitorSimple(args)
	time.Sleep(durationStartGoRoutine)
	lcms.Disconnected(&ns, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durationTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitorSimple_ConnectedWithSharderShouldCallEvictAndClosePeer(t *testing.T) {
	t.Parallel()

	evictedPid := []peer.ID{"evicted"}
	numComputeWasCalled := 0
	numClosedWasCalled := 0
	args := createMockArgsConnectionMonitorSimple()
	args.Sharder = &mock.KadSharderStub{
		ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
			numComputeWasCalled++
			return evictedPid
		},
	}
	knownConnectionCalled := false
	args.ConnectionsWatcher = &mock.ConnectionsWatcherStub{
		NewKnownConnectionCalled: func(pid core.PeerID, connection string) {
			knownConnectionCalled = true
		},
	}
	putConnectionAddressCalled := false
	args.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		PutConnectionAddressCalled: func(peerID core.PeerID, addressSlice string) {
			putConnectionAddressCalled = true
		},
	}
	lcms, _ := NewLibp2pConnectionMonitorSimple(args)

	lcms.Connected(
		&mock.NetworkStub{
			ClosePeerCall: func(id peer.ID) error {
				numClosedWasCalled++
				return nil
			},
			PeersCall: func() []peer.ID {
				return nil
			},
		},
		&mock.ConnStub{
			RemotePeerCalled: func() peer.ID {
				return evictedPid[0]
			},
		},
	)

	assert.Equal(t, 1, numClosedWasCalled)
	assert.Equal(t, 1, numComputeWasCalled)
	assert.True(t, knownConnectionCalled)
	assert.True(t, putConnectionAddressCalled)
}

func TestNewLibp2pConnectionMonitorSimple_DisconnectedShouldRemovePeerFromPreferredPeers(t *testing.T) {
	t.Parallel()

	prefPeerID := "preferred peer 0"
	chRemoveCalled := make(chan struct{}, 1)

	ns := mock.NetworkStub{
		PeersCall: func() []peer.ID {
			// only one connection which is under the threshold
			return []peer.ID{"mock"}
		},
	}

	removeCalled := false
	prefPeersHolder := &p2pmocks.PeersHolderStub{
		RemoveCalled: func(peerID core.PeerID) {
			removeCalled = true
			require.Equal(t, core.PeerID(prefPeerID), peerID)
			chRemoveCalled <- struct{}{}
		},
	}

	args := createMockArgsConnectionMonitorSimple()
	args.PreferredPeersHolder = prefPeersHolder
	lcms, _ := NewLibp2pConnectionMonitorSimple(args)
	lcms.Disconnected(&ns, &mock.ConnStub{
		IDCalled: func() string {
			return prefPeerID
		},
	})

	require.True(t, removeCalled)
	select {
	case <-chRemoveCalled:
	case <-time.After(durationTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitorSimple_EmptyFuncsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panic")
		}
	}()

	netw := &mock.NetworkStub{
		PeersCall: func() []peer.ID {
			return make([]peer.ID, 0)
		},
	}

	args := createMockArgsConnectionMonitorSimple()
	lcms, _ := NewLibp2pConnectionMonitorSimple(args)

	lcms.ClosedStream(netw, nil)
	lcms.Disconnected(netw, nil)
	lcms.Listen(netw, nil)
	lcms.ListenClose(netw, nil)
	lcms.OpenedStream(netw, nil)
}

func TestLibp2pConnectionMonitorSimple_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	args := createMockArgsConnectionMonitorSimple()
	lcms, _ := NewLibp2pConnectionMonitorSimple(args)

	thr := 10
	lcms.SetThresholdMinConnectedPeers(thr, &mock.NetworkStub{})
	thrSet := lcms.ThresholdMinConnectedPeers()

	assert.Equal(t, thr, thrSet)
}

func TestLibp2pConnectionMonitorSimple_SetThresholdMinConnectedPeersNilNetwShouldDoNothing(t *testing.T) {
	t.Parallel()

	minConnPeers := uint32(3)
	args := createMockArgsConnectionMonitorSimple()
	args.ThresholdMinConnectedPeers = minConnPeers
	lcms, _ := NewLibp2pConnectionMonitorSimple(args)

	thr := 10
	lcms.SetThresholdMinConnectedPeers(thr, nil)
	thrSet := lcms.ThresholdMinConnectedPeers()

	assert.Equal(t, uint32(thrSet), minConnPeers)
}
