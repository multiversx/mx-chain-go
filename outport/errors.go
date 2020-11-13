package outport

import "errors"

// ErrNilDriver signals that a nil driver has been provided
var ErrNilDriver = errors.New("nil driver")

//ErrNilUrl signals that the provided url is empty
var ErrNilUrl = errors.New("url is empty")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher provided")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilAccountsDB signals that a nil accounts database has been provided
var ErrNilAccountsDB = errors.New("nil accounts db")

// ErrNilFeeConfig signals that a nil fee config is provided
var ErrNilFeeConfig = errors.New("nil fee config")
