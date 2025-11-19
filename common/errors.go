package common

import "errors"

// ErrInvalidTimeout signals that an invalid timeout period has been provided
var ErrInvalidTimeout = errors.New("invalid timeout value")

// ErrNilWasmChangeLocker signals that a nil wasm change locker has been provided
var ErrNilWasmChangeLocker = errors.New("nil wasm change locker")

// ErrNilStateSyncNotifierSubscriber signals that a nil state sync notifier subscriber has been provided
var ErrNilStateSyncNotifierSubscriber = errors.New("nil state sync notifier subscriber")

// ErrInvalidHeaderProof signals that an invalid equivalent proof has been provided
var ErrInvalidHeaderProof = errors.New("invalid equivalent proof")

// ErrNilHeaderProof signals that a nil equivalent proof has been provided
var ErrNilHeaderProof = errors.New("nil equivalent proof")

// ErrAlreadyExistingEquivalentProof signals that the provided proof was already exiting in the pool
var ErrAlreadyExistingEquivalentProof = errors.New("already existing equivalent proof")

// ErrNilHeaderHandler signals that a nil header handler has been provided
var ErrNilHeaderHandler = errors.New("nil header handler")

// ErrNotEnoughSignatures defines the error for not enough signatures
var ErrNotEnoughSignatures = errors.New("not enough signatures")

// ErrWrongSizeBitmap signals that the provided bitmap's length is bigger than the one that was required
var ErrWrongSizeBitmap = errors.New("wrong size bitmap has been provided")

// ErrInvalidHashShardKey signals that the provided hash-shard key is invalid
var ErrInvalidHashShardKey = errors.New("invalid hash shard key")

// ErrInvalidNonceShardKey signals that the provided nonce-shard key is invalid
var ErrInvalidNonceShardKey = errors.New("invalid nonce shard key")

// ErrNilCommonConfigsHandler signals that a nil common configs handler has been provided
var ErrNilCommonConfigsHandler = errors.New("nil common configs handler")

// ErrWrongTypeAssertion signals that a type assertion failed
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilBaseExecutionResult signals that a nil base execution result has been provided
var ErrNilBaseExecutionResult = errors.New("nil base execution result")

// ErrNilLastExecutionResultHandler signals that a nil last execution result handler has been provided
var ErrNilLastExecutionResultHandler = errors.New("nil last execution result handler")

// ErrMissingHeader signals that header of the block is missing
var ErrMissingHeader = errors.New("missing header")

// ErrMissingMiniBlock signals that mini block is missing
var ErrMissingMiniBlock = errors.New("missing mini block")

// ErrMissingCachedTransactions signals that cached transactions are missing
var ErrMissingCachedTransactions = errors.New("missing cached transactions")

// ErrMissingCachedLogs signals that cached logs events are missing
var ErrMissingCachedLogs = errors.New("missing cached logs")
