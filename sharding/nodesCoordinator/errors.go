package nodesCoordinator

import (
	"errors"
)

// ErrInvalidNumberOfShards signals that an invalid number of shards was passed to the sharding registry
var ErrInvalidNumberOfShards = errors.New("the number of shards must be greater than zero")

// ErrInvalidShardId signals that an invalid shard is was passed
var ErrInvalidShardId = errors.New("shard id must be smaller than the total number of shards")

// ErrNilPubKey signals that the public key is nil
var ErrNilPubKey = errors.New("nil public key")

// ErrInvalidNumberPubKeys signals that an invalid number of public keys was used
var ErrInvalidNumberPubKeys = errors.New("invalid number of public keys")

// ErrNilNodesCoordinator signals that the nodesCoordinator is nil
var ErrNilNodesCoordinator = errors.New("nil nodesCoordinator")

// ErrNilMarshalizer signals that the marshalizer is nil
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to set nil pubkey converter")

// ErrNilInputNodesMap signals that a nil nodes map was provided
var ErrNilInputNodesMap = errors.New("nil input nodes map")

// ErrSmallShardEligibleListSize signals that the eligible validators list's size is less than the consensus size
var ErrSmallShardEligibleListSize = errors.New("small shard eligible list size")

// ErrSmallMetachainEligibleListSize signals that the eligible validators list's size is less than the consensus size
var ErrSmallMetachainEligibleListSize = errors.New("small metachain eligible list size")

// ErrMapSizeZero signals that there are no elements in the map
var ErrMapSizeZero = errors.New("map size zero")

// ErrNilPreviousEpochConfig signals that the previous epoch config is nil
var ErrNilPreviousEpochConfig = errors.New("nil previous epoch config")

// ErrEpochNodesConfigDoesNotExist signals that the epoch nodes configuration is missing
var ErrEpochNodesConfigDoesNotExist = errors.New("epoch nodes configuration does not exist")

// ErrInvalidConsensusGroupSize signals that the consensus size is invalid (e.g. value is negative)
var ErrInvalidConsensusGroupSize = errors.New("invalid consensus group size")

// ErrNilRandomness signals that a nil randomness source has been provided
var ErrNilRandomness = errors.New("nil randomness source")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")

// ErrNilShuffler signals that a nil shuffler was provided
var ErrNilShuffler = errors.New("nil nodes shuffler provided")

// ErrNilBootStorer signals that a nil boot storer was provided
var ErrNilBootStorer = errors.New("nil boot storer provided")

// ErrValidatorNotFound signals that the validator has not been found
var ErrValidatorNotFound = errors.New("validator not found")

// ErrNilWeights signals that nil weights list was provided
var ErrNilWeights = errors.New("nil weights")

// ErrNotImplemented signals a call of a non implemented functionality
var ErrNotImplemented = errors.New("feature not implemented")

// ErrNilCacher signals that a nil cacher has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrInvalidSampleSize signals that an invalid sample size was provided
var ErrInvalidSampleSize = errors.New("invalid sample size")

// ErrInvalidWeight signals an invalid weight was provided
var ErrInvalidWeight = errors.New("invalid weight")

// ErrNilRandomSelector signals that a nil selector was provided
var ErrNilRandomSelector = errors.New("nil selector")

// ErrNilChanceComputer signals that a nil chance computer was provided
var ErrNilChanceComputer = errors.New("nil chance computer")

// ErrWrongTypeAssertion signals wrong type assertion error
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilBlockBody signals that block body is nil
var ErrNilBlockBody = errors.New("nil block body")

// ErrNilShuffledOutHandler signals that a nil shuffled out handler has been provided
var ErrNilShuffledOutHandler = errors.New("nil shuffled out handler")

// ErrNilEpochNotifier signals that the provided epoch notifier is nil
var ErrNilEpochNotifier = errors.New("nil epoch notifier")

// ErrNilEndOfProcessingHandler signals that a nil end of processing handler has been provided
var ErrNilEndOfProcessingHandler = errors.New("nil end of processing handler")

// ErrNilOrEmptyDestinationForDistribute signals that a nil or empty value was provided for destination of distributedNodes
var ErrNilOrEmptyDestinationForDistribute = errors.New("nil or empty destination list for distributeNodes")

// ErrNilNodeShufflerArguments signals that a nil argument pointer was provided for creating the nodes shuffler instance
var ErrNilNodeShufflerArguments = errors.New("nil arguments for the creation of a node shuffler")

// ErrNilNodeStopChannel signals that a nil node stop channel has been provided
var ErrNilNodeStopChannel = errors.New("nil node stop channel")

// ErrValidatorCannotBeFullArchive signals a configuration issue because a validator cannot be a full archive node
var ErrValidatorCannotBeFullArchive = errors.New("validator cannot be a full archive node")

// ErrNilNodeTypeProvider signals that a nil node type provider has been given
var ErrNilNodeTypeProvider = errors.New("nil node type provider")
