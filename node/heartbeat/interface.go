package heartbeat

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// PeerMessenger defines a subset of the p2p.Messenger interface
type PeerMessenger interface {
	Broadcast(topic string, buff []byte)
	ID() p2p.PeerID
	IsInterfaceNil() bool
}

// MessageHandler defines what a message processor for heartbeat should do
type MessageHandler interface {
	CreateHeartbeatFromP2PMessage(message p2p.MessageP2P) (*Heartbeat, error)
	IsInterfaceNil() bool
}

// EligibleListProvider defines what an eligible list provider should do
type EligibleListProvider interface {
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	IsInterfaceNil() bool
}

//Timer defines an interface for tracking time
type Timer interface {
	Now() time.Time
	IsInterfaceNil() bool
}

// HeartbeatStorageHandler defines what a heartbeat's storer should do
type HeartbeatStorageHandler interface {
	LoadGenesisTime() (time.Time, error)
	UpdateGenesisTime(genesisTime time.Time) error
	LoadHbmiDTO(pubKey string) (*HeartbeatDTO, error)
	SavePubkeyData(pubkey []byte, heartbeat *HeartbeatDTO) error
	LoadKeys() ([][]byte, error)
	SaveKeys(peersSlice [][]byte) error
	IsInterfaceNil() bool
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte)
	UpdatePublicKeyShardId(pk []byte, shardId uint32)
	UpdatePeerIdShardId(pid p2p.PeerID, shardId uint32)
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessagesOnTopic(peer p2p.PeerID, topic string, numMessages uint32) error
	IsInterfaceNil() bool
}

// PeerTypeProviderHandler defines what a component which computes the type of a peer should do
type PeerTypeProviderHandler interface {
	ComputeForPubKey(pubKey []byte) (core.PeerType, uint32, error)
	IsInterfaceNil() bool
}

// HardforkTrigger defines the behavior of a hardfork trigger
type HardforkTrigger interface {
	TriggerReceived(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessage() ([]byte, bool)
	CreateData() []byte
	IsInterfaceNil() bool
}

// BlackListHandler can determine if a certain key is or not blacklisted
type BlackListHandler interface {
	Add(key string) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}
