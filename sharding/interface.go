package sharding

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
)

type validator = nodesCoordinator.Validator

// Coordinator defines what a shard state coordinator should hold
type Coordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinator interface {
	nodesCoordinator.NodesCoordinatorLite
	ShuffleOutForEpoch(_ uint32)
	LoadState(key []byte) error
	GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error)
	GetValidatorWithPublicKey(publicKey []byte) (validator validator, shardId uint32, err error)
	GetSavedStateKey() []byte
}

// EpochHandler defines what a component which handles current epoch should be able to do
type EpochHandler interface {
	MetaEpoch() uint32
	IsInterfaceNil() bool
}

//PeerAccountListAndRatingHandler provides Rating Computation Capabilites for the Nodes Coordinator and ValidatorStatistics
type PeerAccountListAndRatingHandler interface {
	//GetChance returns the chances for the the rating
	GetChance(uint32) uint32
	//GetStartRating gets the start rating values
	GetStartRating() uint32
	//GetSignedBlocksThreshold gets the threshold for the minimum signed blocks
	GetSignedBlocksThreshold() float32
	//ComputeIncreaseProposer computes the new rating for the increaseLeader
	ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32
	//ComputeDecreaseProposer computes the new rating for the decreaseLeader
	ComputeDecreaseProposer(shardId uint32, currentRating uint32, consecutiveMisses uint32) uint32
	//RevertIncreaseValidator computes the new rating if a revert for increaseProposer should be done
	RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32
	//ComputeIncreaseValidator computes the new rating for the increaseValidator
	ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32
	//ComputeDecreaseValidator computes the new rating for the decreaseValidator
	ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

//ChanceComputer provides chance computation capabilities based on a rating
type ChanceComputer interface {
	//GetChance returns the chances for the the rating
	GetChance(uint32) uint32
	//IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

// ShuffledOutHandler defines the methods needed for the computation of a shuffled out event
type ShuffledOutHandler interface {
	Process(newShardID uint32) error
	RegisterHandler(handler func(newShardID uint32))
	CurrentShardID() uint32
	IsInterfaceNil() bool
}

// EpochStartActionHandler defines the action taken on epoch start event
type EpochStartActionHandler interface {
	EpochStartAction(hdr data.HeaderHandler)
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyOrder() uint32
}

// GenesisNodesSetupHandler returns the genesis nodes info
type GenesisNodesSetupHandler interface {
	AllInitialNodes() []GenesisNodeInfoHandler
	InitialNodesPubKeys() map[uint32][]string
	GetShardIDForPubKey(pubkey []byte) (uint32, error)
	InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error)
	InitialNodesInfoForShard(shardId uint32) ([]GenesisNodeInfoHandler, []GenesisNodeInfoHandler, error)
	InitialNodesInfo() (map[uint32][]GenesisNodeInfoHandler, map[uint32][]GenesisNodeInfoHandler)
	GetStartTime() int64
	GetRoundDuration() uint64
	GetShardConsensusGroupSize() uint32
	GetMetaConsensusGroupSize() uint32
	NumberOfShards() uint32
	MinNumberOfNodes() uint32
	MinNumberOfShardNodes() uint32
	MinNumberOfMetaNodes() uint32
	GetHysteresis() float32
	GetAdaptivity() bool
	MinNumberOfNodesWithHysteresis() uint32
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
