package executionManager

import "errors"

// ErrNilHeadersExecutor signals that a nil headers executor has been provided
var ErrNilHeadersExecutor = errors.New("nil headers executor")

// ErrNilBlocksQueue signals that a nil blocks queue has been provided
var ErrNilBlocksQueue = errors.New("nil blocks queue")

// ErrNilExecutionResultsTracker signals that a nil execution results tracker has been provided
var ErrNilExecutionResultsTracker = errors.New("nil execution results tracker")

// ErrNilBlockchain signals that a nil blockchain has been provided
var ErrNilBlockchain = errors.New("nil blockchain")

// ErrNilHeadersPool signals that a nil headers pool has been provided
var ErrNilHeadersPool = errors.New("nil headers pool")
