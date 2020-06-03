package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

const TopicRequestSuffix = topicRequestSuffix

func MakeDiffList(
	allConnectedPeers []core.PeerID,
	excludedConnectedPeers []core.PeerID,
) []core.PeerID {

	return makeDiffList(allConnectedPeers, excludedConnectedPeers)
}

func FisherYatesShuffle(indexes []int, randomizer dataRetriever.IntRandomizer) []int {
	return fisherYatesShuffle(indexes, randomizer)
}

func (dplc *DiffPeerListCreator) MainTopic() string {
	return dplc.mainTopic
}

func (dplc *DiffPeerListCreator) ExcludedPeersOnTopic() string {
	return dplc.excludePeersFromTopic
}
