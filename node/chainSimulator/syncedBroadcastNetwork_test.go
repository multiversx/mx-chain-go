package chainSimulator

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func TestSyncedBroadcastNetwork_BroadcastShouldWorkOn3Peers(t *testing.T) {
	t.Parallel()

	network := NewSyncedBroadcastNetwork()
	messages := make(map[core.PeerID]map[string][]byte)

	globalTopic := "global"
	oneTwoTopic := "topic_1_2"
	oneThreeTopic := "topic_1_3"
	twoThreeTopic := "topic_2_3"

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor1 := createMessageProcessor(messages, peer1.ID())
	_ = peer1.CreateTopic(globalTopic, true)
	_ = peer1.RegisterMessageProcessor(globalTopic, "", processor1)
	_ = peer1.CreateTopic(oneTwoTopic, true)
	_ = peer1.RegisterMessageProcessor(oneTwoTopic, "", processor1)
	_ = peer1.CreateTopic(oneThreeTopic, true)
	_ = peer1.RegisterMessageProcessor(oneThreeTopic, "", processor1)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor2 := createMessageProcessor(messages, peer2.ID())
	_ = peer2.CreateTopic(globalTopic, true)
	_ = peer2.RegisterMessageProcessor(globalTopic, "", processor2)
	_ = peer2.CreateTopic(oneTwoTopic, true)
	_ = peer2.RegisterMessageProcessor(oneTwoTopic, "", processor2)
	_ = peer2.CreateTopic(twoThreeTopic, true)
	_ = peer2.RegisterMessageProcessor(twoThreeTopic, "", processor2)

	peer3, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor3 := createMessageProcessor(messages, peer3.ID())
	_ = peer3.CreateTopic(globalTopic, true)
	_ = peer3.RegisterMessageProcessor(globalTopic, "", processor3)
	_ = peer3.CreateTopic(oneThreeTopic, true)
	_ = peer3.RegisterMessageProcessor(oneThreeTopic, "", processor3)
	_ = peer3.CreateTopic(twoThreeTopic, true)
	_ = peer3.RegisterMessageProcessor(twoThreeTopic, "", processor3)

	globalMessage := []byte("global message")
	oneTwoMessage := []byte("1-2 message")
	oneThreeMessage := []byte("1-3 message")
	twoThreeMessage := []byte("2-3 message")

	peer1.Broadcast(globalTopic, globalMessage)
	assert.Equal(t, globalMessage, messages[peer1.ID()][globalTopic])
	assert.Equal(t, globalMessage, messages[peer2.ID()][globalTopic])
	assert.Equal(t, globalMessage, messages[peer3.ID()][globalTopic])

	peer1.Broadcast(oneTwoTopic, oneTwoMessage)
	assert.Equal(t, oneTwoMessage, messages[peer1.ID()][oneTwoTopic])
	assert.Equal(t, oneTwoMessage, messages[peer2.ID()][oneTwoTopic])
	assert.Nil(t, messages[peer3.ID()][oneTwoTopic])

	peer1.Broadcast(oneThreeTopic, oneThreeMessage)
	assert.Equal(t, oneThreeMessage, messages[peer1.ID()][oneThreeTopic])
	assert.Nil(t, messages[peer2.ID()][oneThreeTopic])
	assert.Equal(t, oneThreeMessage, messages[peer3.ID()][oneThreeTopic])

	peer2.Broadcast(twoThreeTopic, twoThreeMessage)
	assert.Nil(t, messages[peer1.ID()][twoThreeTopic])
	assert.Equal(t, twoThreeMessage, messages[peer2.ID()][twoThreeTopic])
	assert.Equal(t, twoThreeMessage, messages[peer3.ID()][twoThreeTopic])
}

func TestSyncedBroadcastNetwork_BroadcastOnAnUnjoinedTopicShouldDiscardMessage(t *testing.T) {
	t.Parallel()

	network := NewSyncedBroadcastNetwork()
	messages := make(map[core.PeerID]map[string][]byte)

	globalTopic := "global"
	twoThreeTopic := "topic_2_3"

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor1 := createMessageProcessor(messages, peer1.ID())
	_ = peer1.CreateTopic(globalTopic, true)
	_ = peer1.RegisterMessageProcessor(globalTopic, "", processor1)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor2 := createMessageProcessor(messages, peer2.ID())
	_ = peer2.CreateTopic(globalTopic, true)
	_ = peer2.RegisterMessageProcessor(globalTopic, "", processor2)
	_ = peer2.CreateTopic(twoThreeTopic, true)
	_ = peer2.RegisterMessageProcessor(twoThreeTopic, "", processor2)

	peer3, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor3 := createMessageProcessor(messages, peer3.ID())
	_ = peer3.CreateTopic(globalTopic, true)
	_ = peer3.RegisterMessageProcessor(globalTopic, "", processor3)
	_ = peer3.CreateTopic(twoThreeTopic, true)
	_ = peer3.RegisterMessageProcessor(twoThreeTopic, "", processor3)

	testMessage := []byte("test message")

	peer1.Broadcast(twoThreeTopic, testMessage)

	assert.Nil(t, messages[peer1.ID()][twoThreeTopic])
	assert.Nil(t, messages[peer2.ID()][twoThreeTopic])
	assert.Nil(t, messages[peer3.ID()][twoThreeTopic])
}

func TestSyncedBroadcastNetwork_SendDirectlyShouldWorkBetween2peers(t *testing.T) {
	t.Parallel()

	network := NewSyncedBroadcastNetwork()
	messages := make(map[core.PeerID]map[string][]byte)

	topic := "topic"
	testMessage := []byte("test message")

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor1 := createMessageProcessor(messages, peer1.ID())
	_ = peer1.CreateTopic(topic, true)
	_ = peer1.RegisterMessageProcessor(topic, "", processor1)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor2 := createMessageProcessor(messages, peer2.ID())
	_ = peer2.CreateTopic(topic, true)
	_ = peer2.RegisterMessageProcessor(topic, "", processor2)

	err = peer1.SendToConnectedPeer(topic, testMessage, peer2.ID())
	assert.Nil(t, err)

	assert.Nil(t, messages[peer1.ID()][topic])
	assert.Equal(t, testMessage, messages[peer2.ID()][topic])
}

func TestSyncedBroadcastNetwork_SendDirectlyToSelfShouldWork(t *testing.T) {
	t.Parallel()

	network := NewSyncedBroadcastNetwork()
	messages := make(map[core.PeerID]map[string][]byte)

	topic := "topic"
	testMessage := []byte("test message")

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor1 := createMessageProcessor(messages, peer1.ID())
	_ = peer1.CreateTopic(topic, true)
	_ = peer1.RegisterMessageProcessor(topic, "", processor1)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor2 := createMessageProcessor(messages, peer2.ID())
	_ = peer2.CreateTopic(topic, true)
	_ = peer2.RegisterMessageProcessor(topic, "", processor2)

	err = peer1.SendToConnectedPeer(topic, testMessage, peer1.ID())
	assert.Nil(t, err)

	assert.Equal(t, testMessage, messages[peer1.ID()][topic])
	assert.Nil(t, messages[peer2.ID()][topic])
}

func TestSyncedBroadcastNetwork_SendDirectlyShouldNotDeadlock(t *testing.T) {
	t.Parallel()

	network := NewSyncedBroadcastNetwork()
	messages := make(map[core.PeerID]map[string][]byte)

	topic := "topic"
	testMessage := []byte("test message")

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor1 := createMessageProcessor(messages, peer1.ID())
	_ = peer1.CreateTopic(topic, true)
	_ = peer1.RegisterMessageProcessor(topic, "", processor1)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	processor2 := &p2pmocks.MessageProcessorStub{
		ProcessReceivedMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
			log.Debug("sending message back to", "pid", fromConnectedPeer.Pretty())
			return source.SendToConnectedPeer(message.Topic(), []byte("reply: "+string(message.Data())), fromConnectedPeer)
		},
	}
	_ = peer2.CreateTopic(topic, true)
	_ = peer2.RegisterMessageProcessor(topic, "", processor2)

	err = peer1.SendToConnectedPeer(topic, testMessage, peer2.ID())
	assert.Nil(t, err)

	assert.Equal(t, "reply: "+string(testMessage), string(messages[peer1.ID()][topic]))
	assert.Nil(t, messages[peer2.ID()][topic])
}

func TestSyncedBroadcastNetwork_ConnectedPeersAndAddresses(t *testing.T) {
	t.Parallel()

	network := NewSyncedBroadcastNetwork()

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)

	peers := peer1.ConnectedPeers()
	assert.Equal(t, 2, len(peers))

	assert.Contains(t, peers, peer1.ID())
	assert.Contains(t, peers, peer2.ID())

	assert.True(t, peer1.IsConnected(peer2.ID()))
	assert.True(t, peer2.IsConnected(peer1.ID()))
	assert.False(t, peer1.IsConnected("no connection"))

	addresses := peer1.ConnectedAddresses()
	assert.Equal(t, 2, len(addresses))
	assert.Contains(t, addresses, fmt.Sprintf(virtualAddressTemplate, peer1.ID().Pretty()))
	assert.Contains(t, addresses, peer1.Addresses()[0])
	assert.Contains(t, addresses, fmt.Sprintf(virtualAddressTemplate, peer2.ID().Pretty()))
	assert.Contains(t, addresses, peer2.Addresses()[0])
}

func TestSyncedBroadcastNetwork_GetConnectedPeersOnTopic(t *testing.T) {
	t.Parallel()

	globalTopic := "global"
	oneTwoTopic := "topic_1_2"
	oneThreeTopic := "topic_1_3"
	twoThreeTopic := "topic_2_3"

	network := NewSyncedBroadcastNetwork()

	peer1, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	_ = peer1.CreateTopic(globalTopic, false)
	_ = peer1.CreateTopic(oneTwoTopic, false)
	_ = peer1.CreateTopic(oneThreeTopic, false)

	peer2, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	_ = peer2.CreateTopic(globalTopic, false)
	_ = peer2.CreateTopic(oneTwoTopic, false)
	_ = peer2.CreateTopic(twoThreeTopic, false)

	peer3, err := NewSyncedMessenger(network)
	assert.Nil(t, err)
	_ = peer3.CreateTopic(globalTopic, false)
	_ = peer3.CreateTopic(oneThreeTopic, false)
	_ = peer3.CreateTopic(twoThreeTopic, false)

	peers := peer1.ConnectedPeersOnTopic(globalTopic)
	assert.Equal(t, 3, len(peers))
	assert.Contains(t, peers, peer1.ID())
	assert.Contains(t, peers, peer2.ID())
	assert.Contains(t, peers, peer3.ID())

	peers = peer1.ConnectedPeersOnTopic(oneTwoTopic)
	assert.Equal(t, 2, len(peers))
	assert.Contains(t, peers, peer1.ID())
	assert.Contains(t, peers, peer2.ID())

	peers = peer3.ConnectedPeersOnTopic(oneThreeTopic)
	assert.Equal(t, 2, len(peers))
	assert.Contains(t, peers, peer1.ID())
	assert.Contains(t, peers, peer3.ID())

	peersInfo := peer1.GetConnectedPeersInfo()
	assert.Equal(t, 3, len(peersInfo.UnknownPeers))
}

func createMessageProcessor(dataMap map[core.PeerID]map[string][]byte, pid core.PeerID) p2p.MessageProcessor {
	return &p2pmocks.MessageProcessorStub{
		ProcessReceivedMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
			m, found := dataMap[pid]
			if !found {
				m = make(map[string][]byte)
				dataMap[pid] = m
			}

			m[message.Topic()] = message.Data()

			return nil
		},
	}
}
