package blockAPI

import "errors"

// ErrInvalidOutputFormat signals that the outport format type is not valid
var ErrInvalidOutputFormat = errors.New("the outport format type is invalid")
