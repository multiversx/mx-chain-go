package executionTrack

import "errors"

var (
	// ErrNilExecutionResult signals  that a nil execution results has been provided
	ErrNilExecutionResult = errors.New("nil execution result")

	// ErrCannotFindExecutionResult signals that an execution result cannot be found
	ErrCannotFindExecutionResult = errors.New("cannot find execution result")
)
