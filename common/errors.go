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

// ErrAlreadyExistingEquivalentProof signals that the provided proof was already exiting in the pool
var ErrAlreadyExistingEquivalentProof = errors.New("already existing equivalent proof")
