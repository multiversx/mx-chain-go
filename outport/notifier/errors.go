package notifier

import (
	"errors"
)

// ErrNilTransactionsPool signals that a nil transactions pool was provided
var ErrNilTransactionsPool = errors.New("nil transactions pool")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilHTTPClientWrapper signals that a nil http client wrapper has been provided
var ErrNilHTTPClientWrapper = errors.New("nil http client wrapper")

// ErrNilMarshaller signals that a nil marshaller has been provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilPubKeyConverter signals that a nil pubkey converter has been provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilHasher is raised when a valid hasher is expected but nil used
var ErrNilHasher = errors.New("hasher is nil")
