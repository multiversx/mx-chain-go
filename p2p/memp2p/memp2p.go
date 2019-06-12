package memp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// MemP2PNetwork -------------------------
type MemP2PNetwork struct {
	Mutex   sync.RWMutex
	PeerIDs []p2p.PeerID
	Peers   map[string]*MemP2PMessenger
}

func NewMemP2PNetwork() *MemP2PNetwork {
	var peerIDs []p2p.PeerID
	return &MemP2PNetwork{
		Mutex:   sync.RWMutex{},
		PeerIDs: peerIDs,
		Peers:   make(map[string]*MemP2PMessenger),
	}
}

func (network *MemP2PNetwork) ListAddresses() []string {
	network.Mutex.Lock()
	addresses := make([]string, len(network.PeerIDs))
	i := 0
	for _, peerID := range network.PeerIDs {
		addresses[i] = fmt.Sprintf("/memp2p/%s", peerID)
		i++
	}
	network.Mutex.Unlock()
	return addresses
}

func (network *MemP2PNetwork) RegisterPeer(messenger *MemP2PMessenger) {
	network.Mutex.RLock()
	network.PeerIDs = append(network.PeerIDs, messenger.ID())
	network.Peers[string(messenger.ID())] = messenger
	network.Mutex.RUnlock()
}

// MemP2PMessenger -------------------------
type MemP2PMessenger struct {
	Network     *MemP2PNetwork
	P2P_ID      p2p.PeerID
	Address     string
	Topics      map[string]p2p.MessageProcessor
	TopicsMutex sync.RWMutex
}

func NewMemP2PMessenger(network *MemP2PNetwork) *MemP2PMessenger {
	ID := fmt.Sprintf("Peer%d", len(network.PeerIDs)+1)
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
	return messenger.Network.PeerIDs
}

func (messenger *MemP2PMessenger) Addresses() []string {
	addresses := make([]string, 1)
	addresses[0] = messenger.Address
	return addresses
}

func (messenger *MemP2PMessenger) ConnectToPeer() error {
	// Do nothing, all peers are connected to each other already.
	return nil
}

func (messenger *MemP2PMessenger) IsConnected(peerID p2p.PeerID) bool {
	return true
}

func (messenger *MemP2PMessenger) ConnectedPeers() []p2p.PeerID {
	return messenger.Network.PeerIDs
}

func (messenger *MemP2PMessenger) ConnectedAddresses() []string {
	return messenger.Network.ListAddresses()
}

func (messenger *MemP2PMessenger) PeerAddress(pid p2p.PeerID) string {
	return fmt.Sprintf("/memp2p/%s", string(pid))
}

func (messenger *MemP2PMessenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	messenger.TopicsMutex.Lock()
	var filteredPeers []p2p.PeerID
	for peerID, peer := range messenger.Network.Peers {
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

func (messenger *MemP2PMessenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) {
}

func (messenger *MemP2PMessenger) ParametricBroadcast(topic string, message p2p.MessageP2P, async bool) {
	for peerID, peer := range messenger.Network.Peers {
		if async == true {
			go peer.ReceiveMessage(topic, message)
		} else {
			peer.ReceiveMessage(topic, message)
		}
	}
}

func (messenger *MemP2PMessenger) ReceiveMessage(topic string, message p2p.MessageP2P) error {
	messenger.TopicsMutex.Lock()
	defer messenger.TopicsMutex.Unlock()
	validator, found := messenger.Topics[topic]

	if !found {
		return nil
	}

	if validator == nil {
		return nil
	}

	validator.ProcessReceivedMessage(message)
	return nil
}

func (messenger *MemP2PMessenger) Close() {
	// Remove messenger from the network it references.
}
