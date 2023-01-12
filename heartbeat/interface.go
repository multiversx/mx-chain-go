package heartbeat

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
)

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	Broadcast(topic string, buff []byte)
	ID() core.PeerID
	Sign(payload []byte) ([]byte, error)
	ConnectedPeersOnTopic(topic string) []core.PeerID
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
	NotifyTriggerReceivedV2() <-chan struct{}
	CreateData() []byte
	IsInterfaceNil() bool
}

// CurrentBlockProvider can provide the current block that the node was able to commit
type CurrentBlockProvider interface {
	GetCurrentBlockHeader() data.HeaderHandler
	SetCurrentBlockHeaderAndRootHash(bh data.HeaderHandler, rootHash []byte) error
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
	GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetValidatorWithPublicKey(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error)
	IsInterfaceNil() bool
}

// PeerShardMapper saves the shard for a peer ID
type PeerShardMapper interface {
	PutPeerIdShardId(pid core.PeerID, shardID uint32)
	IsInterfaceNil() bool
}

// TrieSyncStatisticsProvider is able to provide trie sync statistics
type TrieSyncStatisticsProvider interface {
	NumProcessed() int
	IsInterfaceNil() bool
}
