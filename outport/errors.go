package outport

import (
	"errors"
)

// ErrNilTxCoordinator signals that a dependency is nil
var ErrNilTxCoordinator = errors.New("tx coordinator is nil")

// ErrNilTxLogsProcessor signals that a dependency is nil
var ErrNilLogsProcessor = errors.New("logs processor is nil")

// ErrNilMarshalizer signals that a dependency is nil
var ErrNilMarshalizer = errors.New("marshalizer is nil")
