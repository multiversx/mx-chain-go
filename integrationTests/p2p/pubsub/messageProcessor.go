package peerDisconnecting

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type messageProcessor struct {
	mutMessages sync.Mutex
	messages    map[core.PeerID][]p2p.MessageP2P
}

func newMessageProcessor() *messageProcessor {
	return &messageProcessor{
		messages: make(map[core.PeerID][]p2p.MessageP2P),
	}
}

// ProcessReceivedMessage -
func (mp *messageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	mp.messages[fromConnectedPeer] = append(mp.messages[fromConnectedPeer], message)

	return nil
}

// Messages -
func (mp *messageProcessor) Messages(pid core.PeerID) []p2p.MessageP2P {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	return mp.messages[pid]
}

// AllMessages -
func (mp *messageProcessor) AllMessages() []p2p.MessageP2P {
	result := make([]p2p.MessageP2P, 0)

	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	for _, messages := range mp.messages {
		result = append(result, messages...)
	}

	return result
}

// IsInterfaceNil -
func (mp *messageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
