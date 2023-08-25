package blockchain

import (
	"errors"
)

// ErrHeaderUnitNil defines the error for using a nil storage unit for block headers
var ErrHeaderUnitNil = errors.New("header unit nil")

// ErrWrongTypeInSet defines the error for trying to set the wrong type
var ErrWrongTypeInSet = errors.New("wrong type in setter")

// ErrNilAppStatusHandler defines the error for setting a nil AppStatusHandler
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrOperationNotPermitted signals that operation is not permitted
var ErrOperationNotPermitted = errors.New("operation not permitted")

// ErrNilBlockChain signals that a nil blockchain has been provided
var ErrNilBlockChain = errors.New("nil block chain")
