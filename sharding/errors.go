package sharding

import (
	"errors"
)

// ErrInvalidNumberOfShards signals that an invalid number of shards was passed to the sharding registry
var ErrInvalidNumberOfShards = errors.New("the number of shards must be greater than zero")

// ErrInvalidShardId signals that an invalid shard is was passed
var ErrInvalidShardId = errors.New("shard id must be smaller than the total number of shards")

// ErrShardIdOutOfRange signals an error when shard id is out of range
var ErrShardIdOutOfRange = errors.New("shard id out of range")

// ErrNilPubKey signals that the public key is nil
var ErrNilPubKey = errors.New("nil public key")

// ErrInvalidNumberPubKeys signals that an invalid number of public keys was used
var ErrInvalidNumberPubKeys = errors.New("invalid number of public keys")

// ErrNilNodesCoordinator signals that the nodesCoordinator is nil
var ErrNilNodesCoordinator = errors.New("nil nodesCoordinator")

// ErrNilMarshalizer signals that the marshalizer is nil
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNoPubKeys signals an error when public keys are missing
var ErrNoPubKeys = errors.New("no public keys defined")

// ErrPublicKeyNotFoundInGenesis signals an error when the public key is not in genesis file
var ErrPublicKeyNotFoundInGenesis = errors.New("public key is not valid, it is missing from genesis file")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to set nil pubkey converter")

// ErrCouldNotParsePubKey signals that a given public key could not be parsed
var ErrCouldNotParsePubKey = errors.New("could not parse node's public key")

// ErrCouldNotParseAddress signals that a given address could not be parsed
var ErrCouldNotParseAddress = errors.New("could not parse node's address")

// ErrNegativeOrZeroConsensusGroupSize signals that an invalid consensus group size has been provided
var ErrNegativeOrZeroConsensusGroupSize = errors.New("negative or zero consensus group size")

// ErrMinNodesPerShardSmallerThanConsensusSize signals that an invalid min nodes per shard has been provided
var ErrMinNodesPerShardSmallerThanConsensusSize = errors.New("minimum nodes per shard is smaller than consensus group size")

// ErrNodesSizeSmallerThanMinNoOfNodes signals that there are not enough nodes defined in genesis file
var ErrNodesSizeSmallerThanMinNoOfNodes = errors.New("length of nodes defined is smaller than min nodes per shard required")

// ErrNilInputNodesMap signals that a nil nodes map was provided
var ErrNilInputNodesMap = errors.New("nil input nodes map")

// ErrSmallShardEligibleListSize signals that the eligible validators list's size is less than the consensus size
var ErrSmallShardEligibleListSize = errors.New("small shard eligible list size")

// ErrSmallMetachainEligibleListSize signals that the eligible validators list's size is less than the consensus size
var ErrSmallMetachainEligibleListSize = errors.New("small metachain eligible list size")

// ErrMapSizeZero signals that there are no elements in the map
var ErrMapSizeZero = errors.New("map size zero")

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

// ErrNilOwnPublicKey signals that a nil own public key has been provided
var ErrNilOwnPublicKey = errors.New("nil own public key")

// ErrNilEndOfProcessingHandler signals that a nil end of processing handler has been provided
var ErrNilEndOfProcessingHandler = errors.New("nil end of processing handler")

// ErrNilOrEmptyDestinationForDistribute signals that a nil or empty value was provided for destination of distributedNodes
var ErrNilOrEmptyDestinationForDistribute = errors.New("nil or empty destination list for distributeNodes")
