package topicResolverSender_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func createMockArgTopicResolverSender() topicResolverSender.ArgTopicResolverSender {
	return topicResolverSender.ArgTopicResolverSender{
		Messenger:         &mock.MessageHandlerStub{},
		TopicName:         "topic",
		PeerListCreator:   &mock.PeerListCreatorStub{},
		Marshalizer:       &mock.MarshalizerMock{},
		Randomizer:        &mock.IntRandomizerMock{},
		TargetShardId:     0,
		OutputAntiflooder: &mock.P2PAntifloodHandlerStub{},
	}
}

//------- NewTopicResolverSender

func TestNewTopicResolverSender_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Messenger = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewTopicResolverSender_NilPeersListCreatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PeerListCreator = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilPeerListCreator, err)
}

func TestNewTopicResolverSender_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Marshalizer = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewTopicResolverSender_NilRandomizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Randomizer = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilRandomizer, err)
}

func TestNewTopicResolverSender_NilOutputAntiflooderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.OutputAntiflooder = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
}

func TestNewTopicResolverSender_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.False(t, check.IfNil(trs))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), trs.TargetShardID())
}

//------- SendOnRequestTopic

func TestTopicResolverSender_SendOnRequestTopicMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	arg := createMockArgTopicResolverSender()
	arg.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return nil, errExpected
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.Equal(t, errExpected, err)
}

func TestTopicResolverSender_SendOnRequestTopicNoOneToSendShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		PeerListCalled: func() []p2p.PeerID {
			return make([]p2p.PeerID, 0)
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.Equal(t, dataRetriever.ErrNoConnectedPeerToSendRequest, err)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := p2p.PeerID("peer1")
	sentToPid1 := false

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pID1.Bytes()) {
				sentToPid1 = true
			}

			return nil
		},
	}
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		PeerListCalled: func() []p2p.PeerID {
			return []p2p.PeerID{pID1}
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

func TestTopicResolverSender_SendOnRequestTopicErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	pID1 := p2p.PeerID("peer1")
	sentToPid1 := false

	expectedErr := errors.New("expected error")
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pID1.Bytes()) {
				sentToPid1 = true
			}

			return expectedErr
		},
	}
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		PeerListCalled: func() []p2p.PeerID {
			return []p2p.PeerID{pID1}
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, sentToPid1)
}

//------- Send

func TestTopicResolverSender_SendOutputAntiflooderErrorsShouldNotSendButError(t *testing.T) {
	t.Parallel()

	pID1 := p2p.PeerID("peer1")
	buffToSend := []byte("buff")

	expectedErr := errors.New("can not send to peer")
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			assert.Fail(t, "should not have call send")

			return nil
		},
	}
	arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
			if fromConnectedPeer == pID1 {
				return expectedErr
			}

			assert.Fail(t, "wrong peer provided, should have called with the destination peer")
			return nil
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestTopicResolverSender_SendShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := p2p.PeerID("peer1")
	sentToPid1 := false
	buffToSend := []byte("buff")

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pID1.Bytes()) &&
				bytes.Equal(buff, buffToSend) {
				sentToPid1 = true
			}

			return nil
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

func TestTopicResolverSender_Topic(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	assert.Equal(t, arg.TopicName+topicResolverSender.TopicRequestSuffix, trs.Topic())
}

// ------- FisherYatesShuffle

func TestFisherYatesShuffle_EmptyShouldReturnEmpty(t *testing.T) {
	indexes := make([]int, 0)
	randomizer := &mock.IntRandomizerMock{}

	resultIndexes, err := topicResolverSender.FisherYatesShuffle(indexes, randomizer)

	assert.Nil(t, err)
	assert.Empty(t, resultIndexes)
}

func TestFisherYatesShuffle_OneElementShouldReturnTheSame(t *testing.T) {
	indexes := []int{1}
	randomizer := &mock.IntRandomizerMock{
		IntnCalled: func(n int) (i int, e error) {
			return n - 1, nil
		},
	}

	resultIndexes, err := topicResolverSender.FisherYatesShuffle(indexes, randomizer)

	assert.Nil(t, err)
	assert.Equal(t, indexes, resultIndexes)
}

func TestFisherYatesShuffle_ShouldWork(t *testing.T) {
	indexes := []int{1, 2, 3, 4, 5}
	randomizer := &mock.IntRandomizerMock{
		IntnCalled: func(n int) (i int, e error) {
			return 0, nil
		},
	}

	//this will cause a rotation of the first element:
	//i = 4: 5, 2, 3, 4, 1 (swap 1 <-> 5)
	//i = 3: 4, 2, 3, 5, 1 (swap 5 <-> 4)
	//i = 2: 3, 2, 4, 5, 1 (swap 3 <-> 4)
	//i = 1: 2, 3, 4, 5, 1 (swap 3 <-> 2)

	resultIndexes, err := topicResolverSender.FisherYatesShuffle(indexes, randomizer)
	expectedResult := []int{2, 3, 4, 5, 1}

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, resultIndexes)
}
