package executionTrack

import "errors"

var (
	// ErrNilExecutionResult signals  that a nil execution results has been provided
	ErrNilExecutionResult = errors.New("nil execution result")
)
