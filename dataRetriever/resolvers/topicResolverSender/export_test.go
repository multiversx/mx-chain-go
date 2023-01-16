package topicResolverSender

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

const TopicRequestSuffix = topicRequestSuffix

func MakeDiffList(
	allConnectedPeers []core.PeerID,
	excludedConnectedPeers []core.PeerID,
) []core.PeerID {

	return makeDiffList(allConnectedPeers, excludedConnectedPeers)
}

func (dplc *DiffPeerListCreator) MainTopic() string {
	return dplc.mainTopic
}

func (dplc *DiffPeerListCreator) ExcludedPeersOnTopic() string {
	return dplc.excludePeersFromTopic
}
