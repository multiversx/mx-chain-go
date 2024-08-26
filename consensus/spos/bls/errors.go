package bls

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

var JobIsNotDoneError error = errors.New("job is not done")

var SignatureShareError = errors.New("signature share error")

var SetJobDoneError = errors.New("set job done error")

var JobDoneError = errors.New("job done error")
