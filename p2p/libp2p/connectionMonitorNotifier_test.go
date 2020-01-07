package libp2p

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func createStubConn() *mock.ConnStub {
	return &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return "remote peer"
		},
	}
}

//------- Connected

func TestConnectionMonitorNotifier_ConnectedHandledWithErrShouldCloseConnection(t *testing.T) {
	t.Parallel()

	cms := &mock.ConnectionMonitorStub{
		HandleConnectedPeerCalled: func(pid p2p.PeerID) error {
			return errors.New("expected error")
		},
	}
	cmn := &connectionMonitorNotifier{
		ConnectionMonitor: cms,
	}
	peerCloseCalled := false
	conn := createStubConn()
	ns := &mock.NetworkStub{
		ClosePeerCall: func(id peer.ID) error {
			if id == conn.RemotePeer() {
				peerCloseCalled = true
			}
			return nil
		},
	}

	cmn.Connected(ns, conn)

	assert.True(t, peerCloseCalled)
}

func TestConnectionMonitorNotifier_ConnectedHandledWithErrShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	expectedErr := errors.New("expected error")
	cms := &mock.ConnectionMonitorStub{
		HandleConnectedPeerCalled: func(pid p2p.PeerID) error {
			return expectedErr
		},
	}
	cmn := &connectionMonitorNotifier{
		ConnectionMonitor: cms,
	}
	conn := createStubConn()
	ns := &mock.NetworkStub{
		ClosePeerCall: func(id peer.ID) error {
			return expectedErr
		},
	}

	cmn.Connected(ns, conn)
}

//------- Disconnected

func TestConnectionMonitorNotifier_DisconnectedShouldCallHandler(t *testing.T) {
	t.Parallel()

	handlerCalled := false
	cms := &mock.ConnectionMonitorStub{
		HandleDisconnectedPeerCalled: func(pid p2p.PeerID) {
			handlerCalled = true
		},
	}
	cmn := &connectionMonitorNotifier{
		ConnectionMonitor: cms,
	}
	conn := createStubConn()

	cmn.Disconnected(nil, conn)

	assert.True(t, handlerCalled)
}

//------- handlers

func TestConnectionMonitorNotifier_CallingHandlersShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	cmn := &connectionMonitorNotifier{}

	cmn.Listen(nil, nil)
	cmn.ListenClose(nil, nil)
	cmn.OpenedStream(nil, nil)
	cmn.ClosedStream(nil, nil)
}
