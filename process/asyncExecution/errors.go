package asyncExecution

import "errors"

var (
	// ErrNilHeadersQueue signals that a nil headers queue has been provided
	ErrNilHeadersQueue error = errors.New("nil headers queue")
	// ErrNilExecutionTracker signals that a nil execution tracker has been provided
	ErrNilExecutionTracker error = errors.New("nil execution tracker")
	// ErrNilBlockProcessor signals that a nil block processor has been provided
	ErrNilBlockProcessor error = errors.New("nil block processor")
)
