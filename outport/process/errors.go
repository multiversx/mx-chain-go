package process

import "errors"

// ErrNilGasConsumedProvider signals that a nil gas consumed provider has been provided
var ErrNilGasConsumedProvider = errors.New("nil gas consumed provider")

// ErrNilNodesCoordinator is raised when a valid validator group selector is expected but nil used
var ErrNilNodesCoordinator = errors.New("validator group selector is nil")

// ErrNilTransactionCoordinator signals that transaction coordinator is nil
var ErrNilTransactionCoordinator = errors.New("transaction coordinator is nil")

// ErrNilHeaderHandler signal that provided header handler is nil
var ErrNilHeaderHandler = errors.New("nil header handler")

// ErrNilBodyHandler signal that provided body handler is nil
var ErrNilBodyHandler = errors.New("nil body handler")

var errCannotCastTransaction = errors.New("cannot cast transaction")

var errCannotCastSCR = errors.New("cannot cast smart contract result")

var errCannotCastReward = errors.New("cannot cast reward transaction")

var errCannotCastReceipt = errors.New("cannot cast receipt transaction")

var errCannotCastLog = errors.New("cannot cast log")

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

// ErrExecutedTxNotFoundInOrderedTxs signals that a transaction that was executed was not found in the ordered transactions
var ErrExecutedTxNotFoundInOrderedTxs = errors.New("executed tx not found in ordered txs")

// ErrOrderedTxNotFound signals that an ordered tx hash was not found in the pool
var ErrOrderedTxNotFound = errors.New("ordered tx not found in pool")

// ErrNilMiniBlockHeaderHandler signals that a nil miniBlock header handler has been provided
var ErrNilMiniBlockHeaderHandler = errors.New("nil miniBlock header handler")

// ErrNilMiniBlock signals that a nil miniBlock has been provided
var ErrNilMiniBlock = errors.New("nil miniBlock")

// ErrNilExecutedTxHashes signals that a nil executed tx hashes map has been provided
var ErrNilExecutedTxHashes = errors.New("nil executed tx hashes map")

// ErrNilOrderedTxHashes signals that a nil ordered tx list has been provided
var ErrNilOrderedTxHashes = errors.New("nil ordered tx list")

// ErrIndexOutOfBounds signals that an index is out of bounds
var ErrIndexOutOfBounds = errors.New("index out of bounds")
