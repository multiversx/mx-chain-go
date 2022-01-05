package alteredaccounts

import "errors"

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilPubKeyConverter signals that a nil public key converter has been provided
var ErrNilPubKeyConverter = errors.New("nil public key converter")

// ErrNilAccountsDB signals that a nil accounts DB has been provided
var ErrNilAccountsDB = errors.New("nil accounts DB")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")
