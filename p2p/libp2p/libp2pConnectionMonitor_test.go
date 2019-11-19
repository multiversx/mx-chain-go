package libp2p

import (
	"math"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
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

	cm, err := newLibp2pConnectionMonitor(nil, -1, 0)

	assert.Equal(t, p2p.ErrInvalidValue, err)
	assert.Nil(t, cm)
}

func TestNewLibp2pConnectionMonitor_WithNilReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := newLibp2pConnectionMonitor(nil, 3, 0)

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
		ConnsCalled: func() []network.Conn {
			//only one connection which is under the threshold
			return []network.Conn{
				&mock.ConnStub{},
			}
		},
	}

	cm, _ := newLibp2pConnectionMonitor(&rs, 3, 0)
	time.Sleep(durStartGoRoutine)
	cm.Disconnected(&ns, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestNewLibp2pConnectionMonitor_DefaultTriming(t *testing.T) {
	t.Parallel()

	cm, _ := newLibp2pConnectionMonitor(nil, 3, 0)

	assert.NotNil(t, cm)
	assert.Equal(t, 0, cm.thresholdDiscoveryResume)
	assert.Equal(t, math.MaxInt32, cm.thresholdDiscoveryPause)
	assert.Equal(t, math.MaxInt32, cm.thresholdConnTrim)
}

func TestNewLibp2pConnectionMonitor_Triming(t *testing.T) {
	t.Parallel()

	pauseCallCount := 0
	resumeCallCount := 0

	rc := mock.ReconnecterStub{
		ReconnectToNetworkCalled: func() <-chan struct{} {
			ch := make(chan struct{})
			defer func() { ch <- struct{}{} }()
			return ch
		},
		PauseCall:  func() { pauseCallCount++ },
		ResumeCall: func() { resumeCallCount++ },
	}

	cm, _ := newLibp2pConnectionMonitor(&rc, 3, 10)

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
