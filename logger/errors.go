package logger

import "errors"

// ErrNilWriter signals that a nil writer has been provided
var ErrNilWriter = errors.New("nil writer provided")

// ErrNilFormatter signals that a nil formatter has been provided
var ErrNilFormatter = errors.New("nil formatter provided")

// ErrInvalidLogLevelPattern signals that an un-parsable log level and patter was provided
var ErrInvalidLogLevelPattern = errors.New("un-parsable log level and pattern provided")

// ErrWriterNotFound signals that the provided writer was not found while searching container list
var ErrWriterNotFound = errors.New("writer not found while searching container")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilDisplayByteSliceHandler signals that a nil display byte slice handler has been provided
var ErrNilDisplayByteSliceHandler = errors.New("nil display byte slice handler")
