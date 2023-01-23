package common

import "errors"

// ErrInvalidTimeout signals that an invalid timeout period has been provided
var ErrInvalidTimeout = errors.New("invalid timeout value")

// ErrNilWasmChangeLocker signals that a nil wasm change locker has been provided
var ErrNilWasmChangeLocker = errors.New("nil wasm change locker")

// ErrNilExclusionTree signals that a nil exclusion tree was provided
var ErrNilExclusionTree = errors.New("nil exclusion tree")
