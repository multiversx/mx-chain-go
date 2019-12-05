package peerDisconnecting

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

type messageProcessor struct {
	mutMessages sync.Mutex
	messages    map[p2p.PeerID][]p2p.MessageP2P
}

func newMessageProcessor() *messageProcessor {
	return &messageProcessor{
		messages: make(map[p2p.PeerID][]p2p.MessageP2P),
	}
}

func (mp *messageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID, _ func(buffToSend []byte)) error {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	mp.messages[fromConnectedPeer] = append(mp.messages[fromConnectedPeer], message)

	return nil
}

func (mp *messageProcessor) Messages(pid p2p.PeerID) []p2p.MessageP2P {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	return mp.messages[pid]
}

func (mp *messageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
