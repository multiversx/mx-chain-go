package memp2p

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

const maxQueueSize = 1000

var log = logger.GetOrCreate("p2p/memp2p")

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
	network         *Network
	p2pID           core.PeerID
	address         string
	topics          map[string]struct{}
	topicValidators map[string]p2p.MessageProcessor
	topicsMutex     *sync.RWMutex
	seqNo           uint64
	processQueue    chan p2p.MessageP2P
	numReceived     uint64
}

// NewMessenger constructs a new Messenger that is connected to the
// Network instance provided as argument.
func NewMessenger(network *Network) (*Messenger, error) {
	if network == nil {
		return nil, ErrNilNetwork
	}

	buff := make([]byte, 32)
	_, _ = rand.Reader.Read(buff)
	ID := base64.StdEncoding.EncodeToString(buff)
	Address := fmt.Sprintf("/memp2p/%s", ID)

	messenger := &Messenger{
		network:         network,
		p2pID:           core.PeerID(ID),
		address:         Address,
		topics:          make(map[string]struct{}),
		topicValidators: make(map[string]p2p.MessageProcessor),
		topicsMutex:     &sync.RWMutex{},
		processQueue:    make(chan p2p.MessageP2P, maxQueueSize),
	}
	network.RegisterPeer(messenger)
	go messenger.processFromQueue()

	return messenger, nil
}

// ID returns the P2P ID of the messenger
func (messenger *Messenger) ID() core.PeerID {
	return messenger.p2pID
}

// Peers returns a slice containing the P2P IDs of all the other peers that it
// has knowledge of. Since this is an in-memory network structured as a fully
// connected graph, this function returns the list of the P2P IDs of all the
// peers in the network (assuming this Messenger is connected).
func (messenger *Messenger) Peers() []core.PeerID {
	// If the messenger is connected to the network, it has knowledge of all
	// other peers.
	if !messenger.IsConnectedToNetwork() {
		return []core.PeerID{}
	}
	return messenger.network.PeerIDs()
}

// Addresses returns a list of all the physical addresses that this Messenger
// is bound to and listening to, depending on the available network interfaces
// of the machine. Being an in-memory simulation, the only possible address to
// return is an artificial one, built by the constructor NewMessenger().
func (messenger *Messenger) Addresses() []string {
	addresses := make([]string, 1)
	addresses[0] = messenger.address
	return addresses
}

// ConnectToPeer usually does nothing, because peers connected to the in-memory
// network are already all connected to each other. This function will return
// an error if the Messenger is not connected to the network, though.
func (messenger *Messenger) ConnectToPeer(_ string) error {
	if !messenger.IsConnectedToNetwork() {
		return ErrNotConnectedToNetwork
	}
	// Do nothing, all peers are connected to each other already.
	return nil
}

// IsConnectedToNetwork returns true if this messenger is connected to the
// in-memory network, false otherwise.
func (messenger *Messenger) IsConnectedToNetwork() bool {
	return messenger.network.IsPeerConnected(messenger.ID())
}

// IsConnected returns true if this Messenger is connected to the peer with the
// specified ID. It always returns true if the Messenger is connected to the
// network and false otherwise, regardless of the provided peer ID.
func (messenger *Messenger) IsConnected(_ core.PeerID) bool {
	return messenger.IsConnectedToNetwork()
}

// ConnectedPeers returns a slice of IDs belonging to the peers to which this
// Messenger is connected. If the Messenger is connected to the in₋memory
// network, then the function returns a slice containing the IDs of all the
// other peers connected to the network. Returns false if the Messenger is
// not connected.
func (messenger *Messenger) ConnectedPeers() []core.PeerID {
	if !messenger.IsConnectedToNetwork() {
		return []core.PeerID{}
	}
	return messenger.network.PeerIDsExceptOne(messenger.ID())
}

// ConnectedAddresses returns a slice of peer addresses to which this Messenger
// is connected. If this Messenger is connected to the network, then the
// addresses of all the other peers in the network are returned.
func (messenger *Messenger) ConnectedAddresses() []string {
	if !messenger.IsConnectedToNetwork() {
		return []string{}
	}
	return messenger.network.ListAddressesExceptOne(messenger.ID())
}

// PeerAddresses creates the address string from a given peer ID.
func (messenger *Messenger) PeerAddresses(pid core.PeerID) []string {
	return []string{fmt.Sprintf("/memp2p/%s", string(pid))}
}

// ConnectedPeersOnTopic returns a slice of IDs belonging to the peers in the
// network that have declared their interest in the given topic and are
// listening to messages on that topic.
func (messenger *Messenger) ConnectedPeersOnTopic(topic string) []core.PeerID {
	var filteredPeers []core.PeerID
	if !messenger.IsConnectedToNetwork() {
		return filteredPeers
	}

	allPeersExceptThis := messenger.network.PeersExceptOne(messenger.ID())
	for _, peer := range allPeersExceptThis {
		if peer.HasTopic(topic) {
			filteredPeers = append(filteredPeers, peer.ID())
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
func (messenger *Messenger) CreateTopic(name string, _ bool) error {
	messenger.topicsMutex.Lock()
	defer messenger.topicsMutex.Unlock()

	_, found := messenger.topics[name]
	if found {
		return p2p.ErrTopicAlreadyExists
	}
	messenger.topics[name] = struct{}{}

	return nil
}

// HasTopic returns true if this Messenger has declared interest in the given
// topic; returns false otherwise.
func (messenger *Messenger) HasTopic(name string) bool {
	messenger.topicsMutex.RLock()
	_, found := messenger.topics[name]
	messenger.topicsMutex.RUnlock()

	return found
}

// RegisterMessageProcessor sets the provided message processor to be the
// processor of received messages for the given topic.
func (messenger *Messenger) RegisterMessageProcessor(topic string, _ string, handler p2p.MessageProcessor) error {
	if check.IfNil(handler) {
		return p2p.ErrNilValidator
	}

	messenger.topicsMutex.Lock()
	defer messenger.topicsMutex.Unlock()

	_, found := messenger.topics[topic]
	if !found {
		return fmt.Errorf("%w RegisterMessageProcessor, topic: %s", p2p.ErrNilTopic, topic)
	}

	validator := messenger.topicValidators[topic]
	if !check.IfNil(validator) {
		return p2p.ErrTopicValidatorOperationNotSupported
	}

	messenger.topicValidators[topic] = handler
	return nil
}

// UnregisterMessageProcessor unsets the message processor for the given topic
// (sets it to nil).
func (messenger *Messenger) UnregisterMessageProcessor(topic string, _ string) error {
	messenger.topicsMutex.Lock()
	defer messenger.topicsMutex.Unlock()

	_, found := messenger.topics[topic]
	if !found {
		return fmt.Errorf("%w UnregisterMessageProcessor, topic: %s", p2p.ErrNilTopic, topic)
	}

	validator := messenger.topicValidators[topic]
	if check.IfNil(validator) {
		return p2p.ErrTopicValidatorOperationNotSupported
	}

	messenger.topicValidators[topic] = nil
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
func (messenger *Messenger) BroadcastOnChannelBlocking(_ string, topic string, buff []byte) error {
	return messenger.synchronousBroadcast(topic, buff)
}

// BroadcastOnChannel sends the message to all peers in the network. It calls
// parametricBroadcast() with async=false, which means that peers will have
// their ReceiveMessage() function called synchronously. The call to
// parametricBroadcast() is done as a go-routine, which means this function is,
// in fact, non-blocking, but it is identical with BroadcastOnChannelBlocking()
// in all other regards.
func (messenger *Messenger) BroadcastOnChannel(_ string, topic string, buff []byte) {
	err := messenger.synchronousBroadcast(topic, buff)
	log.LogIfError(err)
}

// Broadcast asynchronously sends the message to all peers in the network. It
// calls parametricBroadcast() with async=true, which means that peers will
// have their ReceiveMessage() function independently called as go-routines.
func (messenger *Messenger) Broadcast(topic string, buff []byte) {
	err := messenger.synchronousBroadcast(topic, buff)
	log.LogIfError(err)
}

// synchronousBroadcast sends a message to all peers in the network in a synchronous way
func (messenger *Messenger) synchronousBroadcast(topic string, data []byte) error {
	if !messenger.IsConnectedToNetwork() {
		return ErrNotConnectedToNetwork
	}

	seqNo := atomic.AddUint64(&messenger.seqNo, 1)
	messageObject := newMessage(topic, data, messenger.ID(), seqNo)

	peers := messenger.network.Peers()
	for _, peer := range peers {
		peer.receiveMessage(messageObject)
	}

	return nil
}

func (messenger *Messenger) processFromQueue() {
	for {
		messageObject := <-messenger.processQueue
		if check.IfNil(messageObject) {
			continue
		}

		topic := messageObject.Topics()[0]
		if topic == "" {
			continue
		}

		messenger.topicsMutex.Lock()
		_, found := messenger.topics[topic]
		if !found {
			messenger.topicsMutex.Unlock()
			continue
		}

		// numReceived gets incremented because the message arrived on a registered topic
		atomic.AddUint64(&messenger.numReceived, 1)
		validator := messenger.topicValidators[topic]
		if check.IfNil(validator) {
			messenger.topicsMutex.Unlock()
			continue
		}
		messenger.topicsMutex.Unlock()

		_ = validator.ProcessReceivedMessage(messageObject, messenger.p2pID)
	}
}

// SendToConnectedPeer sends a message directly to the peer specified by the ID.
func (messenger *Messenger) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if messenger.IsConnectedToNetwork() {
		seqNo := atomic.AddUint64(&messenger.seqNo, 1)
		messageObject := newMessage(topic, buff, messenger.ID(), seqNo)

		receivingPeer, peerFound := messenger.network.Peers()[peerID]
		if !peerFound {
			return ErrReceivingPeerNotConnected
		}

		receivingPeer.receiveMessage(messageObject)

		return nil
	}

	return ErrNotConnectedToNetwork
}

// receiveMessage handles the received message by passing it to the message
// processor of the corresponding topic, given that this Messenger has
// previously registered a message processor for that topic. The Network will
// log the message only if the Network.LogMessages flag is set and only if the
// Messenger has the requested topic and MessageProcessor.
func (messenger *Messenger) receiveMessage(message p2p.MessageP2P) {
	messenger.processQueue <- message
}

// IsConnectedToTheNetwork returns true as this implementation is always connected to its network
func (messenger *Messenger) IsConnectedToTheNetwork() bool {
	return true
}

// SetThresholdMinConnectedPeers does nothing as this implementation is always connected to its network
func (messenger *Messenger) SetThresholdMinConnectedPeers(_ int) error {
	return nil
}

// ThresholdMinConnectedPeers always return 0
func (messenger *Messenger) ThresholdMinConnectedPeers() int {
	return 0
}

// NumMessagesReceived returns the number of messages received
func (messenger *Messenger) NumMessagesReceived() uint64 {
	return atomic.LoadUint64(&messenger.numReceived)
}

// SetPeerShardResolver is a dummy function, not setting anything
func (messenger *Messenger) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// SetPeerDenialEvaluator does nothing
func (messenger *Messenger) SetPeerDenialEvaluator(_ p2p.PeerDenialEvaluator) error {
	return nil
}

// GetConnectedPeersInfo returns a nil object. Not implemented.
func (messenger *Messenger) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	return nil
}

// Close disconnects this Messenger from the network it was connected to.
func (messenger *Messenger) Close() error {
	messenger.network.UnregisterPeer(messenger.ID())
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (messenger *Messenger) IsInterfaceNil() bool {
	return messenger == nil
}
