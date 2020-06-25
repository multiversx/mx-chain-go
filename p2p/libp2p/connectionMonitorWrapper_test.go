package libp2p

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func createStubConn() *mock.ConnStub {
	return &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return "remote peer"
		},
	}
}

func TestNewConnectionMonitorWrapper_ShouldWork(t *testing.T) {
	t.Parallel()

	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{},
		&mock.ConnectionMonitorStub{},
		&mock.PeerDenialEvaluatorStub{},
	)

	assert.False(t, check.IfNil(cmw))
}

//------- Connected

func TestConnectionMonitorNotifier_ConnectedBlackListedShouldCallClose(t *testing.T) {
	t.Parallel()

	peerCloseCalled := false
	conn := createStubConn()
	conn.CloseCalled = func() error {
		peerCloseCalled = true

		return nil
	}
	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{},
		&mock.ConnectionMonitorStub{},
		&mock.PeerDenialEvaluatorStub{
			IsDeniedCalled: func(pid core.PeerID) bool {
				return true
			},
		},
	)

	cmw.Connected(cmw.network, conn)

	assert.True(t, peerCloseCalled)
}

func TestConnectionMonitorNotifier_ConnectedNotBlackListedShouldCallConnected(t *testing.T) {
	t.Parallel()

	peerConnectedCalled := false
	conn := createStubConn()
	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{},
		&mock.ConnectionMonitorStub{
			ConnectedCalled: func(netw network.Network, conn network.Conn) {
				peerConnectedCalled = true
			},
		},
		&mock.PeerDenialEvaluatorStub{
			IsDeniedCalled: func(pid core.PeerID) bool {
				return false
			},
		},
	)

	cmw.Connected(cmw.network, conn)

	assert.True(t, peerConnectedCalled)
}

//------- Functions

func TestConnectionMonitorNotifier_FunctionsShouldCallHandler(t *testing.T) {
	t.Parallel()

	listenCalled := false
	listenCloseCalled := false
	disconnectCalled := false
	openedCalled := false
	closedCalled := false
	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{},
		&mock.ConnectionMonitorStub{
			ListenCalled: func(network.Network, multiaddr.Multiaddr) {
				listenCalled = true
			},
			ListenCloseCalled: func(network.Network, multiaddr.Multiaddr) {
				listenCloseCalled = true
			},
			DisconnectedCalled: func(network.Network, network.Conn) {
				disconnectCalled = true
			},
			OpenedStreamCalled: func(network.Network, network.Stream) {
				openedCalled = true
			},
			ClosedStreamCalled: func(network.Network, network.Stream) {
				closedCalled = true
			},
		},
		&mock.PeerDenialEvaluatorStub{},
	)

	cmw.Listen(nil, nil)
	cmw.ListenClose(nil, nil)
	cmw.Disconnected(nil, nil)
	cmw.OpenedStream(nil, nil)
	cmw.ClosedStream(nil, nil)

	assert.True(t, listenCalled)
	assert.True(t, listenCloseCalled)
	assert.True(t, disconnectCalled)
	assert.True(t, openedCalled)
	assert.True(t, closedCalled)
}

//------- SetBlackListHandler

func TestConnectionMonitorWrapper_SetBlackListHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{},
		&mock.ConnectionMonitorStub{},
		&mock.PeerDenialEvaluatorStub{},
	)

	err := cmw.SetPeerDenialEvaluator(nil)

	assert.Equal(t, p2p.ErrNilPeerDenialEvaluator, err)
}

func TestConnectionMonitorWrapper_SetBlackListHandlerShouldWork(t *testing.T) {
	t.Parallel()

	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{},
		&mock.ConnectionMonitorStub{},
		&mock.PeerDenialEvaluatorStub{},
	)
	newPeerDenialEvaluator := &mock.PeerDenialEvaluatorStub{}

	err := cmw.SetPeerDenialEvaluator(newPeerDenialEvaluator)

	assert.Nil(t, err)
	//pointer testing
	assert.True(t, newPeerDenialEvaluator == cmw.peerDenialEvaluator)
	assert.True(t, newPeerDenialEvaluator == cmw.PeerDenialEvaluator())
}

//------- CheckConnectionsBlocking

func TestConnectionMonitorWrapper_CheckConnectionsBlockingShouldWork(t *testing.T) {
	t.Parallel()

	whiteListPeer := peer.ID("whitelisted")
	blackListPeer := peer.ID("blacklisted")
	closeCalled := 0
	cmw := newConnectionMonitorWrapper(
		&mock.NetworkStub{
			PeersCall: func() []peer.ID {
				return []peer.ID{whiteListPeer, blackListPeer}
			},
			ClosePeerCall: func(id peer.ID) error {
				if id == blackListPeer {
					closeCalled++
					return nil
				}
				assert.Fail(t, "should have called only the black listed peer ")

				return nil
			},
		},
		&mock.ConnectionMonitorStub{},
		&mock.PeerDenialEvaluatorStub{
			IsDeniedCalled: func(pid core.PeerID) bool {
				return bytes.Equal(core.PeerID(blackListPeer).Bytes(), pid.Bytes())
			},
		},
	)

	cmw.CheckConnectionsBlocking()
	assert.Equal(t, 1, closeCalled)
}
