package queue

import "errors"

// ErrHeaderNonceMismatch signals a nonce missmatch
var ErrHeaderNonceMismatch = errors.New("header nonce mismatch")
