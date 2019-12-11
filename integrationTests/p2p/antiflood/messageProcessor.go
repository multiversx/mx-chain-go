package antiflood

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

type messageProcessor struct {
	numMessagesProcessed  uint32
	sizeMessagesProcessed uint64

	numMessagesReceived  uint32
	sizeMessagesReceived uint64

	mutMessages    sync.Mutex
	messages       map[p2p.PeerID][]p2p.MessageP2P
	floodPreventer process.FloodPreventer
}

func newMessageProcessor() *messageProcessor {
	return &messageProcessor{
		messages: make(map[p2p.PeerID][]p2p.MessageP2P),
	}
}

// ProcessReceivedMessage is the callback function from the p2p side whenever a new message is received
func (mp *messageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	atomic.AddUint32(&mp.numMessagesReceived, 1)
	atomic.AddUint64(&mp.sizeMessagesReceived, uint64(len(message.Data())))

	if mp.floodPreventer != nil {
		//protect from directly connected peer
		ok := mp.floodPreventer.Increment(string(fromConnectedPeer), uint64(len(message.Data())))
		if !ok {
			return fmt.Errorf("system flooded")
		}

		if fromConnectedPeer != message.Peer() {
			//protect from the flooding messages that originate from the same source but come from different peers
			ok = mp.floodPreventer.Increment(string(message.Peer()), uint64(len(message.Data())))
			if !ok {
				return fmt.Errorf("system flooded")
			}
		}
	}

	atomic.AddUint32(&mp.numMessagesProcessed, 1)
	atomic.AddUint64(&mp.sizeMessagesProcessed, uint64(len(message.Data())))

	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	mp.messages[fromConnectedPeer] = append(mp.messages[fromConnectedPeer], message)

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

// IsInterfaceNil returns true if there is no value under the interface
func (mp *messageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
