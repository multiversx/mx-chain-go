package core

import (
	"errors"
)

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrNilHasher signals that a nil marshalizer has been provided
var ErrNilHasher = errors.New("nil hashser provided")

// ErrNilCoordinator signals that a nil shardCoordinator has been provided
var ErrNilCoordinator = errors.New("nil coordinator provided")

// ErrNilLogger signals that a nil logger has been provided
var ErrNilLogger = errors.New("nil logger provided")

// ErrInvalidValue signals that a nil value has been provided
var ErrInvalidValue = errors.New("invalid value provided")

// ErrNilInputData signals that a nil data has been provided
var ErrNilInputData = errors.New("nil input data")

//ErrNilUrl signals that the provided url is empty
var ErrNilUrl = errors.New("url is empty")

// ErrWrongTypeAssertion signals that an type assertion failed
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrPemFileIsInvalid signals that a pem file is invalid
var ErrPemFileIsInvalid = errors.New("pem file is invalid")

// ErrNilFile signals that a nil file has been provided
var ErrNilFile = errors.New("nil file provided")

// ErrEmptyFile signals that a empty file has been provided
var ErrEmptyFile = errors.New("empty file provided")

// ErrInvalidIndex signals that an invalid private key index has been provided
var ErrInvalidIndex = errors.New("invalid private key index")

// ErrNotPositiveValue signals that a 0 or negative value has been provided
var ErrNotPositiveValue = errors.New("the provided value is not positive")

// ErrNilAppStatusHandler signals that a nil status handler has been provided
var ErrNilAppStatusHandler = errors.New("appStatusHandler is nil")
