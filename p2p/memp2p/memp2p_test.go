package memp2p_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/memp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestInitializingNetworkAndPeer(t *testing.T) {
	network, err := memp2p.NewNetwork()
	assert.Nil(t, err)

	peer, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(network.Peers()))
	assert.Equal(t, "Peer1", string(peer.ID()))

	assert.Equal(t, 1, len(peer.Addresses()))
	assert.Equal(t, "/memp2p/Peer1", peer.Addresses()[0])

	err = peer.Close()
	assert.Nil(t, err)
}

func TestInitializingNetworkwith4Peers(t *testing.T) {
	network, _ := memp2p.NewNetwork()

	peer1, _ := memp2p.NewMessenger(network)
	_, _ = memp2p.NewMessenger(network)
	_, _ = memp2p.NewMessenger(network)
	peer4, _ := memp2p.NewMessenger(network)

	assert.Equal(t, 4, len(network.Peers()))
	assert.Equal(t, 4, len(peer4.Peers()))

	expectedAddresses := []string{"/memp2p/Peer1", "/memp2p/Peer2", "/memp2p/Peer3", "/memp2p/Peer4"}
	assert.Equal(t, expectedAddresses, network.ListAddresses())

	peer1.Close()
	peerIDs := network.PeerIDs()
	peersMap := network.Peers()
	assert.Equal(t, 3, len(peerIDs))
	assert.Equal(t, 3, len(peersMap))
	assert.Equal(t, p2p.PeerID("Peer2"), peerIDs[0])
	assert.Equal(t, p2p.PeerID("Peer3"), peerIDs[1])
	assert.Equal(t, p2p.PeerID("Peer4"), peerIDs[2])
	assert.NotContains(t, peersMap, "Peer1")

	expectedAddresses = []string{"/memp2p/Peer2", "/memp2p/Peer3", "/memp2p/Peer4"}
	assert.Equal(t, expectedAddresses, network.ListAddresses())
}

func TestRegisteringTopics(t *testing.T) {
	network, err := memp2p.NewNetwork()
	assert.Nil(t, err)

	messenger, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	processor := &mock.MockMessageProcessor{}

	// Cannot register a MessageProcessor to a topic that doesn't exist.
	err = messenger.RegisterMessageProcessor("rocket", processor)
	assert.Equal(t, p2p.ErrNilTopic, err)

	// Create a proper topic.
	assert.False(t, messenger.HasTopic("rocket"))
	assert.Nil(t, messenger.CreateTopic("rocket", false))
	assert.True(t, messenger.HasTopic("rocket"))

	// The newly created topic has no MessageProcessor attached to it, so we
	// attach one now.
	assert.Nil(t, messenger.Topics["rocket"])
	err = messenger.RegisterMessageProcessor("rocket", processor)
	assert.Nil(t, err)
	assert.Equal(t, processor, messenger.Topics["rocket"])

	// Cannot unregister a MessageProcessor from a topic that doesn't exist.
	err = messenger.UnregisterMessageProcessor("albatross")
	assert.Equal(t, p2p.ErrNilTopic, err)

	// Cannot unregister a MessageProcessor from a topic that doesn't have a
	// MessageProcessor, even if the topic itself exists.
	err = messenger.CreateTopic("nitrous_oxide", false)
	assert.Nil(t, err)
	err = messenger.UnregisterMessageProcessor("nitrous_oxide")
	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)

	// Unregister the MessageProcessor from a topic that exists and has a
	// MessageProcessor.
	err = messenger.UnregisterMessageProcessor("rocket")
	assert.Nil(t, err)
	assert.True(t, messenger.HasTopic("rocket"))
	assert.Nil(t, messenger.Topics["rocket"])

	// Disallow creating duplicate topics.
	err = messenger.CreateTopic("more_rockets", false)
	assert.Nil(t, err)
	assert.True(t, messenger.HasTopic("more_rockets"))
	err = messenger.CreateTopic("more_rockets", false)
	assert.NotNil(t, err)
}

func TestBroadcastingMessages(t *testing.T) {
	network, _ := memp2p.NewNetwork()
	network.LogMessages = true

	peer1, _ := memp2p.NewMessenger(network)
	peer2, _ := memp2p.NewMessenger(network)
	peer3, _ := memp2p.NewMessenger(network)
	peer4, _ := memp2p.NewMessenger(network)

	// All peers listen to the topic "rocket"
	_ = peer1.CreateTopic("rocket", false)
	_ = peer1.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer1.ID()))
	_ = peer2.CreateTopic("rocket", false)
	_ = peer2.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer2.ID()))
	_ = peer3.CreateTopic("rocket", false)
	_ = peer3.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer3.ID()))
	_ = peer4.CreateTopic("rocket", false)
	_ = peer4.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer4.ID()))

	// Send a message to everybody.
	_ = peer1.BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, network.GetMessageCount())

	// Send a message after disconnecting. No new messages should appear in the log.
	err := peer1.Close()
	assert.Nil(t, err)
	_ = peer1.BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket again"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, network.GetMessageCount())

	peer2.Broadcast("rocket", []byte("launch another rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 7, network.GetMessageCount())

	peer3.Broadcast("nitrous_oxide", []byte("this is not a rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 7, network.GetMessageCount())
}

func TestConnectivityAndTopics(t *testing.T) {
	network, _ := memp2p.NewNetwork()
	network.LogMessages = true

	// Create 4 peers on the network, all listening to the topic "rocket".
	for i := 1; i <= 4; i++ {
		peer, _ := memp2p.NewMessenger(network)
		_ = peer.CreateTopic("rocket", false)
		processor := mock.NewMockMessageProcessor(peer.ID())
		_ = peer.RegisterMessageProcessor("rocket", processor)
	}

	// Peers 2 and 3 also listen on the topic "carbohydrate"
	peer2 := network.Peers()["Peer2"]
	peer3 := network.Peers()["Peer3"]
	_ = peer2.CreateTopic("carbohydrate", false)
	_ = peer2.RegisterMessageProcessor("carbohydrate", mock.NewMockMessageProcessor(peer2.ID()))
	_ = peer3.CreateTopic("carbohydrate", false)
	_ = peer3.RegisterMessageProcessor("carbohydrate", mock.NewMockMessageProcessor(peer3.ID()))

	peers1234 := []p2p.PeerID{"Peer1", "Peer2", "Peer3", "Peer4"}
	peers234 := []p2p.PeerID{"Peer2", "Peer3", "Peer4"}
	peers23 := []p2p.PeerID{"Peer2", "Peer3"}

	// Test to which peers is Peer1 connected, based on the topics they listen to.
	peer1 := network.Peers()["Peer1"]
	assert.Equal(t, peers1234, network.PeerIDs())
	assert.Equal(t, peers234, peer1.ConnectedPeers())
	assert.Equal(t, peers234, peer1.ConnectedPeersOnTopic("rocket"))
	assert.Equal(t, peers23, peer1.ConnectedPeersOnTopic("carbohydrate"))
}

func TestSendingDirectMessages(t *testing.T) {
	network, _ := memp2p.NewNetwork()
	network.LogMessages = true

	peer1, _ := memp2p.NewMessenger(network)
	peer2, _ := memp2p.NewMessenger(network)

	var err error

	// Peer1 attempts to send a direct message to Peer2 on topic "rocket", but
	// Peer2 is not listening to this topic.
	err = peer1.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer2")
	assert.Equal(t, p2p.ErrNilTopic, err)

	// The same as above, but in reverse (Peer2 sends to Peer1).
	err = peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1")
	assert.Equal(t, p2p.ErrNilTopic, err)

	// The network has logged no processed messages.
	assert.Equal(t, 0, network.GetMessageCount())

	// Create a topic on Peer1. This doesn't help, because Peer2 still can't
	// receive messages on topic "rocket".
	_ = peer1.CreateTopic("nitrous_oxide", false)
	err = peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1")
	assert.Equal(t, p2p.ErrNilTopic, err)

	// The network has still not logged any processed messages.
	assert.Equal(t, 0, network.GetMessageCount())

	// Finally, create the topic "rocket" on Peer1 and register a
	// MessageProcessor. This allows it to receive a message on this topic from Peer2.
	_ = peer1.CreateTopic("rocket", false)
	_ = peer1.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer1.ID()))
	err = peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1")
	assert.Nil(t, err)

	// The network has finally logged a processed message.
	assert.Equal(t, 1, network.GetMessageCount())
}
