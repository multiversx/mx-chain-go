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

// ErrSignedBlock signals that a block is signed
type ErrSignedBlock struct {
	CurrentNonce uint64
}

func (err ErrSignedBlock) Error() string {
	return fmt.Sprintf("the current header with nonce %d is from a signed block\n",
		err.CurrentNonce)
}
