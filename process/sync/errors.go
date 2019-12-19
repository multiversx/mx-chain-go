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

// ErrRollBackBehindFinalHeader signals that a roll back behind final header has been attempted
var ErrRollBackBehindFinalHeader = errors.New("roll back behind final header is not permitted")

// ErrRollBackBehindForkNonce signals that a roll back behind fork nonce is not permitted
var ErrRollBackBehindForkNonce = errors.New("roll back behind fork nonce is not permitted")

// ErrRandomSeedNotValid signals that the random seed is not valid
var ErrRandomSeedNotValid = errors.New("random seed is not valid")

// ErrInvalidShardId signals that an invalid shard id has been provided
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrNilNotarizedHeader signals that an nil notarized header has been provided
var ErrNilNotarizedHeader = errors.New("nil notarized header")

// ErrGenesisTimeMissmatch signals that a received header has a genesis time missmatch
var ErrGenesisTimeMissmatch = errors.New("genesis time missmatch")
