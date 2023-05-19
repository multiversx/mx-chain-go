package notifier

import (
	"errors"
)

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilHTTPClientWrapper signals that a nil http client wrapper has been provided
var ErrNilHTTPClientWrapper = errors.New("nil http client wrapper")

// ErrNilMarshaller signals that a nil marshaller has been provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilBlockContainerHandler signals that a nil block container handler has been provided
var ErrNilBlockContainerHandler = errors.New("nil bock container handler")
