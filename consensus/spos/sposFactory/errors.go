package sposFactory

import (
	"errors"
)

// ErrInvalidConsensusType signals that an invalid consensus type has been provided
var ErrInvalidConsensusType = errors.New("invalid consensus type")

// ErrInvalidShardId signals that an invalid shard id has been provided
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrNilConsensusComponentFactory signals that provided consensus component factory is nil
var ErrNilConsensusComponentFactory = errors.New("nil consensus component factory")
