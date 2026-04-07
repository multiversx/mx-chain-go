package interceptedBlocks

import "errors"

var (
	// ErrInvalidProof signals that an invalid proof has been provided
	ErrInvalidProof = errors.New("invalid proof")

	// ErrProofAlreadyExistsForNonce signals that a proof already exists for the provided nonce
	ErrProofAlreadyExistsForNonce = errors.New("proof already exists for nonce")
)
