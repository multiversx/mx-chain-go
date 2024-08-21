package bls

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrorCreateAndSendSignMessage defines an error for sendSignature function
var ErrorCreateAndSendSignMessage = errors.New("false sent by createAndSendSignMessage")

// ErrorCompleteSigSubround defines an error for sendSignature function
var ErrorCompleteSigSubround = errors.New("false sent by completeSignatureSubRound")
