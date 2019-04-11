package topicResolverSender_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

//------- NewTopicResolverSender

func TestNewTopicResolverSender_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(nil, "topic", &mock.MarshalizerMock{})

	assert.Nil(t, trs)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewTopicResolverSender_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(&mock.MessageHandlerStub{}, "topic", nil)

	assert.Nil(t, trs)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewTopicResolverSender_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	trs, err := topicResolverSender.NewTopicResolverSender(
		&mock.MessageHandlerStub{},
		"topic",
		&mock.MarshalizerMock{})

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
		&mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, errExpected
			},
		})

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
		&mock.MarshalizerMock{},
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
		&mock.MarshalizerMock{},
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
		&mock.MarshalizerMock{},
	)

	err := trs.Send(buffToSend, pID1)

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

// ------- SelectRandomPeers

func TestSelectRandomPeers_ConnectedPeersLen0ShoudRetEmpty(t *testing.T) {
	t.Parallel()

	selectedPeers := topicResolverSender.SelectRandomPeers(make([]p2p.PeerID, 0), 0, nil)

	assert.Equal(t, 0, len(selectedPeers))
}

func TestSelectRandomPeers_ConnectedPeersLenSmallerThanRequiredShoudRetListTest1(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2")}

	selectedPeers := topicResolverSender.SelectRandomPeers(connectedPeers, 3, nil)

	assert.Equal(t, connectedPeers, selectedPeers)
}

func TestSelectRandomPeers_ConnectedPeersLenSmallerThanRequiredShoudRetListTest2(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2")}

	selectedPeers := topicResolverSender.SelectRandomPeers(connectedPeers, 2, nil)

	assert.Equal(t, connectedPeers, selectedPeers)
}

func TestSelectRandomPeers_ConnectedPeersTestRandomizerRepeat0ThreeTimes(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2"), p2p.PeerID("peer 3")}

	valuesGenerated := []int{0, 0, 0, 1}
	idxGenerated := 0

	mr := &mock.IntRandomizerMock{
		IntnCalled: func(n int) int {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val
		},
	}

	selectedPeers := topicResolverSender.SelectRandomPeers(connectedPeers, 2, mr)

	//since iterating a map does not guarantee the order, we have to search in any combination possible
	foundPeer0 := false
	foundPeer1 := false

	for i := 0; i < len(selectedPeers); i++ {
		if selectedPeers[i] == connectedPeers[0] {
			foundPeer0 = true
		}
		if selectedPeers[i] == connectedPeers[1] {
			foundPeer1 = true
		}
	}

	assert.True(t, foundPeer0 && foundPeer1)
	assert.Equal(t, 2, len(selectedPeers))
}

func TestSelectRandomPeers_ConnectedPeersTestRandomizerRepeat2TwoTimes(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2"), p2p.PeerID("peer 3")}

	valuesGenerated := []int{2, 2, 0}
	idxGenerated := 0

	mr := &mock.IntRandomizerMock{
		IntnCalled: func(n int) int {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val
		},
	}

	selectedPeers := topicResolverSender.SelectRandomPeers(connectedPeers, 2, mr)

	//since iterating a map does not guarantee the order, we have to search in any combination possible
	foundPeer0 := false
	foundPeer2 := false

	for i := 0; i < len(selectedPeers); i++ {
		if selectedPeers[i] == connectedPeers[0] {
			foundPeer0 = true
		}
		if selectedPeers[i] == connectedPeers[2] {
			foundPeer2 = true
		}
	}

	assert.True(t, foundPeer0 && foundPeer2)
	assert.Equal(t, 2, len(selectedPeers))
}
