package node

import (
	"io"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/update"
)

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	io.Closer
	Bootstrap(numSecondsToWait uint32) error
	Broadcast(topic string, buff []byte)
	BroadcastOnChannel(channel string, topic string, buff []byte)
	BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error
	CreateTopic(name string, createChannelForTopic bool) error
	HasTopic(name string) bool
	HasTopicValidator(name string) bool
	RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error
	PeerAddresses(pid core.PeerID) []string
	IsConnectedToTheNetwork() bool
	ID() core.PeerID
	Peers() []core.PeerID
	IsInterfaceNil() bool
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIdPublicKey(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardId(pk []byte, shardId uint32)
	UpdatePeerIdShardId(pid core.PeerID, shardId uint32)
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	ApplyConsensusSize(size int)
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsInterfaceNil() bool
}

// HardforkTrigger defines the behavior of a hardfork trigger
type HardforkTrigger interface {
	TriggerReceived(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessage() ([]byte, bool)
	Trigger(epoch uint32) error
	CreateData() []byte
	AddCloser(closer update.Closer) error
	NotifyTriggerReceived() <-chan struct{}
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}

// Throttler can monitor the number of the currently running go routines
type Throttler interface {
	CanProcess() bool
	StartProcessing()
	EndProcessing()
	IsInterfaceNil() bool
}
