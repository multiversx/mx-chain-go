package memp2p

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

type MockMessageProcessor struct {
	Peer p2p.PeerID
}

func NewMockMessageProcessor(peer p2p.PeerID) MockMessageProcessor {
	processor := MockMessageProcessor{}
	processor.Peer = peer
	return processor
}

func (processor MockMessageProcessor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	fmt.Printf("Message received by %s from %s: %s\n", string(processor.Peer), string(message.Peer()), string(message.Data()))
	return nil
}

func Test_Initializing_MemP2PNetwork_with_4_Peers(t *testing.T) {
	network := NewMemP2PNetwork()

	peer1 := NewMemP2PMessenger(network)
	peer2 := NewMemP2PMessenger(network)
	peer3 := NewMemP2PMessenger(network)
	peer4 := NewMemP2PMessenger(network)

	assert.Equal(t, 4, len(network.Peers()))
	assert.Equal(t, "Peer1", string(peer1.ID()))
	assert.Equal(t, "Peer2", string(peer2.ID()))
	assert.Equal(t, "Peer3", string(peer3.ID()))
	assert.Equal(t, "Peer4", string(peer4.ID()))

	assert.Equal(t, "/memp2p/Peer1", peer1.Address)
	assert.Equal(t, "/memp2p/Peer2", peer2.Address)
	assert.Equal(t, "/memp2p/Peer3", peer3.Address)
	assert.Equal(t, "/memp2p/Peer4", peer4.Address)

	peers := peer4.Peers()
	assert.Equal(t, p2p.PeerID("Peer1"), peers[0])
	assert.Equal(t, p2p.PeerID("Peer2"), peers[1])
	assert.Equal(t, p2p.PeerID("Peer3"), peers[2])
	assert.Equal(t, p2p.PeerID("Peer4"), peers[3])

	assert.Equal(t, 1, len(peer2.Addresses()))
	assert.Equal(t, "/memp2p/Peer2", peer2.Addresses()[0])

	// Disallow creating duplicate topics.
	assert.Nil(t, peer1.CreateTopic("rockets", false))
	assert.True(t, peer1.HasTopic("rockets"))
	assert.NotNil(t, peer1.CreateTopic("rockets", false))
	assert.True(t, peer1.HasTopic("rockets"))

	assert.False(t, peer1.HasTopic("nitrous_oxide"))
	assert.Nil(t, peer1.CreateTopic("nitrous_oxide", false))
	assert.True(t, peer1.HasTopic("nitrous_oxide"))

	messenger := peer2
	processor := MockMessageProcessor{}

	assert.Equal(t, p2p.ErrNilTopic, messenger.RegisterMessageProcessor("rocket", processor))
	assert.Nil(t, messenger.CreateTopic("rocket", false))
	assert.True(t, messenger.HasTopic("rocket"))
	assert.Nil(t, messenger.Topics["rocket"])
	assert.Nil(t, messenger.RegisterMessageProcessor("rocket", processor))
	assert.Equal(t, processor, messenger.Topics["rocket"])
	assert.Equal(t, p2p.ErrNilTopic, messenger.UnregisterMessageProcessor("albatross"))
	assert.Nil(t, messenger.CreateTopic("nitrous_oxide", false))
	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, messenger.UnregisterMessageProcessor("nitrous_oxide"))
	assert.Nil(t, messenger.UnregisterMessageProcessor("rocket"))
	assert.True(t, messenger.HasTopic("rocket"))
	assert.Nil(t, messenger.Topics["rocket"])

	peer1.Close()
	peerIDs := network.PeerIDs()
	peersMap := network.Peers()
	assert.Equal(t, 3, len(peerIDs))
	assert.Equal(t, 3, len(peersMap))
	assert.Equal(t, p2p.PeerID("Peer2"), peerIDs[0])
	assert.Equal(t, p2p.PeerID("Peer3"), peerIDs[1])
	assert.Equal(t, p2p.PeerID("Peer4"), peerIDs[2])
	assert.NotContains(t, peersMap, "Peer1")
}

func Test_Broadcasting_Messages(t *testing.T) {
	network := NewMemP2PNetwork()
	network.LogMessages = true

	peer1 := NewMemP2PMessenger(network)
	peer2 := NewMemP2PMessenger(network)
	peer3 := NewMemP2PMessenger(network)
	peer4 := NewMemP2PMessenger(network)

	// All peers listen to the topic "rocket"
	peer1.CreateTopic("rocket", false)
	peer1.RegisterMessageProcessor("rocket", NewMockMessageProcessor(peer1.ID()))
	peer2.CreateTopic("rocket", false)
	peer2.RegisterMessageProcessor("rocket", NewMockMessageProcessor(peer2.ID()))
	peer3.CreateTopic("rocket", false)
	peer3.RegisterMessageProcessor("rocket", NewMockMessageProcessor(peer3.ID()))
	peer4.CreateTopic("rocket", false)
	peer4.RegisterMessageProcessor("rocket", NewMockMessageProcessor(peer4.ID()))

	// Send a message to everybody.
	peer1.BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, len(network.Messages))

	// Send a message after disconnecting. No new messages should appear in the log.
	peer1.Close()
	peer1.BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket again"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, len(network.Messages))

	peer2.Broadcast("rocket", []byte("launch another rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 7, len(network.Messages))

	peer3.Broadcast("nitrous_oxide", []byte("this is not a rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 7, len(network.Messages))
}

func Test_Sending_Direct_Messages(t *testing.T) {
	network := NewMemP2PNetwork()
	network.LogMessages = true

	peer1 := NewMemP2PMessenger(network)
	peer2 := NewMemP2PMessenger(network)

	assert.NotNil(t, peer1.SentToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer2"))
	assert.NotNil(t, peer2.SentToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1"))
	assert.Equal(t, 0, len(network.Messages))

	peer1.CreateTopic("nitrous_oxide", false)
	assert.NotNil(t, peer2.SentToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1"))
	assert.Equal(t, 0, len(network.Messages))

	peer1.CreateTopic("rocket", false)
	peer1.RegisterMessageProcessor("rocket", NewMockMessageProcessor(peer1.ID()))
	assert.Nil(t, peer2.SentToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1"))
	assert.Equal(t, 1, len(network.Messages))
}
