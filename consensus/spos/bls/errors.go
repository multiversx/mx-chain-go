package bls

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// todo: make only 1 error

var errExtraSigShareDataNotFound = errors.New("extra sig share data not found in consensus message")
