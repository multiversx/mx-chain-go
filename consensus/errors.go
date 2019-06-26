package consensus

import (
	"errors"
)

// ErrNilStake signals that a nil stake structure has been provided
var ErrNilStake = errors.New("nil stake")

// ErrNegativeStake signals that the stake is negative
var ErrNegativeStake = errors.New("negative stake")

// ErrNilPubKey signals that the public key is nil
var ErrNilPubKey = errors.New("nil public key")
