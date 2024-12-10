package v2

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrWrongSizeBitmap defines the error for wrong size bitmap
var ErrWrongSizeBitmap = errors.New("wrong size bitmap")

// ErrNotEnoughSignatures defines the error for not enough signatures
var ErrNotEnoughSignatures = errors.New("not enough signatures")
