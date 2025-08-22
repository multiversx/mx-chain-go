package queue

import "errors"

// ErrHeaderNonceMismatch signals a nonce mismatch
var ErrHeaderNonceMismatch = errors.New("header nonce mismatch")
