package v2

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrTimeOut signals that the time is out
var ErrTimeOut = errors.New("time is out")

// ErrProofAlreadyPropagated signals that the proof was already propagated
var ErrProofAlreadyPropagated = errors.New("proof already propagated")
