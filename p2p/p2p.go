package p2p

import (
	"encoding/hex"
	"io"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/mr-tron/base58/base58"
)

const displayLastPidChars = 12

const (
	// PrioBitsSharder is the variant that uses priority bits
	PrioBitsSharder = "PrioBitsSharder"
	// SimplePrioBitsSharder is the variant that computes the distance without prio bits
	SimplePrioBitsSharder = "SimplePrioBitsSharder"
	// ListsSharder is the variant that uses lists
	ListsSharder = "ListsSharder"
	// OneListSharder is the variant that is shard agnostic and uses one list
	OneListSharder = "OneListSharder"
	// NilListSharder is the variant that will not do connection trimming
	NilListSharder = "NilListSharder"
)

// MessageProcessor is the interface used to describe what a receive message processor should do
// All implementations that will be called from Messenger implementation will need to satisfy this interface
// If the function returns a non nil value, the received message will not be propagated to its connected peers
type MessageProcessor interface {
	ProcessReceivedMessage(message MessageP2P, fromConnectedPeer PeerID) error
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

// PeerDiscoverer defines the behaviour of a peer discovery mechanism
type PeerDiscoverer interface {
	Bootstrap() error
	Name() string
	IsInterfaceNil() bool
}

// Reconnecter defines the behaviour of a network reconnection mechanism
type Reconnecter interface {
	ReconnectToNetwork() <-chan struct{}
	IsInterfaceNil() bool
}

// ReconnecterWithPauseResumeAndWatchdog defines a Reconnecter that supports pausing, resuming and watchdog
type ReconnecterWithPauseResumeAndWatchdog interface {
	Reconnecter
	Pause()  // Pause the peer discovery
	Resume() // Resume the peer discovery

	StartWatchdog(time.Duration) error // StartWatchdog start a discovery resume watchdog
	StopWatchdog() error               // StopWatchdog stops the watchdog
	KickWatchdog() error               // KickWatchdog kicks the watchdog
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

	// UnregisterAllMessageProcessors removes all the MessageProcessor set by the
	// Messenger from the list of registered handlers for the messages on the
	// given topic.
	UnregisterAllMessageProcessors() error

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
	SetPeerShardResolver(peerShardResolver PeerShardResolver) error
	SetPeerBlackListHandler(handler BlacklistHandler) error
	GetConnectedPeersInfo() *ConnectedPeersInfo

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// MessageP2P defines what a p2p message can do (should return)
type MessageP2P interface {
	From() []byte
	Data() []byte
	SeqNo() []byte
	Topics() []string
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

// MessageOriginatorPid will output the message peer id in a pretty format
// If it can, it will display the last displayLastPidChars (12) characters from the pid
func MessageOriginatorPid(msg MessageP2P) string {
	return PeerIdToShortString(msg.Peer())
}

// PeerIdToShortString trims the first displayLastPidChars characters of the provided peer ID after
// converting the peer ID to string using the Pretty functionality
func PeerIdToShortString(pid PeerID) string {
	prettyPid := pid.Pretty()
	lenPrettyPid := len(prettyPid)

	if lenPrettyPid > displayLastPidChars {
		return "..." + prettyPid[lenPrettyPid-displayLastPidChars:]
	}

	return prettyPid
}

// MessageOriginatorSeq will output the sequence number as hex
func MessageOriginatorSeq(msg MessageP2P) string {
	return hex.EncodeToString(msg.SeqNo())
}

// PeerShardResolver is able to resolve the link between the provided PeerID and the shardID
type PeerShardResolver interface {
	GetPeerInfo(pid PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// ConnectedPeersInfo represents the DTO structure used to output the metrics for connected peers
type ConnectedPeersInfo struct {
	UnknownPeers         []string
	IntraShardValidators []string
	IntraShardObservers  []string
	CrossShardValidators []string
	CrossShardObservers  []string
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIdPublicKey(pid PeerID, pk []byte)
	IsInterfaceNil() bool
}

// SignerVerifier is used in higher level protocol authentication of 2 peers after the basic p2p connection has been made
type SignerVerifier interface {
	Sign(message []byte) ([]byte, error)
	Verify(message []byte, sig []byte, pk []byte) error
	PublicKey() []byte
	IsInterfaceNil() bool
}

// Marshalizer defines the 2 basic operations: serialize (marshal) and deserialize (unmarshal)
type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
	IsInterfaceNil() bool
}

// PeerCounts represents the DTO structure used to output the count metrics for connected peers
type PeerCounts struct {
	UnknownPeers    int
	IntraShardPeers int
	CrossShardPeers int
}

// CommonSharder represents the common interface implemented by all sharder implementations
type CommonSharder interface {
	SetPeerShardResolver(psp PeerShardResolver) error
	IsInterfaceNil() bool
}

// BlacklistHandler defines the behavior of a component that is able to decide if a key (peer ID) is black listed or not
//TODO merge this interface with the PeerShardResolver => P2PProtocolHandler ?
//TODO move antiflooding inside network messenger
type BlacklistHandler interface {
	Has(key string) bool
	IsInterfaceNil() bool
}

// ConnectionMonitorWrapper uses a connection monitor but checks if the peer is blacklisted or not
//TODO this should be removed after merging of the PeerShardResolver and BlacklistHandler
type ConnectionMonitorWrapper interface {
	CheckConnectionsBlocking()
	SetBlackListHandler(handler BlacklistHandler) error
	IsInterfaceNil() bool
}
