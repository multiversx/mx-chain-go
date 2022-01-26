package nodesCoordinator

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/state"
)

// Validator defines a node that can be allocated to a shard for participation in a consensus group as validator
// or block proposer
type Validator interface {
	PubKey() []byte
	Chances() uint32
	Index() uint32
	Size() int
}

// NodesCoordinatorLite defines the minimum behaviour of a struct able to do validator group selection
type NodesCoordinatorLite interface {
	NodesCoordinatorHelper
	PublicKeysSelector
	ComputeConsensusGroup(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []Validator, err error)
	ConsensusGroupSize(uint32) int
	ShardIdForEpoch(epoch uint32) (uint32, error)
	GetNumTotalEligible() uint64
	SetNodesConfig(nodesConfig map[uint32]*EpochNodesConfig)
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

// NodesCoordinatorHelper provides polymorphism functionality for nodesCoordinator
type NodesCoordinatorHelper interface {
	ValidatorsWeights(validators []Validator) ([]uint32, error)
	ComputeAdditionalLeaving(allValidators []*state.ShardValidatorInfo) (map[uint32][]Validator, error)
	GetChance(uint32) uint32
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

// RandomSelector selects randomly a subset of elements from a set of data
type RandomSelector interface {
	Select(randSeed []byte, sampleSize uint32) ([]uint32, error)
	IsInterfaceNil() bool
}

// NodeTypeProviderHandler defines the actions needed for a component that can handle the node type
type NodeTypeProviderHandler interface {
	SetType(nodeType core.NodeType)
	GetType() core.NodeType
	IsInterfaceNil() bool
}

// ArgsUpdateNodes holds the parameters required by the shuffler to generate a new nodes configuration
type ArgsUpdateNodes struct {
	Eligible          map[uint32][]Validator
	Waiting           map[uint32][]Validator
	NewNodes          []Validator
	UnStakeLeaving    []Validator
	AdditionalLeaving []Validator
	Rand              []byte
	NbShards          uint32
	Epoch             uint32
}

// NodesShuffler provides shuffling functionality for nodes
type NodesShuffler interface {
	UpdateParams(numNodesShard uint32, numNodesMeta uint32, hysteresis float32, adaptivity bool)
	UpdateNodeLists(args ArgsUpdateNodes) (*ResUpdateNodes, error)
	IsInterfaceNil() bool
}

// ResUpdateNodes holds the result of the UpdateNodes method
type ResUpdateNodes struct {
	Eligible       map[uint32][]Validator
	Waiting        map[uint32][]Validator
	Leaving        []Validator
	StillRemaining []Validator
}

// ValidatorsDistributor distributes validators across shards
type ValidatorsDistributor interface {
	DistributeValidators(destination map[uint32][]Validator, source map[uint32][]Validator, rand []byte, balanced bool) error
	IsInterfaceNil() bool
}
