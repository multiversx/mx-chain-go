package process

import "errors"

// ErrNilNodeHandler signals that a nil node handler has been provided
var ErrNilNodeHandler = errors.New("nil node handler")

// ErrNilOriginalHeader signals that a nil header was provided as competing block base
var ErrNilOriginalHeader = errors.New("nil original header")
