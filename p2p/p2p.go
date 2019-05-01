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
	ProcessReceivedMessage(message MessageP2P) error
}

// BroadcastCallbackHandler will be implemented by those message processor instances that need to send back
// a subset of received message (after filtering occurs)
type BroadcastCallbackHandler interface {
	SetBroadcastCallback(callback func(buffToSend []byte))
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
}

// PeerDiscoverer defines the behaviour of a peer discovery mechanism
type PeerDiscoverer interface {
	Bootstrap() error
	Name() string

	ApplyContext(ctxProvider ContextProvider) error
}

// Reconnecter defines the behaviour of a network reconnection mechanism
type Reconnecter interface {
	ReconnectToNetwork() <-chan struct{}
}

// Messenger is the main struct used for communication with other peers
type Messenger interface {
	io.Closer

	ID() PeerID
	Peers() []PeerID

	Addresses() []string
	ConnectToPeer(address string) error
	IsConnected(peerID PeerID) bool
	ConnectedPeers() []PeerID
	ConnectedAddresses() []string
	PeerAddress(pid PeerID) string
	ConnectedPeersOnTopic(topic string) []PeerID
	TrimConnections()
	Bootstrap() error

	CreateTopic(name string, createChannelForTopic bool) error
	HasTopic(name string) bool
	HasTopicValidator(name string) bool
	RegisterMessageProcessor(topic string, handler MessageProcessor) error
	UnregisterMessageProcessor(topic string) error
	OutgoingChannelLoadBalancer() ChannelLoadBalancer
	BroadcastOnChannel(channel string, topic string, buff []byte)
	Broadcast(topic string, buff []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID PeerID) error
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
}

// ChannelLoadBalancer defines what a load balancer that uses chans should do
type ChannelLoadBalancer interface {
	AddChannel(channel string) error
	RemoveChannel(channel string) error
	GetChannelOrDefault(channel string) chan *SendableData
	CollectOneElementFromChannels() *SendableData
}

// DirectSender defines a component that can send direct messages to connected peers
type DirectSender interface {
	NextSeqno(counter *uint64) []byte
	Send(topic string, buff []byte, peer PeerID) error
}

// PeerDiscoveryFactory defines the factory for peer discoverer implementation
type PeerDiscoveryFactory interface {
	CreatePeerDiscoverer() (PeerDiscoverer, error)
}
