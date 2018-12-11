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

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilHeaderDataPool signals that a nil header pool has been provided
var ErrNilHeaderDataPool = errors.New("nil header data pool")

// ErrNilBodyBlockDataPool signals that a nil body block pool has been provided
var ErrNilBodyBlockDataPool = errors.New("nil body block data pool")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilTemplateObj signals that an operation has been attempted to or with a nil template object
var ErrNilTemplateObj = errors.New("nil TemplateObj")
