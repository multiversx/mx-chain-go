package indexer

import (
	"errors"
)

// ErrNilDataDispatcher signals that an operation has been attempted to or with a nil data dispatcher implementation
var ErrNilDataDispatcher = errors.New("nil data dispatcher")

// ErrNilElasticProcessor signals that an operation has been attempted to or with a nil elastic processor implementation
var ErrNilElasticProcessor = errors.New("nil elastic processor")

// ErrNegativeCacheSize signals that an invalid cache size has been provided
var ErrNegativeCacheSize = errors.New("negative cache size")

// ErrEmptyEnabledIndexes signals that an empty slice of enables indexes has been provided
var ErrEmptyEnabledIndexes = errors.New("empty enabled indexes slice")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")
