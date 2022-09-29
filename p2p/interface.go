package p2p

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	p2p "github.com/ElrondNetwork/elrond-go-p2p"
)

// MessageProcessor is the interface used to describe what a receive message processor should do
// All implementations that will be called from Messenger implementation will need to satisfy this interface
// If the function returns a non nil value, the received message will not be propagated to its connected peers
type MessageProcessor = p2p.MessageProcessor

// SendableData represents the struct used in data throttler implementation
type SendableData struct {
	Buff  []byte
	Topic string
}

// Messenger is the main struct used for communication with other peers
type Messenger = p2p.Messenger

// MessageP2P defines what a p2p message can do (should return)
type MessageP2P = p2p.MessageP2P

// ChannelLoadBalancer defines what a load balancer that uses chans should do
type ChannelLoadBalancer interface {
	AddChannel(channel string) error
	RemoveChannel(channel string) error
	GetChannelOrDefault(channel string) chan *SendableData
	CollectOneElementFromChannels() *SendableData
	Close() error
	IsInterfaceNil() bool
}

// MessageOriginatorPid will output the message peer id in a pretty format
// If it can, it will display the last displayLastPidChars (12) characters from the pid
func MessageOriginatorPid(msg MessageP2P) string {
	return PeerIdToShortString(msg.Peer())
}

// PeerIdToShortString trims the first displayLastPidChars characters of the provided peer ID after
// converting the peer ID to string using the Pretty functionality
func PeerIdToShortString(pid core.PeerID) string {
	return p2p.PeerIdToShortString(pid)
}

// MessageOriginatorSeq will output the sequence number as hex
func MessageOriginatorSeq(msg MessageP2P) string {
	return hex.EncodeToString(msg.SeqNo())
}

// PeerShardResolver is able to resolve the link between the provided PeerID and the shardID
type PeerShardResolver = p2p.PeerShardResolver

// ConnectedPeersInfo represents the DTO structure used to output the metrics for connected peers
type ConnectedPeersInfo = p2p.ConnectedPeersInfo

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32)
	IsInterfaceNil() bool
}

// PreferredPeersHolderHandler defines the behavior of a component able to handle preferred peers operations
type PreferredPeersHolderHandler interface {
	PutConnectionAddress(peerID core.PeerID, address string)
	PutShardID(peerID core.PeerID, shardID uint32)
	Get() map[uint32][]core.PeerID
	Contains(peerID core.PeerID) bool
	Remove(peerID core.PeerID)
	Clear()
	IsInterfaceNil() bool
}

// PeerDenialEvaluator defines the behavior of a component that is able to decide if a peer ID is black listed or not
// TODO merge this interface with the PeerShardResolver => P2PProtocolHandler ?
// TODO move antiflooding inside network messenger
type PeerDenialEvaluator = p2p.PeerDenialEvaluator

// SyncTimer represent an entity able to tell the current time
type SyncTimer interface {
	CurrentTime() time.Time
	IsInterfaceNil() bool
}

// PeersRatingHandler represent an entity able to handle peers ratings
type PeersRatingHandler interface {
	AddPeer(pid core.PeerID)
	IncreaseRating(pid core.PeerID)
	DecreaseRating(pid core.PeerID)
	GetTopRatedPeersFromList(peers []core.PeerID, minNumOfPeersExpected int) []core.PeerID
	IsInterfaceNil() bool
}
