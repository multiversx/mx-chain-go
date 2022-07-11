package trieIterators

import "errors"

// ErrCannotCastAccountHandlerToUserAccount signal that returned account is wrong
var ErrCannotCastAccountHandlerToUserAccount = errors.New("cannot cast AccountHandler to UserAccount")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

// ErrNilQueryService signals that a nil query service has been provided
var ErrNilQueryService = errors.New("nil query service")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilMutex signals that a nil mutex has been provided
var ErrNilMutex = errors.New("nil mutex")

// ErrTrieOperationsTimeout signals a timeout during trie operations
var ErrTrieOperationsTimeout = errors.New("trie operations timeout")
