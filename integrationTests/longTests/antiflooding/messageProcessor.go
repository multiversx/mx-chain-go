package antiflooding

import (
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

const millisecondsInSecond = 1000

type messageProcessor struct {
	antiflooder                     process.P2PAntifloodHandler
	messenger                       p2p.Messenger
	sizeMessagesProcessed           uint64
	sizeMessagesReceived            uint64
	sizeMessagesReceivedPerInterval uint64
	numMessagesProcessed            uint32
	numMessagesReceivedPerInterval  uint32
	numMessagesReceived             uint32
}

// NewMessageProcessor creates a new p2p message processor implementation used in the long test
func NewMessageProcessor(antiflooder process.P2PAntifloodHandler, messenger p2p.Messenger) *messageProcessor {
	return &messageProcessor{
		antiflooder: antiflooder,
		messenger:   messenger,
	}
}

// ProcessReceivedMessage is the callback function from the p2p side whenever a new message is received
func (mp *messageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, _ p2p.MessageHandler) error {
	atomic.AddUint32(&mp.numMessagesReceived, 1)
	atomic.AddUint64(&mp.sizeMessagesReceived, uint64(len(message.Data())))
	atomic.AddUint32(&mp.numMessagesReceivedPerInterval, 1)
	atomic.AddUint64(&mp.sizeMessagesReceivedPerInterval, uint64(len(message.Data())))

	err := mp.antiflooder.CanProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	atomic.AddUint32(&mp.numMessagesProcessed, 1)
	atomic.AddUint64(&mp.sizeMessagesProcessed, uint64(len(message.Data())))

	return nil
}

// NumMessagesProcessed returns the number of processed messages
func (mp *messageProcessor) NumMessagesProcessed() uint32 {
	return atomic.LoadUint32(&mp.numMessagesProcessed)
}

// SizeMessagesProcessed returns the total size of the processed messages
func (mp *messageProcessor) SizeMessagesProcessed() uint64 {
	return atomic.LoadUint64(&mp.sizeMessagesProcessed)
}

// NumMessagesReceived returns the number of received messages
func (mp *messageProcessor) NumMessagesReceived() uint32 {
	return atomic.LoadUint32(&mp.numMessagesReceived)
}

// SizeMessagesReceived returns the total size of the received messages
func (mp *messageProcessor) SizeMessagesReceived() uint64 {
	return atomic.LoadUint64(&mp.sizeMessagesReceived)
}

// Messenger returns the associated messenger
func (mp *messageProcessor) Messenger() p2p.Messenger {
	return mp.messenger
}

// NumMessagesReceivedPerInterval returns the average number of messages per second
func (mp *messageProcessor) NumMessagesReceivedPerInterval(basetime time.Duration) uint32 {
	if basetime < time.Millisecond {
		return 0
	}

	num := atomic.SwapUint32(&mp.numMessagesReceivedPerInterval, 0)
	divider := basetime / time.Millisecond
	return num * millisecondsInSecond / uint32(divider)
}

// SizeMessagesReceivedPerInterval returns the average number of bytes per second
func (mp *messageProcessor) SizeMessagesReceivedPerInterval(basetime time.Duration) uint64 {
	if basetime < time.Millisecond {
		return 0
	}

	num := atomic.SwapUint64(&mp.sizeMessagesReceivedPerInterval, 0)
	divider := basetime / time.Millisecond
	return num * millisecondsInSecond / uint64(divider)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *messageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
