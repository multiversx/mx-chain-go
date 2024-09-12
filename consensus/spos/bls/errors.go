package bls

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrAux defines the error that does not have any meaning but helps for debug
var ErrAux = errors.New("auxiliary error")
