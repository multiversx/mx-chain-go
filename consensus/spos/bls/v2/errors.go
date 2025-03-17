package v2

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

var errShouldNotSendProof = errors.New("should not send proof")
