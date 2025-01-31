package v1

import "errors"

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrEquivalentMessagesFlagEnabledWithConsensusV1 defines the error for running with the equivalent messages flag enabled under v1 consensus
var ErrEquivalentMessagesFlagEnabledWithConsensusV1 = errors.New("equivalent messages flag enabled with consensus v1")
