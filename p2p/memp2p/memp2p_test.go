package memp2p_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/memp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestInitializingNetworkwith4Peers(t *testing.T) {
	network, err := memp2p.NewNetwork()
	assert.Nil(t, err)

	peer1, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer2, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer3, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer4, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

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
	processor := mock.MockMessageProcessor{}

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

	err = peer1.Close()
	assert.Nil(t, err)
	peerIDs := network.PeerIDs()
	peersMap := network.Peers()
	assert.Equal(t, 3, len(peerIDs))
	assert.Equal(t, 3, len(peersMap))
	assert.Equal(t, p2p.PeerID("Peer2"), peerIDs[0])
	assert.Equal(t, p2p.PeerID("Peer3"), peerIDs[1])
	assert.Equal(t, p2p.PeerID("Peer4"), peerIDs[2])
	assert.NotContains(t, peersMap, "Peer1")
}

func TestBroadcastingMessages(t *testing.T) {
	network, err := memp2p.NewNetwork()
	assert.Nil(t, err)
	network.LogMessages = true

	peer1, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer2, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer3, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer4, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	// All peers listen to the topic "rocket"
	err = peer1.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer1.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer1.ID()))
	assert.Nil(t, err)
	err = peer2.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer2.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer2.ID()))
	assert.Nil(t, err)
	err = peer3.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer3.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer3.ID()))
	assert.Nil(t, err)
	err = peer4.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer4.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer4.ID()))
	assert.Nil(t, err)

	// Send a message to everybody.
	peer1.BroadcastOnChannelBlocking("rocket", "rocket", []byte("launch the rocket"))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, len(network.Messages))

	// Send a message after disconnecting. No new messages should appear in the log.
	err = peer1.Close()
	assert.Nil(t, err)
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

func TestConnectivityAndTopics(t *testing.T) {
	network, err := memp2p.NewNetwork()
	assert.Nil(t, err)
	network.LogMessages = true

	peer1, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer2, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer3, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer4, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	// All peers listen to the topic "rocket"
	err = peer1.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer1.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer1.ID()))
	assert.Nil(t, err)
	err = peer2.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer2.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer2.ID()))
	assert.Nil(t, err)
	err = peer3.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer3.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer3.ID()))
	assert.Nil(t, err)
	err = peer4.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer4.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer4.ID()))
	assert.Nil(t, err)

	// Peers 2 and 3 also listen on the topic "carbohydrate"
	err = peer2.CreateTopic("carbohydrate", false)
	assert.Nil(t, err)
	err = peer2.RegisterMessageProcessor("carbohydrate", mock.NewMockMessageProcessor(peer2.ID()))
	assert.Nil(t, err)
	err = peer3.CreateTopic("carbohydrate", false)
	assert.Nil(t, err)
	err = peer3.RegisterMessageProcessor("carbohydrate", mock.NewMockMessageProcessor(peer3.ID()))
	assert.Nil(t, err)

	peers1234 := []p2p.PeerID{peer1.ID(), peer2.ID(), peer3.ID(), peer4.ID()}
	peers234 := []p2p.PeerID{peer2.ID(), peer3.ID(), peer4.ID()}
	peers23 := []p2p.PeerID{peer2.ID(), peer3.ID()}

	assert.Equal(t, peers1234, network.PeerIDs())
	assert.Equal(t, peers234, peer1.ConnectedPeers())
	assert.Equal(t, peers234, peer1.ConnectedPeersOnTopic("rocket"))
	assert.Equal(t, peers23, peer1.ConnectedPeersOnTopic("carbohydrate"))
}

func TestSendingDirectMessages(t *testing.T) {
	network, err := memp2p.NewNetwork()
	assert.Nil(t, err)
	network.LogMessages = true

	peer1, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)
	peer2, err := memp2p.NewMessenger(network)
	assert.Nil(t, err)

	assert.NotNil(t, peer1.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer2"))
	assert.NotNil(t, peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1"))
	assert.Equal(t, 0, len(network.Messages))

	err = peer1.CreateTopic("nitrous_oxide", false)
	assert.Nil(t, err)
	assert.NotNil(t, peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1"))
	assert.Equal(t, 0, len(network.Messages))

	err = peer1.CreateTopic("rocket", false)
	assert.Nil(t, err)
	err = peer1.RegisterMessageProcessor("rocket", mock.NewMockMessageProcessor(peer1.ID()))
	assert.Nil(t, err)
	assert.Nil(t, peer2.SendToConnectedPeer("rocket", []byte("try to launch this rocket"), "Peer1"))
	assert.Equal(t, 1, len(network.Messages))
}
