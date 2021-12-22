package syncer

import "errors"

// ErrNilPubkeyConverter signals that a nil public key converter was provided
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")
