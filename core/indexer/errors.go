package indexer

import (
	"errors"
)

// ErrBackOff signals that an error was received from the server
var ErrBackOff = errors.New("back off something is not working well")

// ErrNoElasticUrlProvided -
var ErrNoElasticUrlProvided = errors.New("no elastic url provided")

// ErrCouldNotCreatePolicy -
var ErrCouldNotCreatePolicy = errors.New("could not create policy")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilDataDispatcher signals that an operation has been attempted to or with a nil data dispatcher implementation
var ErrNilDataDispatcher = errors.New("nil data dispatcher")

// ErrNilElasticProcessor signals that an operation has been attempted to or with a nil elastic processor implementation
var ErrNilElasticProcessor = errors.New("nil elastic processor")

// ErrInvalidCacheSize signals that and invalid indexer cache size was provided
var ErrInvalidCacheSize = errors.New("invalid indexer cache size")
