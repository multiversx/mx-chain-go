package blockchain

import (
	"errors"
)

// ErrBadBlocksCacheNil defines the error for using a nil cache for bad blocks
var ErrBadBlocksCacheNil = errors.New("badBlocksCache nil")

// ErrTxUnitNil defines the error for using a nil storage unit for transactions
var ErrTxUnitNil = errors.New("txUnit nil")

// ErrTxBlockUnitNil defines the error for using a nil storage unit for transaction blocks
var ErrTxBlockUnitNil = errors.New("txBlockUnit nil")

// ErrStateBlockUnitNil defines the error for using a nil storage unit for state blocks
var ErrStateBlockUnitNil = errors.New("stateBlockUnit nil")

// ErrPeerBlockUnitNil defines the error for using a nil storage unit for peer blocks
var ErrPeerBlockUnitNil = errors.New("peerBlockUnit nil")

// ErrHeaderUnitNil defines the error for using a nil storage unit for block headers
var ErrHeaderUnitNil = errors.New("header nil")

// ErrMetaBlockUnitNil defines the error for using a nil storage unit for metachain blocks
var ErrMetaBlockUnitNil = errors.New("metablock storage unit is nil")

// ErrNoSuchStorageUnit defines the error for using an invalid storage unit
var ErrNoSuchStorageUnit = errors.New("no such unit type")
