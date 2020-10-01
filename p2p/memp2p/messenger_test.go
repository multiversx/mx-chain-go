package memp2p_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/memp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestInitializingNetworkAndPeer(t *testing.T) {
	network := memp2p.NewNetwork()

	peer, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(network.Peers()))

	assert.Equal(t, 1, len(peer.Addresses()))
	assert.Equal(t, "/memp2p/"+string(peer.ID()), peer.Addresses()[0])

	err = peer.Close()
	assert.Nil(t, err)
}

func TestRegisteringTopics(t *testing.T) {
	network := memp2p.NewNetwork()

	messenger, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	processor := &mock.MessageProcessorStub{}

	// Cannot register a MessageProcessor to a topic that doesn't exist.
	err = messenger.RegisterMessageProcessor("rocket", "", processor)
	assert.True(t, errors.Is(err, p2p.ErrNilTopic))

	// Create a proper topic.
	assert.False(t, messenger.HasTopic("rocket"))
	assert.Nil(t, messenger.CreateTopic("rocket", false))
	assert.True(t, messenger.HasTopic("rocket"))

	// The newly created topic has no MessageProcessor attached to it, so we
	// attach one now.
	assert.Nil(t, messenger.TopicValidator("rocket"))
	err = messenger.RegisterMessageProcessor("rocket", "", processor)
	assert.Nil(t, err)
	assert.Equal(t, processor, messenger.TopicValidator("rocket"))

	// Cannot unregister a MessageProcessor from a topic that doesn't exist.
	err = messenger.UnregisterMessageProcessor("albatross", "")
	assert.True(t, errors.Is(err, p2p.ErrNilTopic))

	// Cannot unregister a MessageProcessor from a topic that doesn't have a
	// MessageProcessor, even if the topic itself exists.
	err = messenger.CreateTopic("nitrous_oxide", false)
	assert.Nil(t, err)
	err = messenger.UnregisterMessageProcessor("nitrous_oxide", "")
	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)

	// Unregister the MessageProcessor from a topic that exists and has a
	// MessageProcessor.
	err = messenger.UnregisterMessageProcessor("rocket", "")
	assert.Nil(t, err)
	assert.True(t, messenger.HasTopic("rocket"))
	assert.Nil(t, messenger.TopicValidator("rocket"))

	// Disallow creating duplicate topics.
	err = messenger.CreateTopic("more_rockets", false)
	assert.Nil(t, err)
	assert.True(t, messenger.HasTopic("more_rockets"))
	err = messenger.CreateTopic("more_rockets", false)
	assert.NotNil(t, err)
}

func TestBroadcastingMessages(t *testing.T) {
	network := memp2p.NewNetwork()

	numPeers := 4
	peers := make([]*memp2p.Messenger, numPeers)
	for i := 0; i < numPeers; i++ {
		peer, _ := memp2p.NewMessenger(network)
		_ = peer.CreateTopic("rocket", false)
		peers[i] = peer
	}

	// Send a message to everybody.
	_ = peers[0].BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket"))
	time.Sleep(1 * time.Second)
	testReceivedMessages(t, peers, map[int]uint64{0: 1, 1: 1, 2: 1, 3: 1, 4: 1})

	// Send a message after disconnecting. No new messages should get broadcast
	err := peers[0].Close()
	assert.Nil(t, err)
	_ = peers[0].BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket again"))
	time.Sleep(1 * time.Second)
	testReceivedMessages(t, peers, map[int]uint64{0: 1, 1: 1, 2: 1, 3: 1, 4: 1})

	peers[2].Broadcast("rocket", []byte("launch another rocket"))
	time.Sleep(1 * time.Second)
	testReceivedMessages(t, peers, map[int]uint64{0: 1, 1: 2, 2: 2, 3: 2, 4: 2})

	peers[2].Broadcast("nitrous_oxide", []byte("this message should not get broadcast"))
	time.Sleep(1 * time.Second)
	testReceivedMessages(t, peers, map[int]uint64{0: 1, 1: 2, 2: 2, 3: 2, 4: 2})
}

func testReceivedMessages(t *testing.T, peers []*memp2p.Messenger, receivedNumMap map[int]uint64) {
	for idx, p := range peers {
		val, found := receivedNumMap[idx]
		if !found {
			assert.Fail(t, fmt.Sprintf("number of messages received was not defined for index %d", idx))
			return
		}

		assert.Equal(t, val, p.NumMessagesReceived(), "for peer on index %d", idx)
	}
}

func TestConnectivityAndTopics(t *testing.T) {
	network := memp2p.NewNetwork()

	// Create 4 peers on the network, all listening to the topic "rocket".
	numPeers := 4
	peers := make([]*memp2p.Messenger, numPeers)
	for i := 0; i < numPeers; i++ {
		peer, _ := memp2p.NewMessenger(network)
		_ = peer.CreateTopic("rocket", false)
		peers[i] = peer
	}

	// Peers 2 and 3 also listen on the topic "carbohydrate"
	_ = peers[2].CreateTopic("carbohydrate", false)
	_ = peers[2].RegisterMessageProcessor("carbohydrate", "", &mock.MessageProcessorStub{})
	_ = peers[3].CreateTopic("carbohydrate", false)
	_ = peers[3].RegisterMessageProcessor("carbohydrate", "", &mock.MessageProcessorStub{})

	// Test to which peers is Peer0 connected, based on the topics they listen to.
	peer0 := peers[0]
	assert.Equal(t, numPeers, len(network.PeerIDs()))
	assert.Equal(t, numPeers-1, len(peer0.ConnectedPeers()))
	assert.Equal(t, numPeers-1, len(peer0.ConnectedPeersOnTopic("rocket")))
	assert.Equal(t, 2, len(peer0.ConnectedPeersOnTopic("carbohydrate")))
}

func TestSendingDirectMessages(t *testing.T) {
	network := memp2p.NewNetwork()

	peer1, _ := memp2p.NewMessenger(network)
	peer2, _ := memp2p.NewMessenger(network)

	// Peer1 attempts to send a direct message to Peer2 on topic "rocket", but
	// Peer2 is not listening to this topic.
	_ = peer1.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), peer2.ID())
	time.Sleep(time.Millisecond * 100)

	// The same as above, but in reverse (Peer2 sends to Peer1).
	_ = peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), peer1.ID())
	time.Sleep(time.Millisecond * 100)

	// Both peers did not get the message
	assert.Equal(t, uint64(0), peer1.NumMessagesReceived())
	assert.Equal(t, uint64(0), peer2.NumMessagesReceived())

	// Create a topic on Peer1. This doesn't help, because Peer2 still can't
	// receive messages on topic "rocket".
	_ = peer1.CreateTopic("nitrous_oxide", false)
	_ = peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), peer1.ID())
	time.Sleep(time.Millisecond * 100)

	// peer1 still did not get the message
	assert.Equal(t, uint64(0), peer1.NumMessagesReceived())

	// Finally, create the topic "rocket" on Peer1
	// This allows it to receive a message on this topic from Peer2.
	_ = peer1.CreateTopic("rocket", false)
	_ = peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), peer1.ID())
	time.Sleep(time.Millisecond * 100)

	// Peer1 got the message
	assert.Equal(t, uint64(1), peer1.NumMessagesReceived())
}
