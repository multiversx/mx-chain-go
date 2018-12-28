package consensus

import (
	"github.com/pkg/errors"
)

// ErrNegativeStake signals that the stake is negative
var ErrNegativeStake = errors.New("negative stake")

// ErrNilPubKey signals that the public key is nil
var ErrNilPubKey = errors.New("nil public key")

// ErrNilInputSlice signals that a nil slice has been provided
var ErrNilInputSlice = errors.New("nil input slice")

// ErrSmallEligibleListSize signals that the eligible validators list's size is less than the consensus size
var ErrSmallEligibleListSize = errors.New("small eligible list size")

// ErrInvalidConsensusSize signals that the consensus size is invalid (e.g. value is negative)
var ErrInvalidConsensusSize = errors.New("invalid consensus size")

// ErrNilRandomness signals that a nil randomness source has been provided
var ErrNilRandomness = errors.New("nil randomness source")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")
