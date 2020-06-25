package antiflood

import (
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	antiflood2 "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
)

// MessageProcessor -
type MessageProcessor struct {
	numMessagesProcessed  uint32
	numMessagesReceived   uint32
	sizeMessagesProcessed uint64
	sizeMessagesReceived  uint64

	mutMessages    sync.Mutex
	messages       map[core.PeerID][]p2p.MessageP2P
	FloodPreventer process.FloodPreventer
}

func newMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		messages: make(map[core.PeerID][]p2p.MessageP2P),
	}
}

// ProcessReceivedMessage is the callback function from the p2p side whenever a new message is received
func (mp *MessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	atomic.AddUint32(&mp.numMessagesReceived, 1)
	atomic.AddUint64(&mp.sizeMessagesReceived, uint64(len(message.Data())))

	if mp.FloodPreventer != nil {
		af, _ := antiflood2.NewP2PAntiflood(&mock.PeerBlackListCacherStub{}, &mock.TopicAntiFloodStub{}, mp.FloodPreventer)
		err := af.CanProcessMessage(message, fromConnectedPeer)
		if err != nil {
			return err
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
func (mp *MessageProcessor) NumMessagesProcessed() uint32 {
	return atomic.LoadUint32(&mp.numMessagesProcessed)
}

// SizeMessagesProcessed returns the total size of the processed messages
func (mp *MessageProcessor) SizeMessagesProcessed() uint64 {
	return atomic.LoadUint64(&mp.sizeMessagesProcessed)
}

// NumMessagesReceived returns the number of received messages
func (mp *MessageProcessor) NumMessagesReceived() uint32 {
	return atomic.LoadUint32(&mp.numMessagesReceived)
}

// SizeMessagesReceived returns the total size of the received messages
func (mp *MessageProcessor) SizeMessagesReceived() uint64 {
	return atomic.LoadUint64(&mp.sizeMessagesReceived)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *MessageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
