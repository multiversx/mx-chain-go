package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func MakeDiffList(
	allConnectedPeers []p2p.PeerID,
	excludedConnectedPeers []p2p.PeerID,
) []p2p.PeerID {

	return makeDiffList(allConnectedPeers, excludedConnectedPeers)
}

func FisherYatesShuffle(indexes []int, randomizer dataRetriever.IntRandomizer) ([]int, error) {
	return fisherYatesShuffle(indexes, randomizer)
}

func (dplc *DiffPeerListCreator) MainTopic() string {
	return dplc.mainTopic
}

func (dplc *DiffPeerListCreator) ExcludedPeersOnTopic() string {
	return dplc.excludePeersFromTopic
}
