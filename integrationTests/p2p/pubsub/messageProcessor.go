package peerDisconnecting

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type messageProcessor struct {
	mutMessages sync.Mutex
	messages    map[core.PeerID][]core.MessageP2P
}

func newMessageProcessor() *messageProcessor {
	return &messageProcessor{
		messages: make(map[core.PeerID][]core.MessageP2P),
	}
}

// ProcessReceivedMessage -
func (mp *messageProcessor) ProcessReceivedMessage(message core.MessageP2P, fromConnectedPeer core.PeerID) error {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	mp.messages[fromConnectedPeer] = append(mp.messages[fromConnectedPeer], message)

	return nil
}

// Messages -
func (mp *messageProcessor) Messages(pid core.PeerID) []core.MessageP2P {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	return mp.messages[pid]
}

// AllMessages -
func (mp *messageProcessor) AllMessages() []core.MessageP2P {
	result := make([]core.MessageP2P, 0)

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
