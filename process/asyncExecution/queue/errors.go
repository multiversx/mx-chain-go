package queue

import "errors"

var (
	// ErrHeaderNonceMismatch signals a nonce mismatch
	ErrHeaderNonceMismatch = errors.New("header nonce mismatch")
	// ErrInvalidHeaderNonce signals the provided nonce is invalid
	ErrInvalidHeaderNonce = errors.New("invalid header nonce")
	// ErrQueueIntegrityViolation signals a queue integrity violation
	ErrQueueIntegrityViolation = errors.New("queue integrity violation")
)
