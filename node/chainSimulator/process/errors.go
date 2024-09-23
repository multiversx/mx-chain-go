package process

import "errors"

// ErrNilNodeHandler signals that a nil node handler has been provided
var ErrNilNodeHandler = errors.New("nil node handler")
