package bls

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

var ErrorCreateAndSendSignMessage = errors.New("false sent by createAndSendSignMessage")

var ErrorCompleteSigSubround = errors.New("false sent by completeSignatureSubRound")
