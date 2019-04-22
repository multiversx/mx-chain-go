package core

import (
	"errors"
)

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrInvalidNonce signals that an invalid block nonce was provided
var ErrInvalidNonce = errors.New("invalid block nonce")
