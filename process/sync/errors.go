package sync

import (
	"errors"
	"fmt"
)

// ErrNilHeader signals that a nil header has been provided
var ErrNilHeader = errors.New("nil header")

// ErrNilHash signals that a nil hash has been provided
var ErrNilHash = errors.New("nil hash")

// ErrNilCurrentHeader signals that the current header is nil
var ErrNilCurrentHeader = errors.New("The current header is nil\n")

// ErrTxBlockBodyMismatch signals that a tx block body can not by asserted to TxBlockBody type
var ErrTxBlockBodyMismatch = errors.New("tx block body mismatch")

type ErrNotEmptyHeader struct {
	CurrentNonce uint64
	PoolNonce    uint64
}

func (err ErrNotEmptyHeader) Error() string {
	return fmt.Sprintf("the current header with nonce %d is not from an empty block, "+
		"try to remove header with nonce %d from pool and request it again\n",
		err.CurrentNonce, err.PoolNonce)
}
