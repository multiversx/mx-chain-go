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
