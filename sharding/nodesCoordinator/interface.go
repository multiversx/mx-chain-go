package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

// Validator defines a node that can be allocated to a shard for participation in a consensus group as validator
// or block proposer
type Validator interface {
	PubKey() []byte
	Chances() uint32
	Index() uint32
	Size() int
}

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinator interface {
	NodesCoordinatorHelper
	PublicKeysSelector
	ComputeConsensusGroup(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []Validator, err error)
	GetValidatorWithPublicKey(publicKey []byte) (validator Validator, shardId uint32, err error)
	LoadState(key []byte) error
	GetSavedStateKey() []byte
	ShardIdForEpoch(epoch uint32) (uint32, error)
	ShuffleOutForEpoch(_ uint32)
	GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error)
	ConsensusGroupSize(uint32) int
	GetNumTotalEligible() uint64
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NodesCoordinatorToRegistry() *NodesCoordinatorRegistry
	IsInterfaceNil() bool
}

// EpochStartEventNotifier provides Register and Unregister functionality for the end of epoch events
type EpochStartEventNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	IsInterfaceNil() bool
}

// PublicKeysSelector allows retrieval of eligible validators public keys
type PublicKeysSelector interface {
	GetValidatorsIndexes(publicKeys []string, epoch uint32) ([]uint64, error)
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetConsensusValidatorsPublicKeys(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetOwnPublicKey() []byte
}

// NodesShuffler provides shuffling functionality for nodes
type NodesShuffler interface {
	UpdateParams(numNodesShard uint32, numNodesMeta uint32, hysteresis float32, adaptivity bool)
	UpdateNodeLists(args ArgsUpdateNodes) (*ResUpdateNodes, error)
	IsInterfaceNil() bool
}

// NodesCoordinatorHelper provides polymorphism functionality for nodesCoordinator
type NodesCoordinatorHelper interface {
	ValidatorsWeights(validators []Validator) ([]uint32, error)
	ComputeAdditionalLeaving(allValidators []*state.ShardValidatorInfo) (map[uint32][]Validator, error)
	GetChance(uint32) uint32
}

//ChanceComputer provides chance computation capabilities based on a rating
type ChanceComputer interface {
	//GetChance returns the chances for the rating
	GetChance(uint32) uint32
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//Cacher provides the capabilities needed to store and retrieve information needed in the NodesCoordinator
type Cacher interface {
	// Clear is used to completely clear the cache.
	Clear()
	// Put adds a value to the cache.  Returns true if an eviction occurred.
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	// Get looks up a key's value from the cache.
	Get(key []byte) (value interface{}, ok bool)
}

// ShuffledOutHandler defines the methods needed for the computation of a shuffled out event
type ShuffledOutHandler interface {
	Process(newShardID uint32) error
	RegisterHandler(handler func(newShardID uint32))
	CurrentShardID() uint32
	IsInterfaceNil() bool
}

// RandomSelector selects randomly a subset of elements from a set of data
type RandomSelector interface {
	Select(randSeed []byte, sampleSize uint32) ([]uint32, error)
	IsInterfaceNil() bool
}

// EpochStartActionHandler defines the action taken on epoch start event
type EpochStartActionHandler interface {
	EpochStartAction(hdr data.HeaderHandler)
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyOrder() uint32
}

// NodeTypeProviderHandler defines the actions needed for a component that can handle the node type
type NodeTypeProviderHandler interface {
	SetType(nodeType core.NodeType)
	GetType() core.NodeType
	IsInterfaceNil() bool
}

// GenesisNodeInfoHandler defines the public methods for the genesis nodes info
type GenesisNodeInfoHandler interface {
	AssignedShard() uint32
	AddressBytes() []byte
	PubKeyBytes() []byte
	GetInitialRating() uint32
	IsInterfaceNil() bool
}

// ValidatorsDistributor distributes validators across shards
type ValidatorsDistributor interface {
	DistributeValidators(destination map[uint32][]Validator, source map[uint32][]Validator, rand []byte, balanced bool) error
	IsInterfaceNil() bool
}

// EpochsConfigUpdateHandler specifies the behaviour needed to update nodes config epochs
type EpochsConfigUpdateHandler interface {
	NodesCoordinator
	SetNodesConfigFromValidatorsInfo(epoch uint32, randomness []byte, validatorsInfo []*state.ShardValidatorInfo) error
	IsEpochInConfig(epoch uint32) bool
}

// NodesCoordinatorWithRaterFactory should create a nodes coordinator with rater
type NodesCoordinatorWithRaterFactory interface {
	CreateNodesCoordinatorWithRater(args *NodesCoordinatorWithRaterArgs) (NodesCoordinator, error)
}
