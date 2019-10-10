package memp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var log = logger.DefaultLogger()

// Messenger is an implementation of the p2p.Messenger interface that
// uses no real networking code, but instead connects to a network simulated in
// memory (the Network struct). The Messenger is intended for use
// in automated tests instead of the real libp2p, in order to speed up their
// execution and reduce resource usage.
//
// All message-sending functions imitate the synchronous/asynchronous
// behavior of the Messenger struct originally implemented for libp2p. Note
// that the Network ensures that all messengers are connected to all
// other messengers, thus when a Messenger is connected to an in-memory
// network, it reports being connected to all the nodes. Consequently,
// broadcasting a message will be received by all the messengers in the
// network.
type Messenger struct {
	Network     *Network
	P2PID       p2p.PeerID
	Address     string
	Topics      map[string]p2p.MessageProcessor
	TopicsMutex sync.RWMutex
}

// NewMessenger constructs a new Messenger that is connected to the
// Network instance provided as argument.
func NewMessenger(network *Network) (*Messenger, error) {
	if network == nil {
		return nil, ErrNilNetwork
	}

	ID := fmt.Sprintf("Peer%d", len(network.PeerIDs())+1)
	Address := fmt.Sprintf("/memp2p/%s", ID)

	messenger := &Messenger{
		Network:     network,
		P2PID:       p2p.PeerID(ID),
		Address:     Address,
		Topics:      make(map[string]p2p.MessageProcessor),
		TopicsMutex: sync.RWMutex{},
	}

	network.RegisterPeer(messenger)

	return messenger, nil
}

// ID returns the P2P ID of the messenger
func (messenger *Messenger) ID() p2p.PeerID {
	return messenger.P2PID
}

// Peers returns a slice containing the P2P IDs of all the other peers that it
// has knowledge of. Since this is an in-memory network structured as a fully
// connected graph, this function returns the list of the P2P IDs of all the
// peers in the network (assuming this Messenger is connected).
func (messenger *Messenger) Peers() []p2p.PeerID {
	// If the messenger is connected to the network, it has knowledge of all
	// other peers.
	if !messenger.IsConnectedToNetwork() {
		return []p2p.PeerID{}
	}
	return messenger.Network.PeerIDs()
}

// Addresses returns a list of all the physical addresses that this Messenger
// is bound to and listening to, depending on the available network interfaces
// of the machine. Being an in-memory simulation, the only possible address to
// return is an artificial one, built by the constructor NewMessenger().
func (messenger *Messenger) Addresses() []string {
	addresses := make([]string, 1)
	addresses[0] = messenger.Address
	return addresses
}

// ConnectToPeer usually does nothing, because peers connected to the in-memory
// network are already all connected to each other. This function will return
// an error if the Messenger is not connected to the network, though.
func (messenger *Messenger) ConnectToPeer(address string) error {
	if !messenger.IsConnectedToNetwork() {
		return ErrNotConnectedToNetwork
	}
	// Do nothing, all peers are connected to each other already.
	return nil
}

// IsConnectedToNetwork returns true if this messenger is connected to the
// in-memory network, false otherwise.
func (messenger *Messenger) IsConnectedToNetwork() bool {
	return messenger.Network.IsPeerConnected(messenger.ID())
}

// IsConnected returns true if this Messenger is connected to the peer with the
// specified ID. It always returns true if the Messenger is connected to the
// network and false otherwise, regardless of the provided peer ID.
func (messenger *Messenger) IsConnected(peerID p2p.PeerID) bool {
	return messenger.IsConnectedToNetwork()
}

// ConnectedPeers returns a slice of IDs belonging to the peers to which this
// Messenger is connected. If the Messenger is connected to the in₋memory
// network, then the function returns a slice containing the IDs of all the
// other peers connected to the network. Returns false if the Messenger is
// not connected.
func (messenger *Messenger) ConnectedPeers() []p2p.PeerID {
	if !messenger.IsConnectedToNetwork() {
		return []p2p.PeerID{}
	}
	return messenger.Network.PeerIDsExceptOne(messenger.ID())
}

// ConnectedAddresses returns a slice of peer addresses to which this Messenger
// is connected. If this Messenger is connected to the network, then the
// addresses of all the other peers in the network are returned.
func (messenger *Messenger) ConnectedAddresses() []string {
	if !messenger.IsConnectedToNetwork() {
		return []string{}
	}
	return messenger.Network.ListAddressesExceptOne(messenger.ID())
}

// PeerAddress creates the address string from a given peer ID.
func (messenger *Messenger) PeerAddress(pid p2p.PeerID) string {
	return fmt.Sprintf("/memp2p/%s", string(pid))
}

// ConnectedPeersOnTopic returns a slice of IDs belonging to the peers in the
// network that have declared their interest in the given topic and are
// listening to messages on that topic.
func (messenger *Messenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	var filteredPeers []p2p.PeerID
	if !messenger.IsConnectedToNetwork() {
		return filteredPeers
	}

	allPeerIDsExceptThis := messenger.Network.PeerIDsExceptOne(messenger.ID())
	allPeersExceptThis := messenger.Network.PeersExceptOne(messenger.ID())
	for _, peerID := range allPeerIDsExceptThis {
		peer := allPeersExceptThis[peerID]
		if peer.HasTopic(topic) {
			filteredPeers = append(filteredPeers, peerID)
		}
	}

	return filteredPeers
}

// TrimConnections does nothing, as it is not applicable to the in-memory
// messenger.
func (messenger *Messenger) TrimConnections() {
}

// Bootstrap does nothing, as it is not applicable to the in-memory messenger.
func (messenger *Messenger) Bootstrap() error {
	return nil
}

// CreateTopic adds the topic provided as argument to the list of topics of
// interest for this Messenger. It also registers a nil message validator to
// handle the messages received on this topic.
func (messenger *Messenger) CreateTopic(name string, createChannelForTopic bool) error {
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
func (messenger *Messenger) HasTopic(name string) bool {
	messenger.TopicsMutex.RLock()
	_, found := messenger.Topics[name]
	messenger.TopicsMutex.RUnlock()

	return found
}

// HasTopicValidator returns true if this Messenger has declared interest in
// the given topic and has registered a non-nil validator on that topic.
// Returns false otherwise.
func (messenger *Messenger) HasTopicValidator(name string) bool {
	messenger.TopicsMutex.RLock()
	validator, found := messenger.Topics[name]
	messenger.TopicsMutex.RUnlock()

	return found && (validator != nil)
}

// RegisterMessageProcessor sets the provided message processor to be the
// processor of received messages for the given topic.
func (messenger *Messenger) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if handler == nil || handler.IsInterfaceNil() {
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
func (messenger *Messenger) UnregisterMessageProcessor(topic string) error {
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
func (messenger *Messenger) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return nil
}

// BroadcastOnChannelBlocking sends the message to all peers in the network. It
// calls parametricBroadcast() with async=false, which means that peers will
// have their ReceiveMessage() function called synchronously. The call
// to parametricBroadcast() is done synchronously as well. This function should
// be called as a go-routine.
func (messenger *Messenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return messenger.parametricBroadcast(topic, buff, false)
}

// BroadcastOnChannel sends the message to all peers in the network. It calls
// parametricBroadcast() with async=false, which means that peers will have
// their ReceiveMessage() function called synchronously. The call to
// parametricBroadcast() is done as a go-routine, which means this function is,
// in fact, non-blocking, but it is identical with BroadcastOnChannelBlocking()
// in all other regards.
func (messenger *Messenger) BroadcastOnChannel(channel string, topic string, buff []byte) {
	err := messenger.parametricBroadcast(topic, buff, false)
	log.LogIfError(err)
}

// Broadcast asynchronously sends the message to all peers in the network. It
// calls parametricBroadcast() with async=true, which means that peers will
// have their ReceiveMessage() function independently called as go-routines.
func (messenger *Messenger) Broadcast(topic string, buff []byte) {
	err := messenger.parametricBroadcast(topic, buff, true)
	log.LogIfError(err)
}

// parametricBroadcast sends a message to all peers in the network, with the
// possibility to choose from asynchronous or synchronous sending.
func (messenger *Messenger) parametricBroadcast(topic string, data []byte, async bool) error {
	if !messenger.IsConnectedToNetwork() {
		return ErrNotConnectedToNetwork
	}

	message, err := NewMessage(topic, data, messenger.ID())
	if err != nil {
		return err
	}

	for _, peer := range messenger.Network.Peers() {
		if async {
			go func(receivingPeer *Messenger) {
				err := receivingPeer.ReceiveMessage(topic, message)
				log.LogIfError(err)
			}(peer)
		} else {
			err = peer.ReceiveMessage(topic, message)
		}
		if err != nil {
			break
		}
	}

	return err
}

// SendToConnectedPeer sends a message directly to the peer specified by the ID.
func (messenger *Messenger) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	if messenger.IsConnectedToNetwork() {
		if peerID == messenger.ID() {
			return ErrCannotSendToSelf
		}
		message, err := NewMessage(topic, buff, messenger.ID())
		if err != nil {
			return err
		}
		receivingPeer, peerFound := messenger.Network.PeersExceptOne(messenger.ID())[peerID]

		if !peerFound {
			return ErrReceivingPeerNotConnected
		}

		return receivingPeer.ReceiveMessage(topic, message)
	}

	return ErrNotConnectedToNetwork
}

// ReceiveMessage handles the received message by passing it to the message
// processor of the corresponding topic, given that this Messenger has
// previously registered a message processor for that topic. The Network will
// log the message only if the Network.LogMessages flag is set and only if the
// Messenger has the requested topic and MessageProcessor.
func (messenger *Messenger) ReceiveMessage(topic string, message p2p.MessageP2P) error {
	messenger.TopicsMutex.Lock()
	validator, found := messenger.Topics[topic]
	messenger.TopicsMutex.Unlock()

	if !found {
		return p2p.ErrNilTopic
	}

	if validator == nil {
		return p2p.ErrNilValidator
	}

	if messenger.Network.LogMessages {
		messenger.Network.LogMessage(message)
	}

	err := validator.ProcessReceivedMessage(message)

	return err
}

// Close disconnects this Messenger from the network it was connected to.
func (messenger *Messenger) Close() error {
	messenger.Network.UnregisterPeer(messenger.ID())
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (messenger *Messenger) IsInterfaceNil() bool {
	if messenger == nil {
		return true
	}
	return false
}
