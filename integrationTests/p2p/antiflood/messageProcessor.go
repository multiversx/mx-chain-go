package antiflood

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

type messageProcessor struct {
	numMessagesReceived uint64
	mutMessages         sync.Mutex
	messages            map[p2p.PeerID][]p2p.MessageP2P
	CountersMap         process.AntifloodProtector
}

func newMessageProcessor() *messageProcessor {
	return &messageProcessor{
		messages: make(map[p2p.PeerID][]p2p.MessageP2P),
	}
}

func (mp *messageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID, _ func(buffToSend []byte)) error {
	atomic.AddUint64(&mp.numMessagesReceived, 1)

	if mp.CountersMap != nil {
		//protect from directly connected peer
		ok := mp.CountersMap.TryIncrement(string(fromConnectedPeer))
		if !ok {
			return fmt.Errorf("system flooded")
		}

		if fromConnectedPeer != message.Peer() {
			//protect from the flooding messages that originate from the same source but come from different peers
			ok = mp.CountersMap.TryIncrement(string(message.Peer()))
			if !ok {
				return fmt.Errorf("system flooded")
			}
		}
	}

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

func (mp *messageProcessor) MessagesReceived() uint64 {
	return atomic.LoadUint64(&mp.numMessagesReceived)
}

func (mp *messageProcessor) MessagesProcessed() uint64 {
	mp.mutMessages.Lock()
	defer mp.mutMessages.Unlock()

	count := 0
	for _, msgs := range mp.messages {
		count += len(msgs)
	}

	return uint64(count)
}

func (mp *messageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
