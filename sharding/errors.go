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
var ErrNoPubKeys = errors.New("shard id out of range")

// ErrNotEnoughValidators signals an error when there are not enough validators for consensus
var ErrNotEnoughValidators = errors.New("not enough validators to make consensus group")

// ErrNoValidPublicKey signals an error when the public key is not in genesis file
var ErrNoValidPublicKey = errors.New("Public key is not valid, it is missing from genesis file")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilAddressConverter signals that a nil address converter has been provided
var ErrNilAddressConverter = errors.New("trying to set nil address converter")
