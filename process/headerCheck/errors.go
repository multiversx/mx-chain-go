package headerCheck

import "errors"

// ErrInvalidReferenceChainID signals that the provided reference chain ID is not valid
var ErrInvalidReferenceChainID = errors.New("invalid reference Chain ID provided")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID")

// ErrNilHeaderVersionHandler signals that the provided header version handler is nil
var ErrNilHeaderVersionHandler = errors.New("nil header version handler")

// ErrIndexOutOfBounds signals that the given index is outside expected bounds
var ErrIndexOutOfBounds = errors.New("index is out of bounds")

// ErrIndexNotSelected signals that the given index is not selected
var ErrIndexNotSelected = errors.New("index is not selected")

// ErrProofShardMismatch signals that the proof shard does not match the header shard
var ErrProofShardMismatch = errors.New("proof shard mismatch")

// ErrProofHeaderHashMismatch signals that the proof header hash does not match the header hash
var ErrProofHeaderHashMismatch = errors.New("proof header hash mismatch")

// ErrProofNotExpected signals that the proof is not expected
var ErrProofNotExpected = errors.New("proof not expected")
