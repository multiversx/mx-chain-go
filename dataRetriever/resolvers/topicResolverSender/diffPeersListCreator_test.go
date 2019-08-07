package topicResolverSender_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNewDiffPeersListCreator_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	dplc, err := topicResolverSender.NewDiffPeersListCreator(nil, "", "")

	assert.Nil(t, dplc)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewDiffPeersListCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	mainTopic := "mainTopic"
	excludedTopic := "excludedTopic"
	dplc, err := topicResolverSender.NewDiffPeersListCreator(
		&mock.MessageHandlerStub{},
		mainTopic,
		excludedTopic,
	)

	assert.Nil(t, err)
	assert.NotNil(t, dplc)
	assert.Equal(t, mainTopic, dplc.MainTopic())
	assert.Equal(t, excludedTopic, dplc.ExcludedPeersOnTopic())
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

//------- PeersList

func TestDiffPeersListCreator_PeersListEmptyMainListShouldRetEmpty(t *testing.T) {
	t.Parallel()

	mainTopic := "mainTopic"
	excludedTopic := "excludedTopic"
	dplc, _ := topicResolverSender.NewDiffPeersListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []p2p.PeerID {
				return make([]p2p.PeerID, 0)
			},
		},
		mainTopic,
		excludedTopic,
	)

	assert.Empty(t, dplc.PeersList())
}

func TestDiffPeersListCreator_PeersListNoExcludedTopicSetShouldRetPeersOnMain(t *testing.T) {
	t.Parallel()

	mainTopic := "mainTopic"
	excludedTopic := ""
	pID1 := p2p.PeerID("peer1")
	pID2 := p2p.PeerID("peer2")
	peersOnMain := []p2p.PeerID{pID1, pID2}
	dplc, _ := topicResolverSender.NewDiffPeersListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []p2p.PeerID {
				return peersOnMain
			},
		},
		mainTopic,
		excludedTopic,
	)

	assert.Equal(t, peersOnMain, dplc.PeersList())
}

func TestDiffPeersListCreator_PeersListDiffShouldWork(t *testing.T) {
	t.Parallel()

	mainTopic := "mainTopic"
	excludedTopic := "excludedTopic"
	pID1 := p2p.PeerID("peer1")
	pID2 := p2p.PeerID("peer2")
	pID3 := p2p.PeerID("peer3")
	peersOnMain := []p2p.PeerID{pID1, pID2}
	peersOnExcluded := []p2p.PeerID{pID2, pID3}
	dplc, _ := topicResolverSender.NewDiffPeersListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []p2p.PeerID {
				switch topic {
				case mainTopic:
					return peersOnMain
				case excludedTopic:
					return peersOnExcluded
				}

				return make([]p2p.PeerID, 0)
			},
		},
		mainTopic,
		excludedTopic,
	)

	resultingList := dplc.PeersList()

	assert.Equal(t, 1, len(resultingList))
	assert.Equal(t, pID1, resultingList[0])
}

func TestDiffPeersListCreator_PeersListNoDifferenceShouldReturnMain(t *testing.T) {
	t.Parallel()

	mainTopic := "mainTopic"
	excludedTopic := "excludedTopic"
	pID1 := p2p.PeerID("peer1")
	pID2 := p2p.PeerID("peer2")
	peersOnMain := []p2p.PeerID{pID1, pID2}
	peersOnExcluded := []p2p.PeerID{pID1, pID2}
	dplc, _ := topicResolverSender.NewDiffPeersListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []p2p.PeerID {
				switch topic {
				case mainTopic:
					return peersOnMain
				case excludedTopic:
					return peersOnExcluded
				}

				return make([]p2p.PeerID, 0)
			},
		},
		mainTopic,
		excludedTopic,
	)

	resultingList := dplc.PeersList()

	assert.Equal(t, peersOnMain, resultingList)
}
