package sync

import (
	"errors"
	"fmt"
)

// ErrNilHeader signals that a nil header has been provided
var ErrNilHeader = errors.New("nil header")

// ErrNilHash signals that a nil hash has been provided
var ErrNilHash = errors.New("nil hash")

// ErrLowerNonceInBlock signals that the nonce in block is lower than the last check point nonce
var ErrLowerNonceInBlock = errors.New("lower nonce in block")

// ErrHigherNonceInBlock signals that the nonce in block is higher than what could exist in the current round
var ErrHigherNonceInBlock = errors.New("higher nonce in block")

// ErrLowerRoundInBlock signals that the round index in block is lower than the checkpoint round
var ErrLowerRoundInBlock = errors.New("lower round in block")

// ErrHigherRoundInBlock signals that the round index in block is higher than the current round of chronology
var ErrHigherRoundInBlock = errors.New("higher round in block")

// ErrBlockIsNotSigned signals that the block is not signed
var ErrBlockIsNotSigned = errors.New("block is not signed")

// ErrSignedBlock signals that a block is signed
type ErrSignedBlock struct {
	CurrentNonce uint64
}

func (err ErrSignedBlock) Error() string {
	return fmt.Sprintf("the current header with nonce %d is from a signed block\n",
		err.CurrentNonce)
}
