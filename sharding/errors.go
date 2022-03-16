package sharding

import (
	"errors"
)

// ErrShardIdOutOfRange signals an error when shard id is out of range
var ErrShardIdOutOfRange = errors.New("shard id out of range")

// ErrNoPubKeys signals an error when public keys are missing
var ErrNoPubKeys = errors.New("no public keys defined")

// ErrPublicKeyNotFoundInGenesis signals an error when the public key is not in genesis file
var ErrPublicKeyNotFoundInGenesis = errors.New("public key is not valid, it is missing from genesis file")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to set nil pubkey converter")

// ErrInvalidMaximumNumberOfShards signals that an invalid maximum number of shards has been provided
var ErrInvalidMaximumNumberOfShards = errors.New("trying to set an invalid maximum number of shards")

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

// ErrNilOwnPublicKey signals that a nil own public key has been provided
var ErrNilOwnPublicKey = errors.New("nil own public key")

// ErrNilEndOfProcessingHandler signals that a nil end of processing handler has been provided
var ErrNilEndOfProcessingHandler = errors.New("nil end of processing handler")
