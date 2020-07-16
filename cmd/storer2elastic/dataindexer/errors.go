package dataindexer

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

// ErrNilGenesisNodesSetup signals that a nil genesis nodes setup handler has been provided
var ErrNilGenesisNodesSetup = errors.New("nil genesis nodes setup")

// ErrWrongTypeAssertion signals that an interface is not of a desired type
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrTimeIsOut signals that time is out when indexing data to elastic
var ErrTimeIsOut = errors.New("time is out when indexing")
