package v2

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrTimeOut signals that the time is out
var ErrTimeOut = errors.New("time is out")

// ErrProofAlreadyPropagated signals that the proof was already propagated
var ErrProofAlreadyPropagated = errors.New("proof already propagated")

// ErrNilRoundSyncController signals that a nil round sync controller has been provided
var ErrNilRoundSyncController = errors.New("nil round sync controller")

// ErrTooManyInvalidSigners signals that too many invalid signers were received
var ErrTooManyInvalidSigners = errors.New("too many invalid signers")

// ErrValidSignatureFromInvalidSigner signals that a valid signature was received on invalid signers message
var ErrValidSignatureFromInvalidSigner = errors.New("valid signature from invalid sender")
