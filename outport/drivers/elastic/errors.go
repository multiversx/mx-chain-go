package elastic

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

// ErrNilAccountsDB signals that a nil accounts database has been provided
var ErrNilAccountsDB = errors.New("nil accounts db")

// ErrEmptyEnabledIndexes signals that an empty slice of enables indexes has been provided
var ErrEmptyEnabledIndexes = errors.New("empty enabled indexes slice")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrReadTemplatesFile signals that a read error occurred while reading template file
var ErrReadTemplatesFile = errors.New("error while reading template file")

// ErrReadPolicyFile signals that a read error occurred while reading policy file
var ErrReadPolicyFile = errors.New("error while reading policy file")

// ErrWriteToBuffer signals that a write error occurred
var ErrWriteToBuffer = errors.New("error while writing to buffer")
