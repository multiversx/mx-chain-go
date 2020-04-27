package factory

import "errors"

// ErrNilConfiguration signals that a nil configuration has been provided
var ErrNilConfiguration = errors.New("nil configuration provided")

// ErrNilPathManager signals that a nil path manager has been provided
var ErrNilPathManager = errors.New("nil path manager provided")

// ErrHasherCreation signals that the hasher cannot be created based on provided data
var ErrHasherCreation = errors.New("error creating hasher")

// ErrMarshalizerCreation signals that the marshalizer cannot be created based on provided data
var ErrMarshalizerCreation = errors.New("error creating marshalizer")
