package connectionMonitor

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewLibp2pConnectionMonitorSimple_WithNegativeThresholdShouldErr(t *testing.T) {
	t.Parallel()

	lcms, err := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, -1)

	assert.Equal(t, p2p.ErrInvalidValue, err)
	assert.Nil(t, lcms)
}

func TestNewLibp2pConnectionMonitorSimple_WithNilReconnecterShouldErr(t *testing.T) {
	t.Parallel()

	lcms, err := NewLibp2pConnectionMonitorSimple(nil, 3)

	assert.Equal(t, p2p.ErrNilReconnecter, err)
	assert.Nil(t, lcms)
}

func TestNewLibp2pConnectionMonitorSimple_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
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

	lcms, _ := NewLibp2pConnectionMonitorSimple(&rs, 3)
	time.Sleep(durStartGoRoutine)
	lcms.Disconnected(&ns, nil)

	select {
	case <-chReconnectCalled:
	case <-time.After(durTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitorSimple_ConnectedNilSharderShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panic")
		}
	}()

	lcms, _ := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3)

	lcms.Connected(nil, nil)
}

func TestLibp2pConnectionMonitorSimple_ConnectedWithSharderShouldCallEvictAndClosePeer(t *testing.T) {
	t.Parallel()

	evictedPid := []peer.ID{"evicted"}
	numComputeWasCalled := 0
	numClosedWasCalled := 0
	lcms, _ := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3)
	err := lcms.SetSharder(&mock.SharderStub{
		ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
			numComputeWasCalled++
			return evictedPid
		},
	})
	assert.Nil(t, err)

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

	lcms, _ := NewLibp2pConnectionMonitorSimple(&mock.ReconnecterStub{}, 3)

	lcms.ClosedStream(netw, nil)
	lcms.Disconnected(netw, nil)
	lcms.Listen(netw, nil)
	lcms.ListenClose(netw, nil)
	lcms.OpenedStream(netw, nil)
}

func TestLibp2pConnectionMonitorSimple_SetSharderNilSharderShouldErr(t *testing.T) {
	t.Parallel()

	lcms, _ := NewLibp2pConnectionMonitorSimple(nil, 3)
	err := lcms.SetSharder(nil)

	assert.Equal(t, p2p.ErrNilSharder, err)
}
