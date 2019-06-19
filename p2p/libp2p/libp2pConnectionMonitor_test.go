package libp2p

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/stretchr/testify/assert"
)

func init() {
	ThresholdMinimumConnectedPeers = 3
	DurationBetweenReconnectAttempts = time.Duration(time.Millisecond)
}

var durTimeoutWaiting = time.Duration(time.Second * 2)
var durStartGoRoutine = time.Duration(time.Second)

func TestNewLibp2pConnectionMonitor_WithNilReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	cm := newLibp2pConnectionMonitor(nil)

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

	cm := newLibp2pConnectionMonitor(&rs)
	time.Sleep(durStartGoRoutine)
	cm.Disconnected(&ns, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}
