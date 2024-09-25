package proofscache

import "errors"

// ErrMissingProof signals that the proof is missing
var ErrMissingProof = errors.New("missing proof")

// ErrNilProof signals that a nil proof has been provided
var ErrNilProof = errors.New("nil proof provided")
