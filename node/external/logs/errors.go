package logs

import (
	"errors"
)

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
// TODO: Move to elrond-go-core
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")
