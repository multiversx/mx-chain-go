package exTransaction

import (
	"errors"
)

// ErrNilAccountsAdapter defines the error when trying to use a nil AccountsAddapter
var ErrNilAccountsAdapter = errors.New("nil AccountsAdapter")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilAddressConverter signals that an operation has been attempted to or with a nil AddressConverter implementation
var ErrNilAddressConverter = errors.New("nil AddressConverter")

// ErrNilTransaction signals that an operation has been attempted to or with a nil transaction
var ErrNilTransaction = errors.New("nil transaction")

// ErrNoVM signals that no SChandler has been set
var ErrNoVM = errors.New("no VM (hook not set)")

// ErrHigherNonceInTransaction signals the nonce in transaction is higher than the account's nonce
var ErrHigherNonceInTransaction = errors.New("higher nonce in transaction, delaying execution")

// ErrLowerNonceInTransaction signals the nonce in transaction is lower than the account's nonce
var ErrLowerNonceInTransaction = errors.New("lower nonce in transaction, dropping")

// ErrInsufficientFunds signals the funds are insufficient
var ErrInsufficientFunds = errors.New("insufficient funds")
