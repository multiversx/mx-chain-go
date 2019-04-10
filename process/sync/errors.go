package sync

import (
	"errors"
	"fmt"
)

// ErrNilHeader signals that a nil header has been provided
var ErrNilHeader = errors.New("nil header")

// ErrNilHash signals that a nil hash has been provided
var ErrNilHash = errors.New("nil hash")

// ErrLowerNonceInBlock signals the nonce in block is lower than the last check point nonce
var ErrLowerNonceInBlock = errors.New("lower nonce in block")

// ErrNilOrEmptyInfoStored signals that info stored si nil or empty
var ErrNilOrEmptyInfoStored = errors.New("info stored is nil or empty")

// ErrNotEmptyHeader signals the header is not an empty block
type ErrNotEmptyHeader struct {
	CurrentNonce uint64
}

func (err ErrNotEmptyHeader) Error() string {
	return fmt.Sprintf("the current header with nonce %d is not from an empty block\n",
		err.CurrentNonce)
}
