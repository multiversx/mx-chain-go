package trieIterators

import "errors"

// ErrCannotCastAccountHandlerToUserAccount signal that returned account is wrong
var ErrCannotCastAccountHandlerToUserAccount = errors.New("cannot cast AccountHandler to UserAccount")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

// ErrNilQueryService signals that a nil query service has been provided
var ErrNilQueryService = errors.New("nil query service")

// ErrNilBlockChain signals that an operation has been attempted to or with a nil blockchain
var ErrNilBlockChain = errors.New("nil block chain")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNodeNotInitialized signals that the node is not initialized and can not compute the required task yet
var ErrNodeNotInitialized = errors.New("the node is not fully initialized")

// ErrNilMutex signals that a nil mutex hax been provided
var ErrNilMutex = errors.New("nil mutex")
