package typeConverters

import (
	"errors"
)

// ErrNilByteSlice signals that a nil byte slice has been provided
var ErrNilByteSlice = errors.New("nil byte slice")

// ErrByteSliceLenShouldHaveBeen8 signals that the byte slice provided has a wrong length (should have been 8)
var ErrByteSliceLenShouldHaveBeen8 = errors.New("byte slice's len should have been 8")
