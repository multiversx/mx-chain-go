package interceptedBlocks

import "errors"

var (
	// ErrInvalidProof signals that an invalid proof has been provided
	ErrInvalidProof = errors.New("invalid proof")
)
