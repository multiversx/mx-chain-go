package bls

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrExtraSigShareDataNotFound signals that a nil entry in consensus message is found for extra sig share data
var ErrExtraSigShareDataNotFound = errors.New("extra sig share data not found in consensus message")
