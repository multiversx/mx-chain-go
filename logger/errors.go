package logger

import "errors"

// ErrNilWriter signals that a nil writer has been provided
var ErrNilWriter = errors.New("nil writer provided")

// ErrNilFormatter signals that a nil formatter has been provided
var ErrNilFormatter = errors.New("nil formatter provided")

// ErrInvalidLogLevelPattern signals that an un-parsable log level and patter was provided
var ErrInvalidLogLevelPattern = errors.New("un-parsable log level and pattern provided")
