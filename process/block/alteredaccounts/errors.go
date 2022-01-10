package alteredaccounts

import "errors"

// errNilShardCoordinator signals that a nil shard coordinator has been provided
var errNilShardCoordinator = errors.New("nil shard coordinator")

// errNilPubKeyConverter signals that a nil public key converter has been provided
var errNilPubKeyConverter = errors.New("nil public key converter")

// errNilAccountsDB signals that a nil accounts DB has been provided
var errNilAccountsDB = errors.New("nil accounts DB")

// errNilMarshalizer signals that a nil marshalizer has been provided
var errNilMarshalizer = errors.New("nil marshalizer")
