package connectionMonitor

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

const durationTimeoutWaiting = time.Second * 2
const durationStartGoRoutine = time.Second

func TestNewLibp2pConnectionMonitorSimple_WithNilReconnecterShouldErr(t *testing.T) {
	t.Parallel()

	lcms, err := NewLibp2pConnectionMonitorSimple(nil, 3, &mock.SharderStub{})

	assert.Equal(t, p2p.ErrNilReconnecter, err)
	assert.True(t, check.IfNil(lcms))
}

func TestNewLibp2pConnectionMonitorSimple_WithNilSharderShouldErr(t *testing.T) {
	t.Parallel()

	lcms, err := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3, nil)

	assert.Equal(t, p2p.ErrNilSharder, err)
	assert.True(t, check.IfNil(lcms))
}

func TestNewLibp2pConnectionMonitorSimple_ShouldWork(t *testing.T) {
	t.Parallel()

	lcms, err := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3, &mock.SharderStub{})

	assert.Nil(t, err)
	assert.False(t, check.IfNil(lcms))
}

func TestNewLibp2pConnectionMonitorSimple_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
	t.Parallel()

	chReconnectCalled := make(chan struct{}, 1)

	rs := mock.ReconnecterStub{
		ReconnectToNetworkCalled: func(ctx context.Context) {
			chReconnectCalled <- struct{}{}
		},
	}

	ns := mock.NetworkStub{
		PeersCall: func() []peer.ID {
			//only one connection which is under the threshold
			return []peer.ID{"mock"}
		},
	}

	lcms, _ := NewLibp2pConnectionMonitorSimple(&rs, 3, &mock.SharderStub{})
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
	lcms, _ := NewLibp2pConnectionMonitorSimple(
		&mock.ReconnecterStub{},
		3,
		&mock.SharderStub{
			ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
				numComputeWasCalled++
				return evictedPid
			},
		},
	)

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
				return ""
			},
		},
	)

	assert.Equal(t, 1, numClosedWasCalled)
	assert.Equal(t, 1, numComputeWasCalled)
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

	lcms, _ := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3, &mock.SharderStub{})

	lcms.ClosedStream(netw, nil)
	lcms.Disconnected(netw, nil)
	lcms.Listen(netw, nil)
	lcms.ListenClose(netw, nil)
	lcms.OpenedStream(netw, nil)
}

func TestLibp2pConnectionMonitorSimple_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	lcms, _ := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3, &mock.SharderStub{})

	thr := 10
	lcms.SetThresholdMinConnectedPeers(thr, &mock.NetworkStub{})
	thrSet := lcms.ThresholdMinConnectedPeers()

	assert.Equal(t, thr, thrSet)
}

func TestLibp2pConnectionMonitorSimple_SetThresholdMinConnectedPeersNilNetwShouldDoNothing(t *testing.T) {
	t.Parallel()

	minConnPeers := uint32(3)
	lcms, _ := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, minConnPeers, &mock.SharderStub{})

	thr := 10
	lcms.SetThresholdMinConnectedPeers(thr, nil)
	thrSet := lcms.ThresholdMinConnectedPeers()

	assert.Equal(t, uint32(thrSet), minConnPeers)
}
