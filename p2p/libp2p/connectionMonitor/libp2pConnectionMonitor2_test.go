package connectionMonitor

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewLibp2pConnectionMonitor2_WithNegativeThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cm, err := NewLibp2pConnectionMonitor2(nil, -1)

	assert.Equal(t, p2p.ErrInvalidValue, err)
	assert.Nil(t, cm)
}

func TestNewLibp2pConnectionMonitor2_WithNilReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := NewLibp2pConnectionMonitor2(nil, 3)

	assert.Nil(t, err)
	assert.NotNil(t, cm)
}

func TestNewLibp2pConnectionMonitor2_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
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

	cm, _ := NewLibp2pConnectionMonitor2(&rs, 3)
	time.Sleep(durStartGoRoutine)
	cm.Disconnected(&ns, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitor2_ConnectedNilSharderShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panic")
		}
	}()

	cm, _ := NewLibp2pConnectionMonitor2(nil, 3)

	cm.Connected(nil, nil)
}

func TestLibp2pConnectionMonitor2_ConnectedWithSharderShouldCallEvictAndClosePeer(t *testing.T) {
	t.Parallel()

	evictedPid := []peer.ID{"evicted"}
	numComputeWasCalled := 0
	numClosedWasCalled := 0
	cm, _ := NewLibp2pConnectionMonitor2(nil, 3)
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

func TestLibp2pConnectionMonitor2_EmptyFuncsShouldNotPanic(t *testing.T) {
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

	cm, _ := NewLibp2pConnectionMonitor2(nil, 3)

	cm.ClosedStream(netw, nil)
	cm.Disconnected(netw, nil)
	cm.Listen(netw, nil)
	cm.ListenClose(netw, nil)
	cm.OpenedStream(netw, nil)
}

func TestLibp2pConnectionMonitor2_SetSharderNilSharderShouldErr(t *testing.T) {
	t.Parallel()

	cm, _ := NewLibp2pConnectionMonitor2(nil, 3)
	err := cm.SetSharder(nil)

	assert.True(t, errors.Is(err, p2p.ErrWrongTypeAssertion))
}
