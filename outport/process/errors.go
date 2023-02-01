package process

import "errors"

// ErrNilGasConsumedProvider signals that a nil gas consumed provider has been provided
var ErrNilGasConsumedProvider = errors.New("nil gas consumed provider")

// ErrNilNodesCoordinator is raised when a valid validator group selector is expected but nil used
var ErrNilNodesCoordinator = errors.New("validator group selector is nil")

// ErrNilTransactionCoordinator signals that transaction coordinator is nil
var ErrNilTransactionCoordinator = errors.New("transaction coordinator is nil")

// errNilHeaderHandler signal that provided header handler is nil
var errNilHeaderHandler = errors.New("nil header handler")

// errNilBodyHandler signal that provided body handler is nil
var errNilBodyHandler = errors.New("nil body handler")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher provided")

// ErrNilStorer is raised when a nil storer has been provided
var ErrNilStorer = errors.New("nil storer")

// ErrNilEnableEpochsHandler signals that a nil enable epochs handler has been provided
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")

// ErrNilExecutionOrderGetter signals that a nil execution order getter has been provided
var ErrNilExecutionOrderGetter = errors.New("nil execution order getter")

// ErrWrongTypeAssertion signals that a wrong type assertion has been done
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrMiniBlocksHeadersMismatch signals that the number of miniBlock headers is different than the number of miniBlocks
var ErrMiniBlocksHeadersMismatch = errors.New("mini blocks headers mismatch")

// ErrTransactionNotFoundInBody signals that a transaction was not found in the block body
var ErrTransactionNotFoundInBody = errors.New("transaction not found in body")

// ErrOrderedTxNotFound signals that an ordered tx hash was not found in the pool
var ErrOrderedTxNotFound = errors.New("ordered tx not found in pool")
