package data

import (
	"errors"
)

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrInvalidHeaderType signals an invalid header pointer was provided
var ErrInvalidHeaderType = errors.New("invalid header type")

// ErrInvalidBodyType signals an invalid header pointer was provided
var ErrInvalidBodyType = errors.New("invalid body type")

// ErrNilBlockBody signals that block body is nil
var ErrNilBlockBody = errors.New("nil block body")

// ErrMiniBlockEmpty signals that mini block is empty
var ErrMiniBlockEmpty = errors.New("mini block is empty")

// ErrWrongTypeAssertion signals that wrong type was provided
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilElrondAddress signals that nil elrond address was provided
var ErrNilElrondAddress = errors.New("nil elrond address")

// ErrNilBurnAddress signals that nil burn address was provided
var ErrNilBurnAddress = errors.New("nil burn address")

// ErrNilAddressConverter signals that nil address converter was provided
var ErrNilAddressConverter = errors.New("nil address converter")

// ErrNilShardCoordinator signals that nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilNodesCoordinator signals that nil shard coordinator was provided
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNilMarshalizer is raised when the NewTrie() function is called, but a marshalizer isn't provided
var ErrNilMarshalizer = errors.New("no marshalizer provided")

// ErrNilDatabase is raised when a database operation is called, but no database is provided
var ErrNilDatabase = errors.New("no database provided")

// ErrInvalidCacheSize is raised when the given size for the cache is invalid
var ErrInvalidCacheSize = errors.New("cache size is invalid")

// ErrInvalidValue signals that an invalid value has been provided such as NaN to an integer field
var ErrInvalidValue = errors.New("invalid value")
