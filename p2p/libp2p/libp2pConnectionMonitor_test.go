package libp2p

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
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
