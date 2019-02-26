package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

func SelectRandomPeers(connectedPeers []p2p.PeerID, peersToSend int, randomizer process.IntRandomizer) []p2p.PeerID {
	return selectRandomPeers(connectedPeers, peersToSend, randomizer)
}
