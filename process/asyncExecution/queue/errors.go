package queue

import "errors"

var (
	// ErrHeaderNonceMismatch signals a nonce mismatch
	ErrHeaderNonceMismatch = errors.New("header nonce mismatch")
	// ErrMissingHeaderNonce signals the provided nonce is missing
	ErrMissingHeaderNonce = errors.New("missing header nonce")
)
