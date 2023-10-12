package chainSimulator

import (
	"errors"
	"sync"

	p2p2 "github.com/multiversx/mx-chain-communication-go/p2p"
	p2pMessage "github.com/multiversx/mx-chain-communication-go/p2p/message"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/p2p"
)

var errNilSeqGenerator = errors.New("nil sequence generator")

type seqGenerator interface {
	GenerateSequence() []byte
	IsInterfaceNil() bool
}

type msgCounter struct {
	received int
	sent     int
}

type testOnlySyncedMessenger struct {
	*syncedMessenger

	mutSeenMessages sync.RWMutex
	seenMessages    map[string]*msgCounter
	uniqueMessages  map[string]struct{}
	seqGenerator    seqGenerator
}

// NewTestOnlySyncedMessenger -
func NewTestOnlySyncedMessenger(network SyncedBroadcastNetworkHandler, seqGenerator seqGenerator) (*testOnlySyncedMessenger, error) {
	if check.IfNil(seqGenerator) {
		return nil, errNilSeqGenerator
	}
	internalMessenger, err := NewSyncedMessenger(network)
	if err != nil {
		return nil, err
	}

	return &testOnlySyncedMessenger{
		syncedMessenger: internalMessenger,
		seenMessages:    make(map[string]*msgCounter),
		uniqueMessages:  make(map[string]struct{}),
		seqGenerator:    seqGenerator,
	}, nil
}

// ProcessReceivedMessage -
func (messenger *testOnlySyncedMessenger) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) error {
	messenger.mutSeenMessages.Lock()
	defer messenger.mutSeenMessages.Unlock()

	key := string(message.From()) + string(message.SeqNo())
	_, alreadySeenSameMessage := messenger.uniqueMessages[key]
	if alreadySeenSameMessage {
		return nil
	}

	messenger.uniqueMessages[key] = struct{}{}

	msgCnt, alreadySeenEquivalentMessage := messenger.seenMessages[string(message.Data())]
	if !alreadySeenEquivalentMessage {
		msgCnt = &msgCounter{
			received: 0,
			sent:     1,
		}

		messenger.seenMessages[string(message.Data())] = msgCnt

		messenger.network.Broadcast(messenger.pid, message)
	}

	msgCnt.received++

	return nil
}

// Broadcast -
func (messenger *testOnlySyncedMessenger) Broadcast(topic string, buff []byte) {
	if !messenger.HasTopic(topic) {
		return
	}

	messenger.mutSeenMessages.Lock()
	defer messenger.mutSeenMessages.Unlock()
	msgCnt, alreadySeen := messenger.seenMessages[string(buff)]
	if !alreadySeen {
		msgCnt = &msgCounter{}
		messenger.seenMessages[string(buff)] = msgCnt
	}
	if msgCnt.sent == 1 {
		return
	}

	msgCnt.sent++

	message := &p2pMessage.Message{
		FromField:            messenger.ID().Bytes(),
		DataField:            buff,
		TopicField:           topic,
		BroadcastMethodField: p2p2.Broadcast,
		SeqNoField:           messenger.seqGenerator.GenerateSequence(),
	}

	messenger.network.Broadcast(messenger.pid, message)
}

func (messenger *testOnlySyncedMessenger) getSeenMessages() map[string]*msgCounter {
	messenger.mutSeenMessages.RLock()
	defer messenger.mutSeenMessages.RUnlock()

	seenMessagesCopy := make(map[string]*msgCounter, len(messenger.seenMessages))
	for msg, counter := range messenger.seenMessages {
		counterCopy := *counter
		seenMessagesCopy[msg] = &counterCopy
	}

	return seenMessagesCopy
}
