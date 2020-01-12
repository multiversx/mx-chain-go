package libp2p

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func init() {
	DurationBetweenReconnectAttempts = time.Millisecond
}

var durTimeoutWaiting = time.Second * 2
var durStartGoRoutine = time.Second

func TestNewLibp2pConnectionMonitor_WithNegativeThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cm, err := newLibp2pConnectionMonitor(nil, -1)

	assert.Equal(t, p2p.ErrInvalidValue, err)
	assert.Nil(t, cm)
}

func TestNewLibp2pConnectionMonitor_WithNilReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := newLibp2pConnectionMonitor(nil, 3)

	assert.Nil(t, err)
	assert.NotNil(t, cm)
}

func TestNewLibp2pConnectionMonitor_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
	t.Parallel()

	chReconnectCalled := make(chan struct{}, 1)

	rs := mock.ReconnecterStub{
		ReconnectToNetworkCalled: func() <-chan struct{} {
			ch := make(chan struct{}, 1)
			ch <- struct{}{}

			chReconnectCalled <- struct{}{}

			return ch
		},
	}

	ns := mock.NetworkStub{
		PeersCall: func() []peer.ID {
			//only one connection which is under the threshold
			return []peer.ID{"mock"}
		},
	}

	cm, _ := newLibp2pConnectionMonitor(&rs, 3)
	time.Sleep(durStartGoRoutine)
	cm.Disconnected(&ns, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitor_ConnectedNilSharderShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panic")
		}
	}()

	cm, _ := newLibp2pConnectionMonitor(nil, 3)

	cm.Connected(nil, nil)
}

func TestLibp2pConnectionMonitor_ConnectedWithSharderShouldCallEvictAndClosePeer(t *testing.T) {
	t.Parallel()

	evictedPid := []peer.ID{"evicted"}
	numComputeWasCalled := 0
	numClosedWasCalled := 0
	cm, _ := newLibp2pConnectionMonitor(nil, 3)
	err := cm.SetSharder(&mock.SharderStub{
		ComputeEvictListCalled: func(pid peer.ID, connected []peer.ID) []peer.ID {
			numComputeWasCalled++
			return evictedPid
		},
	})
	assert.Nil(t, err)

	cm.Connected(
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

func TestLibp2pConnectionMonitor_EmptyFuncsShouldNotPanic(t *testing.T) {
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

	cm, _ := newLibp2pConnectionMonitor(nil, 3)

	cm.ClosedStream(netw, nil)
	cm.Disconnected(netw, nil)
	cm.Listen(netw, nil)
	cm.ListenClose(netw, nil)
	cm.OpenedStream(netw, nil)
}

func TestLibp2pConnectionMonitor_SetSharderNilSharderShouldErr(t *testing.T) {
	t.Parallel()

	cm, _ := newLibp2pConnectionMonitor(nil, 3)
	err := cm.SetSharder(nil)

	assert.Equal(t, p2p.ErrNilSharder, err)
}
