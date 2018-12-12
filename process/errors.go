package process

import (
	"github.com/pkg/errors"
)

// ErrNilMessenger signals that a nil Messenger object was provided
var ErrNilMessenger = errors.New("nil Messenger")

// ErrNilNewer signals that a nil Newer object was provided
var ErrNilNewer = errors.New("nil Newer")

// ErrRegisteringValidator signals that a registration validator occur
var ErrRegisteringValidator = errors.New("error while registering validator")

// ErrNilAddressConverter signals that a nil AddressConverter has been provided
var ErrNilAddressConverter = errors.New("nil AddressConverter")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilTemplateObj signals that an operation has been attempted to or with a nil template object
var ErrNilTemplateObj = errors.New("nil TemplateObj")

// ErrNilTransientPool signals that an operation has been attempted to or with a nil transient pool of data
var ErrNilTransientPool = errors.New("nil transient pool")

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

// ErrNilPeerChangeBlockDataPool signals that a nil peer change pool has been provided
var ErrNilPeerChangeBlockDataPool = errors.New("nil peer change block data pool")

// ErrNilStateBlockDataPool signals that a nil state pool has been provided
var ErrNilStateBlockDataPool = errors.New("nil state data pool")

// ErrNilTxBlockDataPool signals that a nil tx block body pool has been provided
var ErrNilTxBlockDataPool = errors.New("nil tx block data pool")
