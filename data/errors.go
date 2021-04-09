package data

import (
	"errors"
)

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrInvalidHeaderType signals an invalid header pointer was provided
var ErrInvalidHeaderType = errors.New("invalid header type")

// ErrNilBlockBody signals that block body is nil
var ErrNilBlockBody = errors.New("nil block body")

// ErrMiniBlockEmpty signals that mini block is empty
var ErrMiniBlockEmpty = errors.New("mini block is empty")

// ErrNilShardCoordinator signals that nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilMarshalizer is raised when the NewTrie() function is called, but a marshalizer isn't provided
var ErrNilMarshalizer = errors.New("no marshalizer provided")

// ErrNilDatabase is raised when a database operation is called, but no database is provided
var ErrNilDatabase = errors.New("no database provided")

// ErrInvalidCacheSize is raised when the given size for the cache is invalid
var ErrInvalidCacheSize = errors.New("cache size is invalid")

// ErrInvalidValue signals that an invalid value has been provided such as NaN to an integer field
var ErrInvalidValue = errors.New("invalid value")

// ErrNilThrottler signals that nil throttler has been provided
var ErrNilThrottler = errors.New("nil throttler")

// ErrTimeIsOut signals that time is out
var ErrTimeIsOut = errors.New("time is out")

// ErrLeafSizeTooBig signals that the value size of the leaf is too big
var ErrLeafSizeTooBig = errors.New("leaf size too big")

// ErrNilValue signals the value is nil
var ErrNilValue = errors.New("nil value")

// ErrNilSignature signals that a operation has been attempted with a nil signature
var ErrNilSignature = errors.New("nil signature")

// ErrNegativeValue signals that a negative value has been detected and it is not allowed
var ErrNegativeValue = errors.New("negative value")

// ErrInvalidUserNameLength signals that provided user name length is invalid
var ErrInvalidUserNameLength = errors.New("invalid user name length")

// ErrNilTxHash signals that an operation has been attempted with a nil hash
var ErrNilTxHash = errors.New("nil transaction hash")

// ErrNilRcvAddr signals that an operation has been attempted to or with a nil receiver address
var ErrNilRcvAddr = errors.New("nil receiver address")

// ErrNilSndAddr signals that an operation has been attempted to or with a nil sender address
var ErrNilSndAddr = errors.New("nil sender address")

// ErrNilPointerReceiver signals that a nil pointer receiver was used
var ErrNilPointerReceiver = errors.New("nil pointer receiver")

// ErrNilPointerDereference signals that a nil pointer dereference was detected and avoided
var ErrNilPointerDereference = errors.New("nil pointer dereference")

// ErrInvalidTypeAssertion signals an invalid type assertion
var ErrInvalidTypeAssertion = errors.New("invalid type assertion")
