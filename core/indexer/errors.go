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

// ErrNilDatabaseClient signals that an operation has been attempted to or with a nil database client implementation
var ErrNilDatabaseClient = errors.New("nil database client")

// ErrNilOptions signals that structure that contains indexer options is nil
var ErrNilOptions = errors.New("nil options")

// ErrNegativeCacheSize signals that a invalid cache size has been provided
var ErrNegativeCacheSize = errors.New("negative cache size")

// ErrNilAccountsDB signals that a nil accounts db has been provided
var ErrNilAccountsDB = errors.New("nil accounts db")

// ErrEmptyEnabledIndexes signals that an empty slice of enables indexes has been provided
var ErrEmptyEnabledIndexes = errors.New("empty enabled indexes slice")
