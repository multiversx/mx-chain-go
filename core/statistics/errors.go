package statistics

import (
	"errors"
)

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrInvalidRoundDuration signals that an invalid round duration was provided
var ErrInvalidRoundDuration = errors.New("invalid round duration")

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrNilInitialTPSBenchmarks signals that nil TPS benchmarks have been provided
var ErrNilInitialTPSBenchmarks = errors.New("nil initial TPS benchmarks")

// ErrNilConfig signals that a nil config was provided
var ErrNilConfig = errors.New("nil config")

// ErrNilPathHandler signals that a nil path handler was provided
var ErrNilPathHandler = errors.New("nil path handler")
