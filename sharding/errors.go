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

// ErrNilNodesCoordinator signals that the nodesCoordinator is nil
var ErrNilNodesCoordinator = errors.New("nil nodesCoordinator")

// ErrNilRater signals that the rater is nil
var ErrNilRater = errors.New("nil rater")

// ErrNoPubKeys signals an error when public keys are missing
var ErrNoPubKeys = errors.New("no public keys defined")

// ErrPublicKeyNotFoundInGenesis signals an error when the public key is not in genesis file
var ErrPublicKeyNotFoundInGenesis = errors.New("public key is not valid, it is missing from genesis file")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilAddressConverter signals that a nil address converter has been provided
var ErrNilAddressConverter = errors.New("trying to set nil address converter")

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

// ErrInvalidConsensusGroupSize signals that the consensus size is invalid (e.g. value is negative)
var ErrInvalidConsensusGroupSize = errors.New("invalid consensus group size")

// ErrEligibleSelectionMismatch signals a mismatch between the eligible list and the group selection bitmap
var ErrEligibleSelectionMismatch = errors.New("invalid eligible validator selection")

// ErrEligibleTooManySelections signals an invalid selection for consensus group
var ErrEligibleTooManySelections = errors.New("too many selections for consensus group")

// ErrEligibleTooFewSelections signals an invalid selection for consensus group
var ErrEligibleTooFewSelections = errors.New("too few selections for consensus group")

// ErrNilRandomness signals that a nil randomness source has been provided
var ErrNilRandomness = errors.New("nil randomness source")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")

// ErrNilStake signals that a nil stake structure has been provided
var ErrNilStake = errors.New("nil stake")

// ErrNegativeStake signals that the stake is negative
var ErrNegativeStake = errors.New("negative stake")

// ErrNilAddress signals that the address is nil
var ErrNilAddress = errors.New("nil address")

// ErrValidatorNotFound signals that the validator has not been found
var ErrValidatorNotFound = errors.New("validator not found")

// ErrNotImplemented signals a call of a non implemented functionality
var ErrNotImplemented = errors.New("feature not implemented")
