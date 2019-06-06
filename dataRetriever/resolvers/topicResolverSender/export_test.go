package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func SelectRandomPeers(
	connectedPeers []p2p.PeerID,
	excludedConnectedPeers []p2p.PeerID,
	peersToSend int,
	randomizer dataRetriever.IntRandomizer,
) ([]p2p.PeerID, error) {

	return selectRandomPeers(connectedPeers, excludedConnectedPeers, peersToSend, randomizer)
}

func MakeDiffList(
	allConnectedPeers []p2p.PeerID,
	excludedConnectedPeers []p2p.PeerID,
) []p2p.PeerID {

	return makeDiffList(allConnectedPeers, excludedConnectedPeers)
}
