package asyncExecution

import "errors"

// ErrNilHeadersQueue signals that a nil headers queue has been provided
var ErrNilHeadersQueue = errors.New("nil headers queue")

// ErrNilExecutionTracker signals that a nil execution tracker has been provided
var ErrNilExecutionTracker = errors.New("nil execution tracker")

// ErrNilBlockProcessor signals that a nil block processor has been provided
var ErrNilBlockProcessor = errors.New("nil block processor")
