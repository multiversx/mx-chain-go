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

// ErrNegativeOrZeroConsensusGroupSize signals that an invalid consensus group size has been provided
var ErrNegativeOrZeroConsensusGroupSize = errors.New("negative or zero consensus group size")

// ErrMinNodesPerShardSmallerThanConsensusSize signals that an invalid min nodes per shard has been provided
var ErrMinNodesPerShardSmallerThanConsensusSize = errors.New("minimum nodes per shard is smaller than consensus group size")

// ErrNodesSizeSmallerThanMinNoOfNodes signals that there are not enough nodes defined in genesis file
var ErrNodesSizeSmallerThanMinNoOfNodes = errors.New("length of nodes defined is smaller than min nodes per shard required")
