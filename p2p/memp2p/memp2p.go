package memp2p

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MemP2PMessenger struct {
	Network     *MemP2PNetwork
	P2P_ID      p2p.PeerID
	Address     string
	Topics      map[string]p2p.MessageProcessor
	TopicsMutex sync.RWMutex
}

func NewMemP2PMessenger(network *MemP2PNetwork) *MemP2PMessenger {
	ID := fmt.Sprintf("Peer%d", len(network.PeerIDs())+1)
	Address := fmt.Sprintf("/memp2p/%s", string(ID))

	messenger := &MemP2PMessenger{
		Network:     network,
		P2P_ID:      p2p.PeerID(ID),
		Address:     Address,
		Topics:      make(map[string]p2p.MessageProcessor),
		TopicsMutex: sync.RWMutex{},
	}

	network.RegisterPeer(messenger)

	return messenger
}

func (messenger *MemP2PMessenger) ID() p2p.PeerID {
	return messenger.P2P_ID
}

func (messenger *MemP2PMessenger) Peers() []p2p.PeerID {
	// If the messenger is connected to the network, it has knowledge of all
	// other peers.
	if !messenger.IsConnectedToNetwork() {
		return []p2p.PeerID{}
	}
	return messenger.Network.PeerIDs()
}

func (messenger *MemP2PMessenger) Addresses() []string {
	addresses := make([]string, 1)
	addresses[0] = messenger.Address
	return addresses
}

func (messenger *MemP2PMessenger) ConnectToPeer() error {
	if !messenger.IsConnectedToNetwork() {
		return errors.New("Peer not connected to network, can't connect to any other peer.")
	}
	// Do nothing, all peers are connected to each other already.
	return nil
}

func (messenger *MemP2PMessenger) IsConnectedToNetwork() bool {
	return messenger.Network.IsPeerConnected(messenger.ID())
}

func (messenger *MemP2PMessenger) IsConnected(peerID p2p.PeerID) bool {
	// If the messenger is connected to the network, it is connected to all other peers.
	return messenger.IsConnectedToNetwork()
}

func (messenger *MemP2PMessenger) ConnectedPeers() []p2p.PeerID {
	if !messenger.IsConnectedToNetwork() {
		return []p2p.PeerID{}
	}
	return messenger.Network.PeerIDs()
}

func (messenger *MemP2PMessenger) ConnectedAddresses() []string {
	if !messenger.IsConnectedToNetwork() {
		return []string{}
	}
	return messenger.Network.ListAddresses()
}

func (messenger *MemP2PMessenger) PeerAddress(pid p2p.PeerID) string {
	return fmt.Sprintf("/memp2p/%s", string(pid))
}

func (messenger *MemP2PMessenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	messenger.TopicsMutex.Lock()
	var filteredPeers []p2p.PeerID
	if !messenger.Network.IsPeerConnected(messenger.ID()) {
		return filteredPeers
	}
	for peerID, peer := range messenger.Network.Peers() {
		if peer.HasTopic(topic) {
			filteredPeers = append(filteredPeers, p2p.PeerID(peerID))
		}
	}
	messenger.TopicsMutex.Unlock()
	return filteredPeers
}

func (messenger *MemP2PMessenger) TrimConnections() {
	// Do nothing.
}

func (messenger *MemP2PMessenger) Bootstrap() error {
	// Do nothing.
	return nil
}

func (messenger *MemP2PMessenger) CreateTopic(name string, createChannelForTopic bool) error {
	messenger.TopicsMutex.Lock()

	_, found := messenger.Topics[name]
	if found {
		messenger.TopicsMutex.Unlock()
		return p2p.ErrTopicAlreadyExists
	}

	messenger.Topics[name] = nil
	messenger.TopicsMutex.Unlock()
	return nil
}

func (messenger *MemP2PMessenger) HasTopic(name string) bool {
	messenger.TopicsMutex.RLock()
	_, found := messenger.Topics[name]
	messenger.TopicsMutex.RUnlock()

	return found
}

func (messenger *MemP2PMessenger) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if handler == nil {
		return p2p.ErrNilValidator
	}

	messenger.TopicsMutex.Lock()
	defer messenger.TopicsMutex.Unlock()
	validator, found := messenger.Topics[topic]

	if !found {
		return p2p.ErrNilTopic
	}

	if validator != nil {
		return p2p.ErrTopicValidatorOperationNotSupported
	}

	messenger.Topics[topic] = handler
	return nil
}

func (messenger *MemP2PMessenger) UnregisterMessageProcessor(topic string) error {
	messenger.TopicsMutex.Lock()
	defer messenger.TopicsMutex.Unlock()
	validator, found := messenger.Topics[topic]

	if !found {
		return p2p.ErrNilTopic
	}

	if validator == nil {
		return p2p.ErrTopicValidatorOperationNotSupported
	}

	messenger.Topics[topic] = nil
	return nil
}

func (messenger *MemP2PMessenger) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return nil
}

// BroadcastOnChannelBlocking sends the message to all peers in the network. It
// calls ParametricBroadcast() with async=false, which means that peers will
// have their ReceiveMessage() function called synchronously. The call
// to ParametricBroadcast() is done synchronously as well. This function should
// be called as a go-routine.
func (messenger *MemP2PMessenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) {
	messenger.ParametricBroadcast(topic, buff, false)
}

// BroadcastOnChannelBlocking sends the message to all peers in the network. It
// calls ParametricBroadcast() with async=false, which means that peers will
// have their ReceiveMessage() function called synchronously. The call
// to ParametricBroadcast() is done as a go-routine, which means this function
// is, in fact, non-blocking, but it is identical with
// BroadcastOnChannelBlocking() in all other regards.
func (messenger *MemP2PMessenger) BroadcastOnChannel(channel string, topic string, buff []byte) {
	messenger.ParametricBroadcast(topic, buff, false)
}

// Broadcast asynchronously sends the message to all peers in the network. It
// calls ParametricBroadcast() with async=true, which means that peers will
// have their ReceiveMessage() function independently called as go-routines.
func (messenger *MemP2PMessenger) Broadcast(topic string, buff []byte) {
	messenger.ParametricBroadcast(topic, buff, true)
}

// ParametricBroadcast sends a message to all peers in the network, with the
// possibility to choose from asynchronous or synchronous sending.
func (messenger *MemP2PMessenger) ParametricBroadcast(topic string, data []byte, async bool) {
	if messenger.IsConnectedToNetwork() {
		message := NewMemP2PMessage(topic, data, messenger.ID())
		for _, peer := range messenger.Network.Peers() {
			if async == true {
				go peer.ReceiveMessage(topic, message)
			} else {
				peer.ReceiveMessage(topic, message)
			}
		}
	} else {
	}
}

func (messenger *MemP2PMessenger) SentToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	if messenger.IsConnectedToNetwork() {
		message := NewMemP2PMessage(topic, buff, messenger.ID())
		destinationPeer := messenger.Network.Peers()[string(peerID)]
		return destinationPeer.ReceiveMessage(topic, message)
	} else {
		return errors.New("Peer not connected to network, cannot send anything")
	}
}

func (messenger *MemP2PMessenger) ReceiveMessage(topic string, message p2p.MessageP2P) error {
	messenger.TopicsMutex.Lock()
	validator, found := messenger.Topics[topic]
	messenger.TopicsMutex.Unlock()

	if !found {
		return p2p.ErrNilTopic
	}

	if validator == nil {
		return p2p.ErrNilValidator
	}

	validator.ProcessReceivedMessage(message)

	if messenger.Network.LogMessages {
		messenger.Network.LogMessage(message)
	}
	return nil
}

func (messenger *MemP2PMessenger) Close() {
	// Remove messenger from the network it references.
	messenger.Network.UnregisterPeer(messenger.ID())
}
