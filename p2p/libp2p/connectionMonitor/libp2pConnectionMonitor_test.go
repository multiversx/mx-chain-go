package connectionMonitor

import (
	"math"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	ns "github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewLibp2pConnectionMonitor_WithNegativeThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cm, err := NewLibp2pConnectionMonitor(nil, -1, 0, &ns.SimplePrioBitsSharder{})

	assert.Equal(t, p2p.ErrInvalidValue, err)
	assert.True(t, check.IfNil(cm))
}

func TestNewLibp2pConnectionMonitor_WithNilReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := NewLibp2pConnectionMonitor(nil, 3, 0, &ns.SimplePrioBitsSharder{})

	assert.Nil(t, err)
	assert.False(t, check.IfNil(cm))
}

func TestNewLibp2pConnectionMonitor_WithNilSharderShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := NewLibp2pConnectionMonitor(nil, 3, 0, nil)

	assert.Equal(t, p2p.ErrNilSharder, err)
	assert.True(t, check.IfNil(cm))
}

func TestNewLibp2pConnectionMonitor_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
	t.Parallel()

	chReconnectCalled := make(chan struct{}, 1)

	rs := mock.ReconnecterWithPauseAndResumeStub{
		ReconnectToNetworkCalled: func() <-chan struct{} {
			ch := make(chan struct{}, 1)
			ch <- struct{}{}

			chReconnectCalled <- struct{}{}

			return ch
		},
	}

	netw := &mock.NetworkStub{
		ConnsCalled: func() []network.Conn {
			//only one connection which is under the threshold
			return []network.Conn{
				&mock.ConnStub{},
			}
		},
	}

	cm, _ := NewLibp2pConnectionMonitor(&rs, 3, 0, &ns.SimplePrioBitsSharder{})
	time.Sleep(durationStartGoRoutine)
	cm.Disconnected(netw, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durationTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestNewLibp2pConnectionMonitor_DefaultTriming(t *testing.T) {
	t.Parallel()

	cm, _ := NewLibp2pConnectionMonitor(nil, 3, 0, &ns.SimplePrioBitsSharder{})

	assert.NotNil(t, cm)
	assert.Equal(t, 0, cm.thresholdDiscoveryResume)
	assert.Equal(t, math.MaxInt32, cm.thresholdDiscoveryPause)
	assert.Equal(t, math.MaxInt32, cm.thresholdConnTrim)
}

func TestNewLibp2pConnectionMonitor_Triming(t *testing.T) {
	t.Parallel()

	pauseCallCount := 0
	resumeCallCount := 0

	rc := mock.ReconnecterWithPauseAndResumeStub{
		ReconnectToNetworkCalled: func() <-chan struct{} {
			ch := make(chan struct{})
			defer func() { ch <- struct{}{} }()
			return ch
		},
		PauseCall:  func() { pauseCallCount++ },
		ResumeCall: func() { resumeCallCount++ },
	}

	cm, _ := NewLibp2pConnectionMonitor(&rc, 3, 10, &ns.SimplePrioBitsSharder{})

	assert.NotNil(t, cm)
	assert.Equal(t, 8, cm.thresholdDiscoveryResume)
	assert.Equal(t, 10, cm.thresholdDiscoveryPause)
	assert.Equal(t, 12, cm.thresholdConnTrim)

	netFact := func(cnt int) network.Network {
		cntr := cnt
		currentCount := &cntr
		return &mock.NetworkStub{
			ConnsCalled: func() []network.Conn {
				return make([]network.Conn, *currentCount)
			},

			PeersCall: func() []peer.ID {
				return make([]peer.ID, *currentCount)
			},

			ClosePeerCall: func(peer.ID) error {
				*currentCount--
				return nil
			},
		}
	}

	assert.Equal(t, 0, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.Connected(netFact(5), nil)
	assert.Equal(t, 0, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.Connected(netFact(9), nil)
	assert.Equal(t, 0, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	// this is triggering a trim and pause
	cm.Connected(netFact(13), nil)
	assert.Equal(t, 1, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	// this should not resume
	cm.Connected(netFact(9), nil)
	assert.Equal(t, 1, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.Disconnected(netFact(5), nil)
	assert.Equal(t, 1, pauseCallCount)
	assert.Equal(t, 1, resumeCallCount)

	cm.Connected(netFact(13), nil)
	assert.Equal(t, 2, pauseCallCount)
	assert.Equal(t, 1, resumeCallCount)

	cm.Disconnected(netFact(5), nil)
	assert.Equal(t, 2, pauseCallCount)
	assert.Equal(t, 2, resumeCallCount)
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

	lcm, _ := NewLibp2pConnectionMonitor(nil, 3, 10, &ns.SimplePrioBitsSharder{})

	lcm.ClosedStream(netw, nil)
	lcm.Disconnected(netw, nil)
	lcm.Listen(netw, nil)
	lcm.ListenClose(netw, nil)
	lcm.OpenedStream(netw, nil)
}

func TestLibp2pConnectionMonitor_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	lcm, _ := NewLibp2pConnectionMonitor(nil, 3, 10, &ns.SimplePrioBitsSharder{})

	thr := 10
	lcm.SetThresholdMinConnectedPeers(thr, &mock.NetworkStub{})
	thrSet := lcm.ThresholdMinConnectedPeers()

	assert.Equal(t, thr, thrSet)
}

func TestLibp2pConnectionMonitor_SetThresholdMinConnectedPeersNilNetwShouldDoNothing(t *testing.T) {
	t.Parallel()

	minConnPeers := 3
	lcm, _ := NewLibp2pConnectionMonitor(nil, minConnPeers, 10, &ns.SimplePrioBitsSharder{})

	thr := 10
	lcm.SetThresholdMinConnectedPeers(thr, nil)
	thrSet := lcm.ThresholdMinConnectedPeers()

	assert.Equal(t, thrSet, minConnPeers)
}
