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
