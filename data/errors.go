package data

import (
	"errors"
)

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

// ErrNilPeerChangeBlockDataPool signals that a nil peer change pool has been provided
var ErrNilPeerChangeBlockDataPool = errors.New("nil peer change block data pool")

// ErrNilTxBlockDataPool signals that a nil tx block body pool has been provided
var ErrNilTxBlockDataPool = errors.New("nil tx block data pool")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilNonceConverter signals that a nil nonce-byte slice converter has been provided
var ErrNilNonceConverter = errors.New("nil nonce converter")

// ErrNilMetaBlockPool signals that a nil meta block data pool was provided
var ErrNilMetaBlockPool = errors.New("nil meta block data pool")

// ErrNilMiniBlockHashesPool signals that a nil meta block data pool was provided
var ErrNilMiniBlockHashesPool = errors.New("nil meta block mini block hashes data pool")

// ErrNilShardHeaderPool signals that a nil meta block data pool was provided
var ErrNilShardHeaderPool = errors.New("nil meta block shard header data pool")

// ErrInvalidHeaderType signals an invalid header pointer was provided
var ErrInvalidHeaderType = errors.New("invalid header type")

// ErrInvalidBodyType signals an invalid header pointer was provided
var ErrInvalidBodyType = errors.New("invalid body type")

// ErrNilBlockBody signals that block body is nil
var ErrNilBlockBody = errors.New("nil block body")

// ErrMiniBlockEmpty signals that mini block is empty
var ErrMiniBlockEmpty = errors.New("mini block is empty")

// ErrNilCodeHash signals code hash is nil
var ErrNilCodeHash = errors.New("code hash is nil")

// ErrNilRootHash signals root hash is nil
var ErrNilRootHash = errors.New("root hash is nil")

// ErrNilBalance signals balance is nil
var ErrNilBalance = errors.New("balance is nil")

// ErrNegativeBalance signals balance is negative
var ErrNegativeBalance = errors.New("balance is negative")

// ErrNilRegistrationData signals nil registration data
var ErrNilRegistrationData = errors.New("registration data is nil")

// ErrNilBalance signals balance is nil
var ErrNilTxCount = errors.New("TxCount is nil")

// ErrNegativeBalance signals balance is negative
var ErrNegativeTxCount = errors.New("TxCount is negative")

// ErrMiniBlockDataEmpty signals mini block data empty
var ErrMiniBlockDataEmpty = errors.New("MiniBlock data is empty for shard account")

// ErrWrongTypeAssertion signals that wrong type was provided
var ErrWrongTypeAssertion = errors.New("Wrong type assertion")
