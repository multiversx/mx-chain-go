package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MockMessageProcessor -
type MockMessageProcessor struct {
	Peer p2p.PeerID
}

// NewMockMessageProcessor -
func NewMockMessageProcessor(peer p2p.PeerID) *MockMessageProcessor {
	processor := MockMessageProcessor{}
	processor.Peer = peer
	return &processor
}

// ProcessReceivedMessage -
func (processor *MockMessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	fmt.Printf("Message received by %s from %s: %s\n", string(processor.Peer), string(message.Peer()), string(message.Data()))
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *MockMessageProcessor) IsInterfaceNil() bool {
	return processor == nil
}
