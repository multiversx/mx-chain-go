package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func SelectRandomPeers(connectedPeers []p2p.PeerID, peersToSend int, randomizer dataRetriever.IntRandomizer) []p2p.PeerID {
	return selectRandomPeers(connectedPeers, peersToSend, randomizer)
}
