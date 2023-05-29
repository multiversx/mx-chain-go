package syncer

import "errors"

// ErrNilPubkeyConverter signals that a nil public key converter was provided
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilStorageMarker signals that a nil storage marker was provided
var ErrNilStorageMarker = errors.New("nil storage marker")
