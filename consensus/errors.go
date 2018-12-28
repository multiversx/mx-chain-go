package consensus

import (
	"github.com/pkg/errors"
)

var ErrNegativeStake = errors.New("negative stake")

var ErrNilPubKey = errors.New("nil public key")

var ErrNilInputSlice = errors.New("nil input slice")

var ErrSmallEligibleListSize = errors.New("small eligible list size")

var ErrInvalidConsensusSize = errors.New("invalid consensus size")

var ErrNilRandomness = errors.New("nil randomness source")

var ErrNilHasher = errors.New("nil hasher")
