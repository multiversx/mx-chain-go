package asyncExecution

import "errors"

// ErrNilHeadersCache signals that a nil headers cache has been provided
var ErrNilHeadersCache = errors.New("nil headers cache")

// ErrNilExecutionTracker signals that a nil execution tracker has been provided
var ErrNilExecutionTracker = errors.New("nil execution tracker")

// ErrNilBlockProcessor signals that a nil block processor has been provided
var ErrNilBlockProcessor = errors.New("nil block processor")

// ErrNilExecutionResult signals that a nil execution result has been provided
var ErrNilExecutionResult = errors.New("nil execution result")
