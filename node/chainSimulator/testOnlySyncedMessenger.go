package chainSimulator

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
)

type msgCounter struct {
	received int
	sent     int
}
type testOnlySyncedMessenger struct {
	*syncedMessenger

	mutSeenMessages sync.RWMutex
	seenMessages    map[string]*msgCounter
}

// NewTestOnlySyncedMessenger -
func NewTestOnlySyncedMessenger(network SyncedBroadcastNetworkHandler) (*testOnlySyncedMessenger, error) {
	internalMessenger, err := NewSyncedMessenger(network)
	if err != nil {
		return nil, err
	}

	return &testOnlySyncedMessenger{
		syncedMessenger: internalMessenger,
		seenMessages:    make(map[string]*msgCounter),
	}, nil
}

// ProcessReceivedMessage -
func (messenger *testOnlySyncedMessenger) ProcessReceivedMessage(message p2p.MessageP2P, peer core.PeerID, _ p2p.MessageHandler) error {
	messenger.mutSeenMessages.Lock()
	defer messenger.mutSeenMessages.Unlock()

	msgCnt, alreadySeen := messenger.seenMessages[string(message.Data())]

	if !alreadySeen {
		msgCnt = &msgCounter{
			received: 0,
			sent:     1,
		}
		messenger.seenMessages[string(message.Data())] = msgCnt

		messenger.network.Broadcast(messenger.pid, message.Topic(), message.Data())
	}

	msgCnt.received++

	//println(fmt.Sprintf("%s got message from %s, alreadySeen %t, sent %d times, received %d times",
	//	messenger.ID().Pretty(),
	//	peer.Pretty(),
	//	alreadySeen,
	//	msgCnt.sent,
	//	msgCnt.received,
	//))

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

	messenger.network.Broadcast(messenger.pid, topic, buff)
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
