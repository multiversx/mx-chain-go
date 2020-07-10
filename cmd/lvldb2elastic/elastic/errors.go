package elastic

import "errors"

// ErrInvalidElasticURL signals that the provided elastic search URL is not valid
var ErrInvalidElasticURL = errors.New("invalid elastic search URL")

// ErrEmptyElasticUsername signals that an empty elastic search username has been provided
var ErrEmptyElasticUsername = errors.New("empty elastic search username")

// ErrEmptyElasticPassword signals that an empty elastic search password has been provided
var ErrEmptyElasticPassword = errors.New("empty elastic search password")

// ErrNilPubKeyConverter signals that a nil public key converter has been passed
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")
