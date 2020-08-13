package elastic

import "errors"

// ErrNilPubKeyConverter signals that a nil public key converter has been passed
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")
