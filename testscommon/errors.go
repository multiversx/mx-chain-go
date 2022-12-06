package testscommon

import "errors"

// ErrNotImplemented signals that a method is not implemented for an interface implementation
var ErrNotImplemented = errors.New("not implemented")
