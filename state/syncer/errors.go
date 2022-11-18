package syncer

import "errors"

// ErrNilPubkeyConverter signals that a nil public key converter was provided
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilEnableEpochsHandler signals that a nil enable epochs handler was provided
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")
