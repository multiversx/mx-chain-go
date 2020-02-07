package node

import (
	"io"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	io.Closer
	Bootstrap() error
	Broadcast(topic string, buff []byte)
	BroadcastOnChannel(channel string, topic string, buff []byte)
	BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error
	CreateTopic(name string, createChannelForTopic bool) error
	HasTopic(name string) bool
	HasTopicValidator(name string) bool
	RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error
	PeerAddress(pid p2p.PeerID) string
	IsConnectedToTheNetwork() bool
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessageOnTopic(peer p2p.PeerID, topic string) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	IsInterfaceNil() bool
}

// Accumulator defines the interface able to accumulate data and periodically evict them
type Accumulator interface {
	AddData(data interface{})
	OutputChan() <-chan []interface{}
	IsInterfaceNil() bool
}
