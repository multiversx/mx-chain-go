package dataprocessor

import "errors"

// ErrNilElasticIndexer signals that a nil elastic indexer has been provided
var ErrNilElasticIndexer = errors.New("nil elastic indexer")

// ErrNilDatabaseReader signals that a nil databse reader has been provided
var ErrNilDatabaseReader = errors.New("nil database reader")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")

// ErrNilTPSBenchmarkUpdater signals that a nil tps benchmark updater has been provided
var ErrNilTPSBenchmarkUpdater = errors.New("nil tps benchmark updater")

// ErrNilRatingProcessor signals that a nil rating processor has been provided
var ErrNilRatingProcessor = errors.New("nil rating processor")

// ErrNilUint64ByteSliceConverter signals that a nil uint64 byte slice converter has been provided
var ErrNilUint64ByteSliceConverter = errors.New("nil uint64 byte slice converter")

// ErrNilGenesisNodesSetup signals that a nil genesis nodes setup handler has been provided
var ErrNilGenesisNodesSetup = errors.New("nil genesis nodes setup")

// ErrNoMetachainDatabase signals that no metachain database hasn't been found
var ErrNoMetachainDatabase = errors.New("no metachain database - cannot index")

// ErrDatabaseInfoNotFound signals that a database information hasn't been found
var ErrDatabaseInfoNotFound = errors.New("database info not found")

// ErrNilHeaderMarshalizer signals that a nil header marshalizer has been provided
var ErrNilHeaderMarshalizer = errors.New("nil header marshalizer")

// ErrRangeIsOver signals that the range cannot be continued as the handler returned false
var ErrRangeIsOver = errors.New("range is over")

// ErrNilDataReplayer signals that a nil data replayer has been provided
var ErrNilDataReplayer = errors.New("nil data replayer")

// ErrNilPubKeyConverter signals that a nil public key converter has been provided
var ErrNilPubKeyConverter = errors.New("nil public key converter")

// ErrOsSignalIntercepted is returned when a signal from the OS is received
var ErrOsSignalIntercepted = errors.New("os signal intercepted")

// ErrNilHandlerFunc signals that a nil handler function for raning has been provided
var ErrNilHandlerFunc = errors.New("nil handler function for ranging")
