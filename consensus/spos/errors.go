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

// ErrNilPrivateKey is raised when a valid private key was expected but nil was used
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrNilConsensusData is raised when valid consensus data was expected but nil was received
var ErrNilConsensusData = errors.New("consensus data is nil")

// ErrNilSignature is raised when a valid signature was expected but nil was used
var ErrNilSignature = errors.New("signature is nil")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil was used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilBlockHeader is raised when a valid block header is expected but nil was used
var ErrNilBlockHeader = errors.New("block header is nil")

// ErrNilTxBlockBody is raised when a valid tx block body is expected but nil was used
var ErrNilTxBlockBody = errors.New("tx block body is nil")

// ErrNilOnBroadcastHeader is raised when a valid header broadcast function pointer is expected but nil used
var ErrNilOnBroadcastHeader = errors.New("header broadcast function pointer is nil")

// ErrNilOnBroadcastTxBlockBody is raised when a valid block broadcast function pointer is expected but nil used
var ErrNilOnBroadcastTxBlockBody = errors.New("tx block body broadcast function pointer is nil")

// ErrNilMultiSigner is raised when a valid multiSigner is expected but nil used
var ErrNilMultiSigner = errors.New("multiSigner is nil")

// ErrNilConsensus is raised when a valid consensus is expected but nil used
var ErrNilConsensus = errors.New("consensus is nil")

// ErrNilBlockChain is raised when a valid blockchain is expected but nil used
var ErrNilBlockChain = errors.New("blockchain is nil")

// ErrNilHasher is raised when a valid hasher is expected but nil used
var ErrNilHasher = errors.New("hasher is nil")

// ErrNilMarshalizer is raised when a valid marshalizer is expected but nil used
var ErrNilMarshalizer = errors.New("marshalizer is nil")

// ErrNilBlockProcessor is raised when a valid block processor is expected but nil used
var ErrNilBlockProcessor = errors.New("block processor is nil")

// ErrInvalidKey is raised when an invalid key is used with a map
var ErrInvalidKey = errors.New("map key is invalid")

// ErrNilRoundState is raised when a valid round state is expected but nil used
var ErrNilRoundState = errors.New("round state is nil")
