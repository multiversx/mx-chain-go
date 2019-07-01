package memp2p

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MemP2PMessenger is an implementation of the p2p.Messenger interface that
// uses no real networking code, but instead connects to a network simulated in
// memory (the MemP2PNetwork struct). The MemP2PMessenger is intended for use
// in automated tests instead of the real libp2p, in order to speed up their
// execution and reduce resource usage.
//
// All message-sending functions imitate the synchronous/asynchronous
// behavior of the Messenger struct originally implemented for libp2p. Note
// that the MemP2PNetwork ensures that all messengers are connected to all
// other messengers, thus when a MemP2PMessenger is connected to an in-memory
// network, it reports being connected to all the nodes. Consequently,
// broadcasting a message will be received by all the messengers in the
// network.
type MemP2PMessenger struct {
	Network     *MemP2PNetwork
	P2P_ID      p2p.PeerID
	Address     string
	Topics      map[string]p2p.MessageProcessor
	TopicsMutex sync.RWMutex
}

// NewMemP2PMessenger constructs a new MemP2PMessenger that is connected to the
// MemP2PNetwork instance provided as argument.
func NewMemP2PMessenger(network *MemP2PNetwork) (*MemP2PMessenger, error) {
	if network == nil {
		return nil, errors.New("Cannot create a MemP2PMessenger for a nil network")
	}

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

	return messenger, nil
}

// ID returns the P2P ID of the messenger
func (messenger *MemP2PMessenger) ID() p2p.PeerID {
	return messenger.P2P_ID
}

// Peers returns a slice containing the P2P IDs of all the other peers that is
// has knowledge of. Since this is an in-memory network structured as a fully
// connected graph, this function returns the list of the P2P IDs of all the
// peers in the network (assuming this Messenger is connected).
func (messenger *MemP2PMessenger) Peers() []p2p.PeerID {
	fmt.Printf("memp2p:Peers\n")
	// If the messenger is connected to the network, it has knowledge of all
	// other peers.
	if !messenger.IsConnectedToNetwork() {
		return []p2p.PeerID{}
	}
	return messenger.Network.PeerIDs()
}

// Addresses returns a list of all the addresses that this Messenger is bound
// to and listening to. Being an in-memory simulation, the only possible
// address to return is an artificial one, built by the constructor
// NewMemP2PMessenger().
func (messenger *MemP2PMessenger) Addresses() []string {
	addresses := make([]string, 1)
	addresses[0] = messenger.Address
	return addresses
}

// ConnectToPeer usually does nothing, because peers connected to the in-memory
// network are already all connected to each other. This function will return
// an error if the Messenger is not connected to the network, though.
func (messenger *MemP2PMessenger) ConnectToPeer(address string) error {
	fmt.Printf("memp2p:ConnectToPeer\n")
	if !messenger.IsConnectedToNetwork() {
		return errors.New("Peer not connected to network, can't connect to any other peer.")
	}
	// Do nothing, all peers are connected to each other already.
	return nil
}

// IsConnectedToNetwork returns true if this messenger is connected to the
// in-memory network, false otherwise.
func (messenger *MemP2PMessenger) IsConnectedToNetwork() bool {
	fmt.Printf("memp2p:IsConnectedToNetwork\n")
	return messenger.Network.IsPeerConnected(messenger.ID())
}

// IsConnected returns true if this Messenger is connected to the peer with the
// specified ID. It always returns true if the Messenger is connected to the
// network and false otherwise, regardless of the provided peer ID.
func (messenger *MemP2PMessenger) IsConnected(peerID p2p.PeerID) bool {
	fmt.Printf("memp2p:IsConnected\n")
	return messenger.IsConnectedToNetwork()
}

// ConnectedPeers returns a slice of IDs belonging to the peers to which this
// Messenger is connected. If the Messenger is connected to the in₋memory
// network, then the function returns a slice containing the IDs of all the
// other peers connected to the network. Returns false if the Messenger is
// not connected.
func (messenger *MemP2PMessenger) ConnectedPeers() []p2p.PeerID {
	fmt.Printf("memp2p:ConnectedPeers\n")
	if !messenger.IsConnectedToNetwork() {
		return []p2p.PeerID{}
	}
	return messenger.Network.PeerIDsExceptOne(messenger.ID())
}

// ConnectedAddresses returns a slice of peer addresses to which this Messenger
// is connected. If this Messenger is connected to the network, then the
// addresses of all the other peers in the network are returned.
func (messenger *MemP2PMessenger) ConnectedAddresses() []string {
	fmt.Printf("memp2p:ConnectedAddresses\n")
	if !messenger.IsConnectedToNetwork() {
		return []string{}
	}
	return messenger.Network.ListAddressesExceptOne(messenger.ID())
}

// PeerAddress creates the address string from a given peer ID.
func (messenger *MemP2PMessenger) PeerAddress(pid p2p.PeerID) string {
	return fmt.Sprintf("/memp2p/%s", string(pid))
}

// ConnectedPeersOnTopic returns a slice of IDs belonging to the peers in the
// network that have declared their interest in the given topic and are
// listening to messages on that topic.
func (messenger *MemP2PMessenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	fmt.Printf("memp2p:ConnectedPeersOnTopic\n")
	var filteredPeers []p2p.PeerID
	if !messenger.IsConnectedToNetwork() {
		return filteredPeers
	}

	messenger.TopicsMutex.Lock()
	defer messenger.TopicsMutex.Unlock()

	allPeerIDsExceptThis := messenger.Network.PeerIDsExceptOne(messenger.ID())
	allPeersExceptThis := messenger.Network.PeersExceptOne(messenger.ID())
	for _, peerID := range allPeerIDsExceptThis {
		peer := allPeersExceptThis[peerID]
		if peer.HasTopic(topic) {
			filteredPeers = append(filteredPeers, p2p.PeerID(peerID))
		}
	}

	return filteredPeers
}

// TrimConnections does nothing, as it is not applicable to the in-memory
// messenger.
func (messenger *MemP2PMessenger) TrimConnections() {
}

// Bootstrap does nothing, as it is not applicable to the in-memory messenger.
func (messenger *MemP2PMessenger) Bootstrap() error {
	return nil
}

// CreateTopic adds the topic provided as argument to the list of topics of
// interest for this Messenger. It also registers a nil message validator to
// handle the messages received on this topic.
func (messenger *MemP2PMessenger) CreateTopic(name string, createChannelForTopic bool) error {
	fmt.Printf("memp2p:CreateTopic\n")
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

// HasTopic returns true if this Messenger has declared interest in the given
// topic; returns false otherwise.
func (messenger *MemP2PMessenger) HasTopic(name string) bool {
	fmt.Printf("memp2p:HasTopic\n")
	messenger.TopicsMutex.RLock()
	_, found := messenger.Topics[name]
	messenger.TopicsMutex.RUnlock()

	return found
}

// HasTopicValidator returns true if this Messenger has declared interest in
// the given topic and has registered a non-nil validator on that topic.
// Returns false otherwise.
func (messenger *MemP2PMessenger) HasTopicValidator(name string) bool {
	fmt.Printf("memp2p:HasTopicValidator\n")
	messenger.TopicsMutex.RLock()
	validator, found := messenger.Topics[name]
	messenger.TopicsMutex.RUnlock()

	return found && (validator != nil)
}

// RegisterMessageProcessor sets the provided message processor to be the
// processor of received messages for the given topic.
func (messenger *MemP2PMessenger) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	fmt.Printf("memp2p:RegisterMessageProcessor\n")
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

// UnregisterMessageProcessor unsets the message processor for the given topic
// (sets it to nil).
func (messenger *MemP2PMessenger) UnregisterMessageProcessor(topic string) error {
	fmt.Printf("memp2p:UnregisterMessageProcessor\n")
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

// OutgoingChannelLoadBalancer does nothing, as it is not applicable to the in-memory network.
func (messenger *MemP2PMessenger) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return nil
}

// BroadcastOnChannelBlocking sends the message to all peers in the network. It
// calls parametricBroadcast() with async=false, which means that peers will
// have their ReceiveMessage() function called synchronously. The call
// to parametricBroadcast() is done synchronously as well. This function should
// be called as a go-routine.
func (messenger *MemP2PMessenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) {
	fmt.Printf("memp2p:BroadcastOnChannelBlocking\n")
	messenger.parametricBroadcast(topic, buff, false)
}

// BroadcastOnChannel sends the message to all peers in the network. It calls
// parametricBroadcast() with async=false, which means that peers will have
// their ReceiveMessage() function called synchronously. The call to
// parametricBroadcast() is done as a go-routine, which means this function is,
// in fact, non-blocking, but it is identical with BroadcastOnChannelBlocking()
// in all other regards.
func (messenger *MemP2PMessenger) BroadcastOnChannel(channel string, topic string, buff []byte) {
	fmt.Printf("memp2p:BroadcastOnChannel\n")
	messenger.parametricBroadcast(topic, buff, false)
}

// Broadcast asynchronously sends the message to all peers in the network. It
// calls parametricBroadcast() with async=true, which means that peers will
// have their ReceiveMessage() function independently called as go-routines.
func (messenger *MemP2PMessenger) Broadcast(topic string, buff []byte) {
	fmt.Printf("memp2p:Broadcast\n")
	messenger.parametricBroadcast(topic, buff, true)
}

// parametricBroadcast sends a message to all peers in the network, with the
// possibility to choose from asynchronous or synchronous sending.
func (messenger *MemP2PMessenger) parametricBroadcast(topic string, data []byte, async bool) {
	if messenger.IsConnectedToNetwork() {
		message, _ := NewMemP2PMessage(topic, data, messenger.ID())
		for _, peer := range messenger.Network.Peers() {
			if async == true {
				go peer.ReceiveMessage(topic, message)
			} else {
				peer.ReceiveMessage(topic, message)
			}
		}
	}
}

// SendToConnectedPeer sends a message directly to the peer specified by the ID.
func (messenger *MemP2PMessenger) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	fmt.Printf("memp2p:SendToConnectedPeer\n")
	if messenger.IsConnectedToNetwork() {
		if peerID == messenger.ID() {
			return errors.New("Peer cannot send a direct message to itself")
		}
		message, _ := NewMemP2PMessage(topic, buff, messenger.ID())
		destinationPeer, peerFound := messenger.Network.PeersExceptOne(messenger.ID())[peerID]

		if peerFound == false {
			return errors.New("Destination peer is not connected to the network")
		}

		return destinationPeer.ReceiveMessage(topic, message)
	}

	return errors.New("Peer not connected to network, cannot send anything")
}

// ReceiveMessage handles the received message by passing it to the message
// processor of the corresponding topic, given that this Messenger has
// previously registered a message processor for that topic.
func (messenger *MemP2PMessenger) ReceiveMessage(topic string, message p2p.MessageP2P) error {
	fmt.Printf("memp2p:ReceiveMessage\n")
	messenger.TopicsMutex.Lock()
	validator, found := messenger.Topics[topic]
	messenger.TopicsMutex.Unlock()

	if !found {
		return p2p.ErrNilTopic
	}

	if validator == nil {
		return p2p.ErrNilValidator
	}

	_ = validator.ProcessReceivedMessage(message)

	if messenger.Network.LogMessages {
		messenger.Network.LogMessage(message)
	}
	return nil
}

// Close disconnects this Messenger from the network it was connected to.
func (messenger *MemP2PMessenger) Close() error {
	fmt.Printf("memp2p:Close\n")
	messenger.Network.UnregisterPeer(messenger.ID())
	return nil
}
