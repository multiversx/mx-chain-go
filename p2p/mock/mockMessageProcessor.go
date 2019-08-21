package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MockMessageProcessor struct {
	Peer p2p.PeerID
}

func NewMockMessageProcessor(peer p2p.PeerID) *MockMessageProcessor {
	processor := MockMessageProcessor{}
	processor.Peer = peer
	return &processor
}

func (processor *MockMessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	fmt.Printf("Message received by %s from %s: %s\n", string(processor.Peer), string(message.Peer()), string(message.Data()))
	return nil
}
