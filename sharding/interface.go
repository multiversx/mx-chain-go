package sharding

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// Coordinator defines what a shard state coordinator should hold
type Coordinator interface {
	NumberOfShards() uint32
	ComputeId(address state.AddressContainer) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress state.AddressContainer) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// Validator defines a node that can be allocated to a shard for participation in a consensus group as validator
// or block proposer
type Validator interface {
	PubKey() []byte
	Address() []byte
}

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinator interface {
	PublicKeysSelector
	SetNodesPerShards(eligible map[uint32][]Validator, waiting map[uint32][]Validator, epoch uint32, updatePeers bool) error
	ComputeConsensusGroup(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []Validator, err error)
	GetValidatorWithPublicKey(publicKey []byte, epoch uint32) (validator Validator, shardId uint32, err error)
	UpdatePeersListAndIndex() error
	LoadState(key []byte) error
	GetSavedStateKey() []byte
	ShardIdForEpoch(epoch uint32) (uint32, error)
	GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error)
	ConsensusGroupSize(uint32) int
	GetNumTotalEligible() uint64
	IsInterfaceNil() bool
}

// PublicKeysSelector allows retrieval of eligible validators public keys
type PublicKeysSelector interface {
	GetValidatorsIndexes(publicKeys []string, epoch uint32) ([]uint64, error)
	GetEligiblePublicKeysPerShard(epoch uint32) (map[uint32][][]byte, error)
	GetWaitingPublicKeysPerShard(epoch uint32) (map[uint32][][]byte, error)
	GetSelectedPublicKeys(selection []byte, shardId uint32, epoch uint32) (publicKeys []string, err error)
	GetConsensusValidatorsPublicKeys(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetConsensusValidatorsRewardsAddresses(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetOwnPublicKey() []byte
}

// EpochHandler defines what a component which handles current epoch should be able to do
type EpochHandler interface {
	Epoch() uint32
	IsInterfaceNil() bool
}

// ArgsUpdateNodes holds the parameters required by the shuffler to generate a new nodes configuration
type ArgsUpdateNodes struct {
	Eligible map[uint32][]Validator
	Waiting  map[uint32][]Validator
	NewNodes []Validator
	Leaving  []Validator
	Rand     []byte
	NbShards uint32
}

// NodesShuffler provides shuffling functionality for nodes
type NodesShuffler interface {
	UpdateParams(numNodesShard uint32, numNodesMeta uint32, hysteresis float32, adaptivity bool)
	UpdateNodeLists(args ArgsUpdateNodes) (map[uint32][]Validator, map[uint32][]Validator, []Validator)
	IsInterfaceNil() bool
}

//PeerAccountListAndRatingHandler provides Rating Computation Capabilites for the Nodes Coordinator and ValidatorStatistics
type PeerAccountListAndRatingHandler interface {
	RatingReader
	// UpdateListAndIndex updated the list and the index for a peer
	UpdateListAndIndex(pubKey string, shardID uint32, list string, index int) error
	//GetStartRating gets the start rating values
	GetStartRating() uint32
	//ComputeIncreaseProposer computes the new rating for the increaseLeader
	ComputeIncreaseProposer(val uint32) uint32
	//ComputeDecreaseProposer computes the new rating for the decreaseLeader
	ComputeDecreaseProposer(val uint32) uint32
	//ComputeIncreaseValidator computes the new rating for the increaseValidator
	ComputeIncreaseValidator(val uint32) uint32
	//ComputeDecreaseValidator computes the new rating for the decreaseValidator
	ComputeDecreaseValidator(val uint32) uint32
}

// ListIndexUpdaterHandler defines what a component which can update the list and index for a peer should do
type ListIndexUpdaterHandler interface {
	// UpdateListAndIndex updated the list and the index for a peer
	UpdateListAndIndex(pubKey string, shardID uint32, list string, index int) error
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

// ListIndexUpdaterSetter provides the capabilities to set a ListIndexUpdater
type ListIndexUpdaterSetter interface {
	// SetListIndexUpdater will set the updater
	SetListIndexUpdater(updater ListIndexUpdaterHandler)
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//RatingReader provides rating reading capabilities for the ratingHandler
type RatingReader interface {
	//GetRating gets the rating for the public key
	GetRating(string) uint32
	//GetRatings gets all the ratings as a map[pk] ratingValue
	GetRatings([]string) map[string]uint32
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//RatingReaderSetter provides the capabilities to set a RatingReader
type RatingReaderSetter interface {
	//SetRatingReader sets the rating
	SetRatingReader(RatingReader)
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//Cacher provides the capabilities needed to store and retrieve information needed in the NodesCoordinator
type Cacher interface {
	// Put adds a value to the cache.  Returns true if an eviction occurred.
	Put(key []byte, value interface{}) (evicted bool)
	// Get looks up a key's value from the cache.
	Get(key []byte) (value interface{}, ok bool)
}
