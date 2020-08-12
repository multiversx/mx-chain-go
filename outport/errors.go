package outport

import (
	"errors"
)

// ErrNilTxCoordinator signals that a dependency is nil
var ErrNilTxCoordinator = errors.New("tx coordinator is nil")

// ErrNilLogsProcessor signals that a dependency is nil
var ErrNilLogsProcessor = errors.New("logs processor is nil")

// ErrNilSender signals that a dependency is nil
var ErrNilSender = errors.New("sender is nil")
