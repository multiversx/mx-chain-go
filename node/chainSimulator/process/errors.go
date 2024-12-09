package process

import "errors"

// ErrNilNodeHandler signals that a nil node handler has been provided
var ErrNilNodeHandler = errors.New("nil node handler")

// ErrNilBlockProcessor signals that a nil block process has been provided
var ErrNilBlockProcessor = errors.New("nil block processor")
