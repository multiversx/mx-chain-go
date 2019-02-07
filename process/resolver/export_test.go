package resolver

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

func (tr *topicResolver) RequestValidator(message p2p.MessageP2P) error {
	return tr.requestValidator(message)
}

func SelectRandomPeers(connectedPeers []p2p.PeerID, peersToSend int, randomizer process.IntRandomizer) []p2p.PeerID {
	return selectRandomPeers(connectedPeers, peersToSend, randomizer)
}
