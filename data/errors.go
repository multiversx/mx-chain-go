package data

import (
	"errors"
)

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

// ErrNilPeerChangeBlockDataPool signals that a nil peer change pool has been provided
var ErrNilPeerChangeBlockDataPool = errors.New("nil peer change block data pool")

// ErrNilStateBlockDataPool signals that a nil state pool has been provided
var ErrNilStateBlockDataPool = errors.New("nil state data pool")

// ErrNilTxBlockDataPool signals that a nil tx block body pool has been provided
var ErrNilTxBlockDataPool = errors.New("nil tx block data pool")
