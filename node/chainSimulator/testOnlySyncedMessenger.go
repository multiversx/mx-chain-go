package chainSimulator

import (
	"encoding/binary"
	"errors"
	"strconv"
	"sync"

	"github.com/mr-tron/base58"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	p2pMessage "github.com/multiversx/mx-chain-communication-go/p2p/message"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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

	mutSeenMessages            sync.RWMutex
	seenMessages               map[string]*msgCounter
	uniqueMessages             map[string]int
	seqGenerator               seqGenerator
	isMalicious                bool
	equivalentMessagesFilterOn bool
}

// NewTestOnlySyncedMessenger -
func NewTestOnlySyncedMessenger(network SyncedBroadcastNetworkHandler, seqGenerator seqGenerator, isMalicious bool, equivalentMessagesFilterOn bool) (*testOnlySyncedMessenger, error) {
	if check.IfNil(seqGenerator) {
		return nil, errNilSeqGenerator
	}
	internalMessenger, err := NewSyncedMessenger(network)
	if err != nil {
		return nil, err
	}

	return &testOnlySyncedMessenger{
		syncedMessenger:            internalMessenger,
		seenMessages:               make(map[string]*msgCounter),
		uniqueMessages:             make(map[string]int),
		seqGenerator:               seqGenerator,
		isMalicious:                isMalicious,
		equivalentMessagesFilterOn: equivalentMessagesFilterOn,
	}, nil
}

// ProcessReceivedMessage -
func (messenger *testOnlySyncedMessenger) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) error {
	messenger.mutSeenMessages.Lock()
	defer messenger.mutSeenMessages.Unlock()

	key := base58.Encode(message.From()) + strconv.FormatUint(binary.BigEndian.Uint64(message.SeqNo()), 10)
	_, alreadySeenSameMessage := messenger.uniqueMessages[key]
	messenger.uniqueMessages[key]++
	if alreadySeenSameMessage {
		return nil
	}

	// filter for equivalent messages is not active(old behaviour)
	if !messenger.equivalentMessagesFilterOn {
		msgCnt, alreadySeenEquivalentMessage := messenger.seenMessages[string(message.Data())]
		if !alreadySeenEquivalentMessage {
			msgCnt = &msgCounter{}
			messenger.seenMessages[string(message.Data())] = msgCnt
		}

		if !messenger.isMalicious {
			msgCnt.sent++
			messenger.network.Broadcast(messenger.pid, message)
		}

		msgCnt.received++

		return nil
	}

	// filter for equivalent messages is active
	msgCnt, alreadySeenEquivalentMessage := messenger.seenMessages[string(message.Data())]
	if !alreadySeenEquivalentMessage {
		msgCnt = &msgCounter{}
		messenger.seenMessages[string(message.Data())] = msgCnt

		if !messenger.isMalicious {
			msgCnt.sent++
			messenger.network.Broadcast(messenger.pid, message)
		}
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

	if !messenger.isMalicious {
		msgCnt.sent++

		message := &p2pMessage.Message{
			FromField:            messenger.ID().Bytes(),
			DataField:            buff,
			TopicField:           topic,
			BroadcastMethodField: p2p.Broadcast,
			SeqNoField:           messenger.seqGenerator.GenerateSequence(),
		}
		messenger.network.Broadcast(messenger.pid, message)
	}
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

func (messenger *testOnlySyncedMessenger) getUniqueMessages() map[string]int {
	messenger.mutSeenMessages.RLock()
	defer messenger.mutSeenMessages.RUnlock()

	uniqueMessagesCopy := make(map[string]int, len(messenger.uniqueMessages))
	for key, counter := range messenger.uniqueMessages {
		uniqueMessagesCopy[key] = counter
	}

	return uniqueMessagesCopy
}
