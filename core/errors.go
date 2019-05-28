package core

import (
	"errors"
)

// ErrNilMarshalizer signals that a new marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrInvalidValue signals that a nil value has been provided
var ErrInvalidValue = errors.New("invalid value provided")

// ErrNilInputData signals that a nil data has been provided
var ErrNilInputData = errors.New("nil input data")
