package outport

import "errors"

// ErrNilDriver signals that a nil driver has been provided
var ErrNilDriver = errors.New("nil driver")

// ErrNilArgsOutportFactory signals that arguments that are needed for elastic driver factory are nil
var ErrNilArgsOutportFactory = errors.New("nil args outport driver factory")

// ErrInvalidRetrialInterval signals that an invalid retrial interval was provided
var ErrInvalidRetrialInterval = errors.New("invalid retrial interval")

// ErrNilPubKeyConverter signals that a nil pubkey converter has been provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

var errNilSaveBlockArgs = errors.New("nil save blocks args provided")

var errNilHeaderAndBodyArgs = errors.New("nil header and body args provided")
