package topicSender_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/topicSender"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgTopicResolverSender() topicSender.ArgTopicResolverSender {
	return topicSender.ArgTopicResolverSender{
		ArgBaseTopicSender: topicSender.ArgBaseTopicSender{
			Messenger:         &mock.MessageHandlerStub{},
			TopicName:         "topic",
			OutputAntiflooder: &mock.P2PAntifloodHandlerStub{},
			PreferredPeersHolder: &p2pmocks.PeersHolderStub{
				GetCalled: func() map[uint32][]core.PeerID {
					return map[uint32][]core.PeerID{}
				},
			},
			TargetShardId: 0,
		},
	}
}

//------- NewTopicResolverSender

func TestNewTopicResolverSender_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Messenger = nil
	trs, err := topicSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewTopicResolverSender_NilOutputAntiflooderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.OutputAntiflooder = nil
	trs, err := topicSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
}

func TestNewTopicResolverSender_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PreferredPeersHolder = nil
	trs, err := topicSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilPreferredPeersHolder, err)
}

func TestNewTopicResolverSender_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, err := topicSender.NewTopicResolverSender(arg)

	assert.False(t, check.IfNil(trs))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), trs.TargetShardID())
}

//------- Send

func TestTopicResolverSender_SendOutputAntiflooderErrorsShouldNotSendButError(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	buffToSend := []byte("buff")

	expectedErr := errors.New("can not send to peer")
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
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
	trs, _ := topicSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestTopicResolverSender_SendShouldNotCheckAntifloodForPreferred(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	buffToSend := []byte("buff")
	sendWasCalled := false

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
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
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		ContainsCalled: func(peerID core.PeerID) bool {
			return peerID == pID1
		},
	}
	trs, _ := topicSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)
	require.NoError(t, err)
	require.True(t, sendWasCalled)
}

func TestTopicResolverSender_SendShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	sentToPid1 := false
	buffToSend := []byte("buff")

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pID1.Bytes()) &&
				bytes.Equal(buff, buffToSend) {
				sentToPid1 = true
			}

			return nil
		},
	}
	trs, _ := topicSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

func TestTopicResolverSender_Topic(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicSender.NewTopicResolverSender(arg)

	assert.Equal(t, arg.TopicName+topicSender.TopicRequestSuffix, trs.RequestTopic())
}

func TestTopicResolverSender_ResolverDebugHandler(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicSender.NewTopicResolverSender(arg)

	handler := &mock.ResolverDebugHandler{}

	err := trs.SetResolverDebugHandler(handler)
	assert.Nil(t, err)

	assert.True(t, handler == trs.ResolverDebugHandler()) //pointer testing
}

func TestTopicResolverSender_SetResolverDebugHandlerNilShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicSender.NewTopicResolverSender(arg)

	err := trs.SetResolverDebugHandler(nil)
	assert.Equal(t, dataRetriever.ErrNilResolverDebugHandler, err)
}
