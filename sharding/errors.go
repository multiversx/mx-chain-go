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
