package p2p

import (
	"io"
	"strings"

	"github.com/mr-tron/base58/base58"
)

// PeerDiscoveryType defines the peer discovery mechanism to use
type PeerDiscoveryType int

func (pdt PeerDiscoveryType) String() string {
	switch pdt {
	case PeerDiscoveryOff:
		return "off"
	case PeerDiscoveryKadDht:
		return "kad-dht"
	case PeerDiscoveryMdns:
		return "mdns"
	default:
		return "unknown"
	}
}

// LoadPeerDiscoveryTypeFromString outputs a peer discovery type by parsing the string argument
// Errors if string is not recognized
func LoadPeerDiscoveryTypeFromString(str string) (PeerDiscoveryType, error) {
	str = strings.ToLower(str)

	if str == PeerDiscoveryOff.String() {
		return PeerDiscoveryOff, nil
	}

	if str == PeerDiscoveryMdns.String() {
		return PeerDiscoveryMdns, nil
	}

	if str == PeerDiscoveryKadDht.String() {
		return PeerDiscoveryKadDht, nil
	}

	return PeerDiscoveryOff, ErrPeerDiscoveryNotImplemented
}

const (
	// PeerDiscoveryOff will not enable peer discovery
	PeerDiscoveryOff PeerDiscoveryType = iota
	// PeerDiscoveryMdns will enable mdns mechanism
	PeerDiscoveryMdns
	// PeerDiscoveryKadDht wil enable kad-dht mechanism
	PeerDiscoveryKadDht
)

// MessageProcessor is the interface used to describe what a receive message processor should do
// All implementations that will be called from Messenger implementation will need to satisfy this interface
type MessageProcessor interface {
	ProcessReceivedMessage(message MessageP2P) error
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

// Messenger is the main struct used for communication with other peers
type Messenger interface {
	io.Closer

	ID() PeerID
	Peers() []PeerID

	Addresses() []string
	ConnectToPeer(address string) error
	KadDhtDiscoverNewPeers() error
	IsConnected(peerID PeerID) bool
	ConnectedPeers() []PeerID
	TrimConnections()
	Bootstrap() error

	CreateTopic(name string, createPipeForTopic bool) error
	HasTopic(name string) bool
	HasTopicValidator(name string) bool
	RegisterMessageProcessor(topic string, handler MessageProcessor) error
	UnregisterMessageProcessor(topic string) error
	OutgoingPipeLoadBalancer() PipeLoadBalancer
	BroadcastOnPipe(pipe string, topic string, buff []byte)
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

// PipeLoadBalancer defines what a load balancer that uses pipes should do
type PipeLoadBalancer interface {
	AddPipe(pipe string) error
	RemovePipe(pipe string) error
	GetChannelOrDefault(pipe string) chan *SendableData
	CollectFromPipes() []*SendableData
}

// DirectSender defines a component that can send direct messages to connected peers
type DirectSender interface {
	NextSeqno(counter *uint64) []byte
	Send(topic string, buff []byte, peer PeerID) error
}
