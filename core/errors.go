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
