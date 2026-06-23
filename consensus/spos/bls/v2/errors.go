package v2

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrTimeOut signals that the time is out
var ErrTimeOut = errors.New("time is out")

// ErrProofAlreadyPropagated signals that the proof was already propagated
var ErrProofAlreadyPropagated = errors.New("proof already propagated")

// ErrValidSignatureFromInvalidSigner signals that a valid signature was received on invalid signers message
var ErrValidSignatureFromInvalidSigner = errors.New("valid signature from invalid sender")

// ErrHeaderHashMismatch signals that header hash does not match
var ErrHeaderHashMismatch = errors.New("header hash does not match")
