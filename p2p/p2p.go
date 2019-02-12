package p2p

import (
	"io"

	"github.com/mr-tron/base58/base58"
)

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
	DiscoverNewPeers() error
	IsConnected(peerID PeerID) bool
	ConnectedPeers() []PeerID
	TrimConnections() error

	CreateTopic(name string, createPipeForTopic bool) error
	HasTopic(name string) bool
	HasTopicValidator(name string) bool
	SendDataThrottler() DataThrottler
	BroadcastData(pipe string, topic string, buff []byte)
	SetTopicValidator(topic string, handler func(message MessageP2P) error) error
	SendDirectToConnectedPeer(topic string, buff []byte, peerID PeerID) error
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

// DataThrottler defines what a throttle software device can do
type DataThrottler interface {
	AddPipe(pipe string) error
	RemovePipe(pipe string) error
	GetChannelOrDefault(pipe string) chan *SendableData
	CollectFromPipes() []*SendableData
}

// DirectSender defines a component that can send direct messages to connected peers
type DirectSender interface {
	NextSeqno(counter *uint64) []byte
	SendDirectToConnectedPeer(topic string, buff []byte, peer PeerID) error
}
