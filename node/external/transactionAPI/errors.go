package transactionAPI

import "errors"

// ErrTransactionNotFound signals that a transaction was not found
var ErrTransactionNotFound = errors.New("transaction not found")

// ErrCannotRetrieveTransaction signals that a transaction was not found
var ErrCannotRetrieveTransaction = errors.New("transaction cannot be retrieved")

// ErrNilStatusComputer signals that user account has a nil data trie
var ErrNilStatusComputer = errors.New("nil transaction status computer")
