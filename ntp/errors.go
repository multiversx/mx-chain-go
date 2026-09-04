package ntp

import (
	"errors"
)

// ErrIndexOutOfBounds is raised when an out of bound index is used
var ErrIndexOutOfBounds = errors.New("index is out of bounds")

// ErrNoClockOffsets is raised when no clock offsets are available
var ErrNoClockOffsets = errors.New("no clock offsets available")
