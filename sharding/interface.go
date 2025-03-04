package sharding

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// Coordinator defines what a shard state coordinator should hold
type Coordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// EpochHandler defines what a component which handles current epoch should be able to do
type EpochHandler interface {
	MetaEpoch() uint32
	IsInterfaceNil() bool
}

// PeerAccountListAndRatingHandler provides Rating Computation Capabilites for the Nodes Coordinator and ValidatorStatistics
type PeerAccountListAndRatingHandler interface {
	// GetChance returns the chances for the rating
	GetChance(uint32) uint32
	// GetStartRating gets the start rating values
	GetStartRating() uint32
	// GetSignedBlocksThreshold gets the threshold for the minimum signed blocks
	GetSignedBlocksThreshold() float32
	// ComputeIncreaseProposer computes the new rating for the increaseLeader
	ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32
	// ComputeDecreaseProposer computes the new rating for the decreaseLeader
	ComputeDecreaseProposer(shardId uint32, currentRating uint32, consecutiveMisses uint32) uint32
	// RevertIncreaseValidator computes the new rating if a revert for increaseProposer should be done
	RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32
	// ComputeIncreaseValidator computes the new rating for the increaseValidator
	ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32
	// ComputeDecreaseValidator computes the new rating for the decreaseValidator
	ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32
	// IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

// GenesisNodesSetupHandler returns the genesis nodes info
type GenesisNodesSetupHandler interface {
	AllInitialNodes() []nodesCoordinator.GenesisNodeInfoHandler
	InitialNodesPubKeys() map[uint32][]string
	GetShardIDForPubKey(pubkey []byte) (uint32, error)
	InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error)
	InitialNodesInfoForShard(shardId uint32) ([]nodesCoordinator.GenesisNodeInfoHandler, []nodesCoordinator.GenesisNodeInfoHandler, error)
	InitialNodesInfo() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
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
	MinShardHysteresisNodes() uint32
	MinMetaHysteresisNodes() uint32
	ExportNodesConfig() config.NodesConfig
	IsInterfaceNil() bool
}

// EpochStartEventNotifier provides Register and Unregister functionality for the end of epoch events
type EpochStartEventNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	IsInterfaceNil() bool
}

// ChainParametersHandler defines the actions that need to be done by a component that can handle chain parameters
type ChainParametersHandler interface {
	CurrentChainParameters() config.ChainParametersByEpochConfig
	ChainParametersForEpoch(epoch uint32) (config.ChainParametersByEpochConfig, error)
	IsInterfaceNil() bool
}

// ChainParametersNotifierHandler defines the actions that need to be done by a component that can handle chain parameters changes
type ChainParametersNotifierHandler interface {
	UpdateCurrentChainParameters(params config.ChainParametersByEpochConfig)
	IsInterfaceNil() bool
}
