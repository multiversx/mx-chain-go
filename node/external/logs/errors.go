package logs

import (
	"errors"
)

// ErrNilStorer signals that a nil storer has been provided
// TODO: Move to elrond-go-core
var ErrNilStorer = errors.New("nil storer")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
// TODO: Move to elrond-go-core
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")
