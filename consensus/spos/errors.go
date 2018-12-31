package spos

import (
	"errors"
)

// ErrNilChronology is raised when an operation is attempted with a nil chronology
var ErrNilChronology = errors.New("chronology is null")

// ErrNilRound is raised when an operation is attempted with a nil round
var ErrNilRound = errors.New("round is null")

// ErrNegativeRoundIndex is raised when an operation is attempted with a negative round index
var ErrNegativeRoundIndex = errors.New("round index is negative")

// ErrNilConsensusGroup is raised when an operation is attempted with a nil consensus group
var ErrNilConsensusGroup = errors.New("consensusGroup is null")

// ErrEmptyConsensusGroup is raised when an operation is attempted with an empty consensus group
var ErrEmptyConsensusGroup = errors.New("consensusGroup is empty")

// ErrSelfNotFoundInConsensus is raised when self expected in consensus group but not found
var ErrSelfNotFoundInConsensus = errors.New("self not found in consensus group")

// ErrNilPublicKey is raised when a valid public key was expected but nil was used
var ErrNilPublicKey = errors.New("public key is nil")

// ErrNilPublicKey is raised when a valid public key was expected but nil was used
var ErrNilPrivateKey = errors.New("public key is nil")