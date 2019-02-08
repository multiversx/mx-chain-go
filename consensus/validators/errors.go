package validators

import (
	"github.com/pkg/errors"
)

// ErrNegativeStake signals that the stake is negative
var ErrNegativeStake = errors.New("negative stake")

// ErrNilPubKey signals that the public key is nil
var ErrNilPubKey = errors.New("nil public key")
