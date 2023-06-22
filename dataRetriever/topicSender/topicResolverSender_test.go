package topicsender_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/topicSender"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgTopicResolverSender() topicsender.ArgTopicResolverSender {
	return topicsender.ArgTopicResolverSender{
		ArgBaseTopicSender: createMockArgBaseTopicSender(),
	}
}

func TestNewTopicResolverSender_NilMainMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.MainMessenger = nil
	trs, err := topicsender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewTopicResolverSender_NilFullArchiveMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.FullArchiveMessenger = nil
	trs, err := topicsender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewTopicResolverSender_NilOutputAntiflooderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.OutputAntiflooder = nil
	trs, err := topicsender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
}

func TestNewTopicResolverSender_NilMainPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.MainPreferredPeersHolder = nil
	trs, err := topicsender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewTopicResolverSender_NilFullArchivePreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.FullArchivePreferredPeersHolder = nil
	trs, err := topicsender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewTopicResolverSender_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, err := topicsender.NewTopicResolverSender(arg)

	assert.False(t, check.IfNil(trs))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), trs.TargetShardID())
}

func TestTopicResolverSender_SendOutputAntiflooderErrorsShouldNotSendButError(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	buffToSend := []byte("buff")

	expectedErr := errors.New("can not send to peer")
	arg := createMockArgTopicResolverSender()
	arg.MainMessenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			assert.Fail(t, "send shouldn't have been called")

			return nil
		},
	}
	arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			if fromConnectedPeer == pID1 {
				return expectedErr
			}

			assert.Fail(t, "wrong peer provided, should have been called with the destination peer")
			return nil
		},
	}
	trs, _ := topicsender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1, arg.MainMessenger)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestTopicResolverSender_SendShouldNotCheckAntifloodForPreferred(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	buffToSend := []byte("buff")
	sendWasCalled := false

	arg := createMockArgTopicResolverSender()
	arg.MainMessenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			require.Fail(t, "CanProcessMessage should have not be called for preferred peer")

			return nil
		},
	}
	arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
		ContainsCalled: func(peerID core.PeerID) bool {
			return peerID == pID1
		},
	}
	trs, _ := topicsender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1, arg.MainMessenger)
	require.NoError(t, err)
	require.True(t, sendWasCalled)
}

func TestTopicResolverSender_SendShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	sentToPid1 := false
	buffToSend := []byte("buff")
	t.Run("on main network", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicResolverSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pID1.Bytes()) &&
					bytes.Equal(buff, buffToSend) {
					sentToPid1 = true
				}

				return nil
			},
			TypeCalled: func() p2p.MessageHandlerType {
				return p2p.RegularMessageHandler
			},
		}
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			IsConnectedCalled: func(peerID core.PeerID) bool {
				return false
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Fail(t, "should have not been called")

				return nil
			},
			TypeCalled: func() p2p.MessageHandlerType {
				return p2p.FullArchiveMessageHandler
			},
		}
		wasCalled := false
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			ContainsCalled: func(peerID core.PeerID) bool {
				wasCalled = true
				return false
			},
		}
		arg.FullArchivePreferredPeersHolder = &p2pmocks.PeersHolderStub{
			ContainsCalled: func(peerID core.PeerID) bool {
				assert.Fail(t, "should have not been called")

				return false
			},
		}
		trs, _ := topicsender.NewTopicResolverSender(arg)

		err := trs.Send(buffToSend, pID1, arg.MainMessenger)

		assert.Nil(t, err)
		assert.True(t, sentToPid1)
		assert.True(t, wasCalled)
	})
	t.Run("on full archive network", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicResolverSender()
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			IsConnectedCalled: func(peerID core.PeerID) bool {
				return true
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pID1.Bytes()) &&
					bytes.Equal(buff, buffToSend) {
					sentToPid1 = true
				}

				return nil
			},
			TypeCalled: func() p2p.MessageHandlerType {
				return p2p.FullArchiveMessageHandler
			},
		}
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Fail(t, "should have not been called")

				return nil
			},
			TypeCalled: func() p2p.MessageHandlerType {
				return p2p.RegularMessageHandler
			},
		}
		wasCalled := false
		arg.FullArchivePreferredPeersHolder = &p2pmocks.PeersHolderStub{
			ContainsCalled: func(peerID core.PeerID) bool {
				wasCalled = true
				return false
			},
		}
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			ContainsCalled: func(peerID core.PeerID) bool {
				assert.Fail(t, "should have not been called")

				return false
			},
		}
		trs, _ := topicsender.NewTopicResolverSender(arg)

		err := trs.Send(buffToSend, pID1, arg.FullArchiveMessenger)

		assert.Nil(t, err)
		assert.True(t, sentToPid1)
		assert.True(t, wasCalled)
	})
}

func TestTopicResolverSender_Topic(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicsender.NewTopicResolverSender(arg)

	assert.Equal(t, arg.TopicName+core.TopicRequestSuffix, trs.RequestTopic())
}

func TestTopicResolverSender_DebugHandler(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicsender.NewTopicResolverSender(arg)

	handler := &mock.DebugHandler{}

	err := trs.SetDebugHandler(handler)
	assert.Nil(t, err)

	assert.True(t, handler == trs.DebugHandler()) // pointer testing
}

func TestTopicResolverSender_SetDebugHandlerNilShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicsender.NewTopicResolverSender(arg)

	err := trs.SetDebugHandler(nil)
	assert.Equal(t, dataRetriever.ErrNilDebugHandler, err)
}
