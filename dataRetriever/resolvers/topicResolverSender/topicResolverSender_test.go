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

// ------- SelectRandomPeers

func TestSelectRandomPeers_ConnectedPeersLen0ShoudRetEmpty(t *testing.T) {
	t.Parallel()

	allPeerList := make([]p2p.PeerID, 0)
	excludedPeerList := make([]p2p.PeerID, 0)
	selectedPeers, err := topicResolverSender.SelectRandomPeers(allPeerList, excludedPeerList, 0, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(selectedPeers))
}

func TestSelectRandomPeers_ConnectedPeersLenSmallerThanRequiredShoudRetListTest1(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2")}
	excludedPeerList := make([]p2p.PeerID, 0)
	selectedPeers, err := topicResolverSender.SelectRandomPeers(connectedPeers, excludedPeerList, 3, nil)

	assert.Nil(t, err)
	assert.Equal(t, connectedPeers, selectedPeers)
}

func TestSelectRandomPeers_ConnectedPeersLenSmallerThanRequiredShoudRetListTest2(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2")}
	excludedPeerList := make([]p2p.PeerID, 0)
	selectedPeers, err := topicResolverSender.SelectRandomPeers(connectedPeers, excludedPeerList, 2, nil)

	assert.Nil(t, err)
	assert.Equal(t, connectedPeers, selectedPeers)
}

func TestSelectRandomPeers_ConnectedPeersTestRandomizerRepeat0ThreeTimes(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2"), p2p.PeerID("peer 3")}

	valuesGenerated := []int{0, 0, 0, 1}
	idxGenerated := 0

	mr := &mock.IntRandomizerMock{
		IntnCalled: func(n int) (int, error) {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val, nil
		},
	}

	excludedPeerList := make([]p2p.PeerID, 0)
	selectedPeers, _ := topicResolverSender.SelectRandomPeers(connectedPeers, excludedPeerList, 2, mr)

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
		IntnCalled: func(n int) (int, error) {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val, nil
		},
	}

	excludedPeerList := make([]p2p.PeerID, 0)
	selectedPeers, _ := topicResolverSender.SelectRandomPeers(connectedPeers, excludedPeerList, 2, mr)

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

func TestSelectRandomPeers_AllConnectedPeersIsExcludedListShouldPickFromAllConnectedList(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2"), p2p.PeerID("peer 3")}
	excludedPeers := make([]p2p.PeerID, len(connectedPeers))
	copy(excludedPeers, connectedPeers)

	valuesGenerated := []int{0, 1, 2}
	idxGenerated := 0

	mr := &mock.IntRandomizerMock{
		IntnCalled: func(n int) (int, error) {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val, nil
		},
	}

	selectedPeers, _ := topicResolverSender.SelectRandomPeers(connectedPeers, excludedPeers, 2, mr)

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
