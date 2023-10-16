package chainSimulator

import (
	"errors"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/p2p"
)

type testOnlyMessageProcessor struct {
	messenger p2p.Messenger

	mutSeenMessages sync.RWMutex
	seenMessages    map[string]*msgCounter
	isMalicious     bool
}

// NewTestOnlyMessageProcessor -
func NewTestOnlyMessageProcessor(messenger p2p.Messenger, isMalicious bool) (*testOnlyMessageProcessor, error) {
	if check.IfNil(messenger) {
		return nil, errors.New("nil messenger")
	}
	return &testOnlyMessageProcessor{
		messenger:    messenger,
		seenMessages: make(map[string]*msgCounter),
		isMalicious:  isMalicious,
	}, nil
}

// ProcessReceivedMessage -
func (proc *testOnlyMessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) error {
	proc.mutSeenMessages.Lock()
	defer proc.mutSeenMessages.Unlock()

	msgCnt, alreadySeenEquivalentMessage := proc.seenMessages[string(message.Data())]
	if !alreadySeenEquivalentMessage {
		msgCnt = &msgCounter{}
		proc.seenMessages[string(message.Data())] = msgCnt

		if !proc.isMalicious {
			msgCnt.sent++
			proc.messenger.Broadcast(message.Topic(), message.Data())
		}
	}

	msgCnt.received++

	return nil
}

// GetSeenMessages -
func (proc *testOnlyMessageProcessor) GetSeenMessages() map[string]*msgCounter {
	proc.mutSeenMessages.RLock()
	defer proc.mutSeenMessages.RUnlock()

	seenMessagesCopy := make(map[string]*msgCounter, len(proc.seenMessages))
	for msg, counter := range proc.seenMessages {
		counterCopy := *counter
		seenMessagesCopy[msg] = &counterCopy
	}

	return seenMessagesCopy
}

// IsMalicious -
func (proc *testOnlyMessageProcessor) IsMalicious() bool {
	return proc.isMalicious
}

// IsInterfaceNil -
func (proc *testOnlyMessageProcessor) IsInterfaceNil() bool {
	return proc == nil
}
