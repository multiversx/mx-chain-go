package host

import "errors"

// ErrHostIsClosed signals that the host was closed while trying to perform actions
var ErrHostIsClosed = errors.New("server is closed")

// ErrNilHost signals that a nil host has been provided
var ErrNilHost = errors.New("nil host provided")

// ErrNilStateChangesCollector signals that nil state changes collector has been provided
var ErrNilStateChangesCollector = errors.New("nil state changes collector provided")
