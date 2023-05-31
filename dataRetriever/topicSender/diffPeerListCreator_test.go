package topicsender_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/topicSender"
	"github.com/stretchr/testify/assert"
)

const mainTopic = "mainTopic"
const intraTopic = "intraTopic"
const excludedTopic = "excluded"
const emptyTopic = ""

func TestNewDiffPeerListCreator_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	dplc, err := topicsender.NewDiffPeerListCreator(
		nil,
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	assert.True(t, check.IfNil(dplc))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewDiffPeerListCreator_EmptyMainTopicShouldErr(t *testing.T) {
	t.Parallel()

	dplc, err := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{},
		emptyTopic,
		intraTopic,
		excludedTopic,
	)

	assert.True(t, check.IfNil(dplc))
	assert.True(t, errors.Is(err, dataRetriever.ErrEmptyString))
}

func TestNewDiffPeerListCreator_EmptyIntraTopicShouldErr(t *testing.T) {
	t.Parallel()

	dplc, err := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{},
		mainTopic,
		emptyTopic,
		excludedTopic,
	)

	assert.True(t, check.IfNil(dplc))
	assert.True(t, errors.Is(err, dataRetriever.ErrEmptyString))
}

func TestNewDiffPeerListCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	dplc, err := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{},
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(dplc))
	assert.Equal(t, mainTopic, dplc.MainTopic())
	assert.Equal(t, excludedTopic, dplc.ExcludedPeersOnTopic())
}

func TestMakeDiffList_EmptyExcludedShouldRetAllPeersList(t *testing.T) {
	t.Parallel()

	allPeers := []core.PeerID{core.PeerID("peer1"), core.PeerID("peer2")}
	excludedPeerList := make([]core.PeerID, 0)
	diff := topicsender.MakeDiffList(allPeers, excludedPeerList)

	assert.Equal(t, allPeers, diff)
}

func TestMakeDiffList_AllFoundInExcludedShouldRetEmpty(t *testing.T) {
	t.Parallel()

	allPeers := []core.PeerID{core.PeerID("peer1"), core.PeerID("peer2")}
	excluded := make([]core.PeerID, len(allPeers))
	copy(excluded, allPeers)

	diff := topicsender.MakeDiffList(allPeers, excluded)

	assert.Empty(t, diff)
}

func TestMakeDiffList_SomeFoundInExcludedShouldRetTheDifference(t *testing.T) {
	t.Parallel()

	allPeers := []core.PeerID{core.PeerID("peer1"), core.PeerID("peer2")}
	excluded := []core.PeerID{core.PeerID("peer1"), core.PeerID("peer3")}

	diff := topicsender.MakeDiffList(allPeers, excluded)

	assert.Equal(t, 1, len(diff))
	assert.Equal(t, allPeers[1], diff[0])
}

func TestMakeDiffList_NoneFoundInExcludedShouldRetAllPeers(t *testing.T) {
	t.Parallel()

	allPeers := []core.PeerID{core.PeerID("peer1"), core.PeerID("peer2")}
	excluded := []core.PeerID{core.PeerID("peer3"), core.PeerID("peer4")}

	diff := topicsender.MakeDiffList(allPeers, excluded)

	assert.Equal(t, allPeers, diff)
}

func TestDiffPeerListCreator_CrossShardPeersListEmptyMainListShouldRetEmpty(t *testing.T) {
	t.Parallel()

	dplc, _ := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				return make([]core.PeerID, 0)
			},
		},
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	assert.Empty(t, dplc.CrossShardPeerList())
}

func TestDiffPeerListCreator_CrossShardPeersListNoExcludedTopicSetShouldRetPeersOnMain(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	pID2 := core.PeerID("peer2")
	peersOnMain := []core.PeerID{pID1, pID2}
	dplc, _ := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				return peersOnMain
			},
		},
		mainTopic,
		intraTopic,
		emptyTopic,
	)

	assert.Equal(t, peersOnMain, dplc.CrossShardPeerList())
}

func TestDiffPeerListCreator_CrossShardPeersListDiffShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	pID2 := core.PeerID("peer2")
	pID3 := core.PeerID("peer3")
	peersOnMain := []core.PeerID{pID1, pID2}
	peersOnExcluded := []core.PeerID{pID2, pID3}
	dplc, _ := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				switch topic {
				case mainTopic:
					return peersOnMain
				case excludedTopic:
					return peersOnExcluded
				}

				return make([]core.PeerID, 0)
			},
		},
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	resultingList := dplc.CrossShardPeerList()

	assert.Equal(t, 1, len(resultingList))
	assert.Equal(t, pID1, resultingList[0])
}

func TestDiffPeerListCreator_CrossShardPeersListNoDifferenceShouldReturnMain(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	pID2 := core.PeerID("peer2")
	peersOnMain := []core.PeerID{pID1, pID2}
	peersOnExcluded := []core.PeerID{pID1, pID2}
	dplc, _ := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				switch topic {
				case mainTopic:
					return peersOnMain
				case excludedTopic:
					return peersOnExcluded
				}

				return make([]core.PeerID, 0)
			},
		},
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	resultingList := dplc.CrossShardPeerList()

	assert.Equal(t, peersOnMain, resultingList)
}

func TestDiffPeerListCreator_IntraShardPeersList(t *testing.T) {
	t.Parallel()

	peerList := []core.PeerID{"pid1", "pid2"}
	dplc, _ := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				if topic == intraTopic {
					return peerList
				}

				return nil
			},
		},
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	assert.Equal(t, peerList, dplc.IntraShardPeerList())
}

func TestDiffPeerListCreator_FullHistoryList(t *testing.T) {
	t.Parallel()

	peerList := []core.PeerID{"pid1", "pid2"}
	dplc, _ := topicsender.NewDiffPeerListCreator(
		&mock.MessageHandlerStub{
			ConnectedFullHistoryPeersOnTopicCalled: func(topic string) []core.PeerID {
				return peerList
			},
		},
		mainTopic,
		intraTopic,
		excludedTopic,
	)

	assert.Equal(t, peerList, dplc.FullHistoryList())
}
