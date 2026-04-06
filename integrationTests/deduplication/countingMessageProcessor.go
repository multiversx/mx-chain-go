package deduplication

import (
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("api/gin")

type countingMessageProcessor struct {
	messenger           p2p.Messenger
	numMessagesReceived uint32
	messengerName       string
}

// NewCountingMessageProcessor creates a new countingMessageProcessor instance
func NewCountingMessageProcessor(messenger p2p.Messenger, messengerName string) *countingMessageProcessor {
	return &countingMessageProcessor{
		messenger:     messenger,
		messengerName: messengerName,
	}
}

// ProcessReceivedMessage is the callback function from the p2p side whenever a new message is received
func (mp *countingMessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
	num := atomic.LoadUint32(&mp.numMessagesReceived) + 1
	atomic.AddUint32(&mp.numMessagesReceived, 1)
	log.Debug("countingMessageProcessor: ", "name:", mp.messengerName, "received message on topic:", message.Topic(), "total count:", num)
	return nil, nil
}

// NumMessagesReceived returns the number of received messages
func (mp *countingMessageProcessor) NumMessagesReceived() uint32 {
	return atomic.LoadUint32(&mp.numMessagesReceived)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cmp *countingMessageProcessor) IsInterfaceNil() bool {
	return cmp == nil
}
