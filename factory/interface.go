package factory

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	IsInterfaceNil() bool
}

// NodesSetupHandler defines which actions should be done for handling initial nodes setup
type NodesSetupHandler interface {
	InitialNodesPubKeys() map[uint32][]string
	InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error)
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessagesOnTopic(peer p2p.PeerID, topic string, numMessages uint32) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	ApplyConsensusSize(size int)
	IsInterfaceNil() bool
}
