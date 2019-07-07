package topicResolverSender_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

//------- NewTopicResolverSender

func TestNewTopicResolverSender_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(
		nil,
		"topic",
		"",
		&mock.MarshalizerMock{},
		&mock.IntRandomizerMock{},
	)

	assert.Nil(t, trs)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewTopicResolverSender_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{},
		"topic",
		"",
		nil,
		&mock.IntRandomizerMock{},
	)

	assert.Nil(t, trs)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewTopicResolverSender_NilRandomizerShouldErr(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{},
		"topic",
		"",
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Nil(t, trs)
	assert.Equal(t, dataRetriever.ErrNilRandomizer, err)
}

func TestNewTopicResolverSender_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{},
		"topic",
		"",
		&mock.MarshalizerMock{},
		&mock.IntRandomizerMock{},
	)

	assert.NotNil(t, trs)
	assert.Nil(t, err)
}

//------- SendOnRequestTopic

func TestTopicResolverSender_SendOnRequestTopicMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	trs, _ := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{},
		"topic",
		"",
		&mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, errExpected
			},
		},
		&mock.IntRandomizerMock{},
	)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.Equal(t, errExpected, err)
}

func TestTopicResolverSender_SendOnRequestTopicNoOneToSendShouldErr(t *testing.T) {
	t.Parallel()

	trs, _ := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []p2p.PeerID {
				return make([]p2p.PeerID, 0)
			},
		},
		"topic",
		"",
		&mock.MarshalizerMock{},
		&mock.IntRandomizerMock{},
	)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.Equal(t, dataRetriever.ErrNoConnectedPeerToSendRequest, err)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := p2p.PeerID("peer1")
	sentToPid1 := false

	trs, _ := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []p2p.PeerID {
				return []p2p.PeerID{pID1}
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pID1.Bytes()) {
					sentToPid1 = true
				}

				return nil
			},
		},
		"topic",
		"",
		&mock.MarshalizerMock{},
		&mock.IntRandomizerMock{},
	)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{})

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

//------- Send

func TestTopicResolverSender_SendShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := p2p.PeerID("peer1")
	sentToPid1 := false
	buffToSend := []byte("buff")

	trs, _ := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pID1.Bytes()) &&
					bytes.Equal(buff, buffToSend) {
					sentToPid1 = true
				}

				return nil
			},
		},
		"topic",
		"",
		&mock.MarshalizerMock{},
		&mock.IntRandomizerMock{},
	)

	err := trs.Send(buffToSend, pID1)

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

// ------- MakeDiffList

func TestMakeDiffList_EmptyExcludedShoudRetAllPeersList(t *testing.T) {
	t.Parallel()

	allPeers := []p2p.PeerID{p2p.PeerID("peer1"), p2p.PeerID("peer2")}
	excludedPeerList := make([]p2p.PeerID, 0)
	diff := topicResolverSender.MakeDiffList(allPeers, excludedPeerList)

	assert.Equal(t, allPeers, diff)
}

func TestMakeDiffList_AllFoundInExcludedShouldRetEmpty(t *testing.T) {
	t.Parallel()

	allPeers := []p2p.PeerID{p2p.PeerID("peer1"), p2p.PeerID("peer2")}
	excluded := make([]p2p.PeerID, len(allPeers))
	copy(excluded, allPeers)

	diff := topicResolverSender.MakeDiffList(allPeers, excluded)

	assert.Empty(t, diff)
}

func TestMakeDiffList_SomeFoundInExcludedShouldRetTheDifference(t *testing.T) {
	t.Parallel()

	allPeers := []p2p.PeerID{p2p.PeerID("peer1"), p2p.PeerID("peer2")}
	excluded := []p2p.PeerID{p2p.PeerID("peer1"), p2p.PeerID("peer3")}

	diff := topicResolverSender.MakeDiffList(allPeers, excluded)

	assert.Equal(t, 1, len(diff))
	assert.Equal(t, allPeers[1], diff[0])
}

func TestMakeDiffList_NoneFoundInExcludedShouldRetAllPeers(t *testing.T) {
	t.Parallel()

	allPeers := []p2p.PeerID{p2p.PeerID("peer1"), p2p.PeerID("peer2")}
	excluded := []p2p.PeerID{p2p.PeerID("peer3"), p2p.PeerID("peer4")}

	diff := topicResolverSender.MakeDiffList(allPeers, excluded)

	assert.Equal(t, allPeers, diff)
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
