package p2p

import (
	"context"
	"io"

	"github.com/mr-tron/base58/base58"
)

// MessageProcessor is the interface used to describe what a receive message processor should do
// All implementations that will be called from Messenger implementation will need to satisfy this interface
// If the function returns a non nil value, the received message will not be propagated to its connected peers
type MessageProcessor interface {
	ProcessReceivedMessage(message MessageP2P, broadcastHandler func(buffToSend []byte)) error
	IsInterfaceNil() bool
}

// BroadcastCallbackHandler will be implemented by those message processor instances that need to send back
// a subset of received message (after filtering occurs)
type BroadcastCallbackHandler interface {
	SetBroadcastCallback(callback func(buffToSend []byte))
	IsInterfaceNil() bool
}

// SendableData represents the struct used in data throttler implementation
type SendableData struct {
	Buff  []byte
	Topic string
}

// PeerID is a p2p peer identity.
type PeerID string

// Bytes returns the peer ID as byte slice
func (pid PeerID) Bytes() []byte {
	return []byte(pid)
}

// Pretty returns a b58-encoded string of the peer id
func (pid PeerID) Pretty() string {
	return base58.Encode(pid.Bytes())
}

// ContextProvider defines an interface for providing context to various messenger components
type ContextProvider interface {
	Context() context.Context
	IsInterfaceNil() bool
}

// PeerDiscoverer defines the behaviour of a peer discovery mechanism
type PeerDiscoverer interface {
	Bootstrap() error
	Name() string

	ApplyContext(ctxProvider ContextProvider) error
	IsInterfaceNil() bool
}

// Reconnecter defines the behaviour of a network reconnection mechanism
type Reconnecter interface {
	ReconnectToNetwork() <-chan struct{}
	Pause()  // pause the peer discovery
	Resume() // resume the peer discovery
	IsInterfaceNil() bool
}

// Messenger is the main struct used for communication with other peers
type Messenger interface {
	io.Closer

	// ID is the Messenger's unique peer identifier across the network (a
	// string). It is derived from the public key of the P2P credentials.
	ID() PeerID

	// Peers is the list of IDs of peers known to the Messenger.
	Peers() []PeerID

	// Addresses is the list of addresses that the Messenger is currently bound
	// to and listening to.
	Addresses() []string

	// ConnectToPeer explicitly connect to a specific peer with a known address (note that the
	// address contains the peer ID). This function is usually not called
	// manually, because any underlying implementation of the Messenger interface
	// should be keeping connections to peers open.
	ConnectToPeer(address string) error

	// IsConnected returns true if the Messenger are connected to a specific peer.
	IsConnected(peerID PeerID) bool

	// ConnectedPeers returns the list of IDs of the peers the Messenger is
	// currently connected to.
	ConnectedPeers() []PeerID

	// ConnectedAddresses returns the list of addresses of the peers to which the
	// Messenger is currently connected.
	ConnectedAddresses() []string

	// PeerAddress builds an address for the given peer ID, e.g.
	// ConnectToPeer(PeerAddress(somePeerID)).
	PeerAddress(pid PeerID) string

	// ConnectedPeersOnTopic returns the IDs of the peers to which the Messenger
	// is currently connected, but filtered by a topic they are registered to.
	ConnectedPeersOnTopic(topic string) []PeerID

	// TrimConnections tries to optimize the number of open connections, closing
	// those that are considered expendable.
	TrimConnections()

	// Bootstrap runs the initialization phase which includes peer discovery,
	// setting up initial connections and self-announcement in the network.
	Bootstrap() error

	// CreateTopic defines a new topic for sending messages, and optionally
	// creates a channel in the LoadBalancer for this topic (otherwise, the topic
	// will use a default channel).
	CreateTopic(name string, createChannelForTopic bool) error

	// HasTopic returns true if the Messenger has declared interest in a topic
	// and it is listening to messages referencing it.
	HasTopic(name string) bool

	// HasTopicValidator returns true if the Messenger has registered a custom
	// validator for a given topic name.
	HasTopicValidator(name string) bool

	// RegisterMessageProcessor adds the provided MessageProcessor to the list
	// of handlers that are invoked whenever a message is received on the
	// specified topic.
	RegisterMessageProcessor(topic string, handler MessageProcessor) error

	// UnregisterMessageProcessor removes the MessageProcessor set by the
	// Messenger from the list of registered handlers for the messages on the
	// given topic.
	UnregisterMessageProcessor(topic string) error

	// OutgoingChannelLoadBalancer returns the ChannelLoadBalancer instance
	// through which the Messenger is sending messages to the network.
	OutgoingChannelLoadBalancer() ChannelLoadBalancer

	// BroadcastOnChannelBlocking asynchronously waits until it can send a
	// message on the channel, but once it is able to, it synchronously sends the
	// message, blocking until sending is completed.
	BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error

	// BroadcastOnChannel asynchronously sends a message on a given topic
	// through a specified channel.
	BroadcastOnChannel(channel string, topic string, buff []byte)

	// Broadcast is a convenience function that calls BroadcastOnChannelBlocking,
	// but implicitly sets the channel to be identical to the specified topic.
	Broadcast(topic string, buff []byte)

	// SendToConnectedPeer asynchronously sends a message to a peer directly,
	// bypassing pubsub and topics. It opens a new connection with the given
	// peer, but reuses a connection and a stream if possible.
	SendToConnectedPeer(topic string, buff []byte, peerID PeerID) error

	IsConnectedToTheNetwork() bool
	ThresholdMinConnectedPeers() int
	SetThresholdMinConnectedPeers(minConnectedPeers int) error

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// MessageP2P defines what a p2p message can do (should return)
type MessageP2P interface {
	From() []byte
	Data() []byte
	SeqNo() []byte
	TopicIDs() []string
	Signature() []byte
	Key() []byte
	Peer() PeerID
	IsInterfaceNil() bool
}

// ChannelLoadBalancer defines what a load balancer that uses chans should do
type ChannelLoadBalancer interface {
	AddChannel(channel string) error
	RemoveChannel(channel string) error
	GetChannelOrDefault(channel string) chan *SendableData
	CollectOneElementFromChannels() *SendableData
	IsInterfaceNil() bool
}

// DirectSender defines a component that can send direct messages to connected peers
type DirectSender interface {
	NextSeqno(counter *uint64) []byte
	Send(topic string, buff []byte, peer PeerID) error
	IsInterfaceNil() bool
}

// PeerDiscoveryFactory defines the factory for peer discoverer implementation
type PeerDiscoveryFactory interface {
	CreatePeerDiscoverer() (PeerDiscoverer, error)
	IsInterfaceNil() bool
}
