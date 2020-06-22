package peerDiscovery

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// SimpleMessageProcessor records the last received message
type SimpleMessageProcessor struct {
	mutMessage sync.RWMutex
	message    []byte
}

// ProcessReceivedMessage records the message
func (smp *SimpleMessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	smp.mutMessage.Lock()
	smp.message = message.Data()
	smp.mutMessage.Unlock()

	return nil
}

// GetLastMessage returns the last message received
func (smp *SimpleMessageProcessor) GetLastMessage() []byte {
	smp.mutMessage.RLock()
	defer smp.mutMessage.RUnlock()

	return smp.message
}

// IsInterfaceNil returns true if there is no value under the interface
func (smp *SimpleMessageProcessor) IsInterfaceNil() bool {
	return smp == nil
}
