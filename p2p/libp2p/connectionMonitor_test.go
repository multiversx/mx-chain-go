package libp2p

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

var durTimeoutWaiting = time.Second * 2

func createNetwork(numPeers int) network.Network {
	numPeersCopy := numPeers
	currentCount := &numPeersCopy
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

func createStubConnectableHost() *mock.ConnectableHostStub {
	return &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{}
		},
	}
}

func TestNewConnectionMonitor_InvalidMinConnectedPeersShouldErr(t *testing.T) {
	t.Parallel()

	libp2pContext := &Libp2pContext{
		connHost: createStubConnectableHost(),
	}
	cm, err := newConnectionMonitor(nil, libp2pContext, -1, 1)

	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	assert.True(t, check.IfNil(cm))
}

func TestNewConnectionMonitor_InvalidTargetCountShouldErr(t *testing.T) {
	t.Parallel()

	libp2pContext := &Libp2pContext{
		connHost: createStubConnectableHost(),
	}
	cm, err := newConnectionMonitor(nil, libp2pContext, 1, -1)

	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	assert.True(t, check.IfNil(cm))
}

func TestNewConnectionMonitor_NilLibP2pContextShouldErr(t *testing.T) {
	t.Parallel()

	cm, err := newConnectionMonitor(&mock.ReconnecterStub{}, nil, 3, 1)

	assert.True(t, errors.Is(err, p2p.ErrNilContextProvider))
	assert.True(t, check.IfNil(cm))
}

func TestNewConnectionMonitor_WithNilReconnecterShouldWork(t *testing.T) {
	t.Parallel()

	libp2pContext := &Libp2pContext{
		connHost: createStubConnectableHost(),
	}
	cm, err := newConnectionMonitor(nil, libp2pContext, 3, 1)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(cm))
}

func TestConnectionMonitor_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
	t.Parallel()

	chReconnectCalled := make(chan struct{}, 1)

	rs := mock.ReconnecterStub{
		ReconnectToNetworkCalled: func() {
			chReconnectCalled <- struct{}{}
		},
	}

	ns := &mock.NetworkStub{
		ConnsCalled: func() []network.Conn {
			//only one connection which is under the threshold
			return []network.Conn{
				&mock.ConnStub{},
			}
		},
		PeersCall: func() []peer.ID { return nil },
	}

	libp2pContext := &Libp2pContext{
		connHost: &mock.ConnectableHostStub{
			NetworkCalled: func() network.Network {
				return ns
			},
		},
		blacklistHandler: &mock.NilBlacklistHandler{},
	}
	cm, _ := newConnectionMonitor(&rs, libp2pContext, 3, 1)

	go func() {
		for {
			cm.DoReconnectionBlocking()
		}
	}()

	//wait for the reconnection blocking go routine to start
	time.Sleep(time.Second)

	cm.HandleDisconnectedPeer("")

	select {
	case <-chReconnectCalled:
	case <-time.After(durTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestConnectionMonitor_HandleConnectedPeerShouldTrim(t *testing.T) {
	t.Parallel()

	pauseCallCount := 0
	resumeCallCount := 0

	rc := mock.ReconnecterStub{
		ReconnectToNetworkCalled: func() {},
		PauseCall:                func() { pauseCallCount++ },
		ResumeCall:               func() { resumeCallCount++ },
	}

	libp2pContext := &Libp2pContext{
		blacklistHandler: &mock.NilBlacklistHandler{},
		connHost:         createStubConnectableHost(),
	}
	cm, _ := newConnectionMonitor(&rc, libp2pContext, 3, 10)

	assert.NotNil(t, cm)
	assert.Equal(t, 8, cm.thresholdDiscoveryResume)
	assert.Equal(t, 10, cm.thresholdDiscoveryPause)
	assert.Equal(t, 12, cm.thresholdConnTrim)

	pid := p2p.PeerID("remote peer")

	assert.Equal(t, 0, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.netw = createNetwork(5)
	_ = cm.HandleConnectedPeer(pid)
	assert.Equal(t, 0, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.netw = createNetwork(9)
	_ = cm.HandleConnectedPeer(pid)
	assert.Equal(t, 0, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.netw = createNetwork(11)
	_ = cm.HandleConnectedPeer(pid)
	assert.Equal(t, 1, pauseCallCount)
	assert.Equal(t, 0, resumeCallCount)

	cm.netw = createNetwork(5)
	cm.HandleDisconnectedPeer(pid)
	assert.Equal(t, 1, pauseCallCount)
	assert.Equal(t, 1, resumeCallCount)

	cm.netw = createNetwork(13)
	_ = cm.HandleConnectedPeer(pid)
	assert.Equal(t, 2, pauseCallCount)
	assert.Equal(t, 1, resumeCallCount)

	cm.netw = createNetwork(5)
	cm.HandleDisconnectedPeer(pid)
	assert.Equal(t, 2, pauseCallCount)
	assert.Equal(t, 2, resumeCallCount)
}

func TestConnectionMonitor_BlackListedPeerShouldErr(t *testing.T) {
	t.Parallel()

	pid := p2p.PeerID("remote peer")
	libp2pContext := &Libp2pContext{
		blacklistHandler: &mock.BlacklistHandlerStub{
			HasCalled: func(key string) bool {
				return key == pid.Pretty()
			},
		},
		connHost: createStubConnectableHost(),
	}
	cm, _ := newConnectionMonitor(nil, libp2pContext, 1, 1)

	err := cm.HandleConnectedPeer(pid)

	assert.True(t, errors.Is(err, p2p.ErrPeerBlacklisted))
}

func TestConnectionMonitor_RemoveAnAlreadyConnectedBlacklistedPeer(t *testing.T) {
	t.Parallel()

	pid := peer.ID("remote peer")
	numCloseCalled := 0

	ns := &mock.NetworkStub{
		ClosePeerCall: func(id peer.ID) error {
			if id == pid {
				numCloseCalled++
			}
			return nil
		},
		PeersCall: func() []peer.ID {
			return []peer.ID{pid}
		},
	}

	libp2pContext := &Libp2pContext{
		blacklistHandler: &mock.BlacklistHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		connHost: &mock.ConnectableHostStub{
			NetworkCalled: func() network.Network {
				return ns
			},
		},
	}
	cm, _ := newConnectionMonitor(nil, libp2pContext, 1, 1)

	cm.CheckConnectionsBlocking()

	assert.Equal(t, 1, numCloseCalled)
}

func TestConnectionMonitor_ClosePeerErrorsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	closePeerCalled := false
	expectedErr := errors.New("expected error")
	cm := &connectionMonitor{
		netw: &mock.NetworkStub{
			ClosePeerCall: func(id peer.ID) error {
				closePeerCalled = true
				return expectedErr
			},
		},
	}

	cm.closePeer("pid")

	assert.True(t, closePeerCalled)
}
