package blockchain

import (
	"errors"
)

// ErrBadBlocksCacheNil defines the error for using a nil cache for bad blocks
var ErrBadBlocksCacheNil = errors.New("badBlocksCache nil")

// ErrHeaderUnitNil defines the error for using a nil storage unit for block headers
var ErrHeaderUnitNil = errors.New("header unit nil")

// ErrWrongTypeInSet defines the error for trying to set the wrong type
var ErrWrongTypeInSet = errors.New("wrong type in setter")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")
