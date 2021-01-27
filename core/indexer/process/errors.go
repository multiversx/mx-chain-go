package process

import (
	"errors"
)

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilAccountsDB signals that a nil accounts database has been provided
var ErrNilAccountsDB = errors.New("nil accounts db")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrReadTemplatesFile signals that a read error occurred while reading template file
var ErrReadTemplatesFile = errors.New("error while reading template file")

// ErrReadPolicyFile signals that a read error occurred while reading policy file
var ErrReadPolicyFile = errors.New("error while reading policy file")

// ErrWriteToBuffer signals that a write error occurred
var ErrWriteToBuffer = errors.New("error while writing to buffer")

// ErrHeaderTypeAssertion signals that body type assertion failed
var ErrHeaderTypeAssertion = errors.New("elasticsearch - header type assertion failed")
