package statistics

import (
	"errors"
)

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrInvalidRoundDuration signals that an invalid round duration was provided
var ErrInvalidRoundDuration = errors.New("invalid round duration")

// ErrNilFileToWriteStats signals that the file where statistics should be written is nil
var ErrNilFileToWriteStats = errors.New("nil file to write statistics")

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrNilInitialTPSBenchmarks signals that nil TPS benchmarks have been provided
var ErrNilInitialTPSBenchmarks = errors.New("nil initial TPS benchmarks")
