package blockchain

import (
	"errors"
)

// ErrBadBlocksCacheNil defines the error for using a nil cache for bad blocks
var ErrBadBlocksCacheNil = errors.New("badBlocksCache nil")

// ErrTxUnitNil defines the error for using a nil storage unit for transactions
var ErrTxUnitNil = errors.New("txUnit nil")

// ErrPeerBlockUnitNil defines the error for using a nil storage unit for peer blocks
var ErrPeerBlockUnitNil = errors.New("peerBlockUnit nil")

// ErrHeaderUnitNil defines the error for using a nil storage unit for block headers
var ErrHeaderUnitNil = errors.New("header unit nil")

// ErrMetachainHeaderUnitNil defines the error for using a nil storage unit for metachain block headers
var ErrMetachainHeaderUnitNil = errors.New("metachain header unit nil")

// ErrMetaBlockUnitNil defines the error for using a nil storage unit for metachain blocks
var ErrMetaBlockUnitNil = errors.New("metablock storage unit is nil")

// ErrShardDataUnitNil defines the error for using a nil storage unit for metachain shard data
var ErrShardDataUnitNil = errors.New("metachain shard data storage unit is nil")

// ErrPeerDataUnitNil defines the error for using a nil storage unit for metachain peer data
var ErrPeerDataUnitNil = errors.New("metachain peer data storage unit is nil")

// ErrNoSuchStorageUnit defines the error for using an invalid storage unit
var ErrNoSuchStorageUnit = errors.New("no such unit type")

// ErrMiniBlockUnitNil defines the error for using a nil storage unit for mini blocks
var ErrMiniBlockUnitNil = errors.New("nil mini block unit")

// ErrWrongTypeInSet defines the error for trying to set the wrong type
var ErrWrongTypeInSet = errors.New("wrong type in setter")
