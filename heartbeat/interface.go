package heartbeat

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/common"
	heartbeatData "github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
)

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	Broadcast(topic string, buff []byte)
	ID() core.PeerID
	Sign(payload []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// MessageHandler defines what a message processor for heartbeat should do
type MessageHandler interface {
	CreateHeartbeatFromP2PMessage(message p2p.MessageP2P) (*heartbeatData.Heartbeat, error)
	IsInterfaceNil() bool
}

// EligibleListProvider defines what an eligible list provider should do
type EligibleListProvider interface {
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	IsInterfaceNil() bool
}

// Timer defines an interface for tracking time
type Timer interface {
	Now() time.Time
	IsInterfaceNil() bool
}

// HeartbeatStorageHandler defines what a heartbeat's storer should do
type HeartbeatStorageHandler interface {
	LoadGenesisTime() (time.Time, error)
	UpdateGenesisTime(genesisTime time.Time) error
	LoadHeartBeatDTO(pubKey string) (*heartbeatData.HeartbeatDTO, error)
	SavePubkeyData(pubkey []byte, heartbeat *heartbeatData.HeartbeatDTO) error
	LoadKeys() ([][]byte, error)
	SaveKeys(peersSlice [][]byte) error
	IsInterfaceNil() bool
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32)
	PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType)
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsInterfaceNil() bool
}

// PeerTypeProviderHandler defines what a component which computes the type of a peer should do
type PeerTypeProviderHandler interface {
	ComputeForPubKey(pubKey []byte) (common.PeerType, uint32, error)
	GetAllPeerTypeInfos() []*state.PeerTypeInfo
	IsInterfaceNil() bool
}

// HardforkTrigger defines the behavior of a hardfork trigger
type HardforkTrigger interface {
	TriggerReceived(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessage() ([]byte, bool)
	NotifyTriggerReceived() <-chan struct{} // TODO: remove it with heartbeat v1 cleanup
	NotifyTriggerReceivedV2() <-chan struct{}
	CreateData() []byte
	IsInterfaceNil() bool
}

// PeerBlackListHandler can determine if a certain peer ID is or not blacklisted
type PeerBlackListHandler interface {
	Add(pid core.PeerID) error
	Has(pid core.PeerID) bool
	Sweep()
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessor is the interface for consensus participation statistics
type ValidatorStatisticsProcessor interface {
	RootHash() ([]byte, error)
	GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
	IsInterfaceNil() bool
}

// CurrentBlockProvider can provide the current block that the node was able to commit
type CurrentBlockProvider interface {
	GetCurrentBlockHeader() data.HeaderHandler
	IsInterfaceNil() bool
}

// NodeRedundancyHandler defines the interface responsible for the redundancy management of the node
type NodeRedundancyHandler interface {
	IsRedundancyNode() bool
	IsMainMachineActive() bool
	ObserverPrivateKey() crypto.PrivateKey
	IsInterfaceNil() bool
}

// NodesCoordinator defines the behavior of a struct able to do validator selection
type NodesCoordinator interface {
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetValidatorWithPublicKey(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error)
	IsInterfaceNil() bool
}
