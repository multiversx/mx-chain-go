package execution

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
var ErrHigherNonceInTransaction = errors.New("higher nonce in transaction")

// ErrLowerNonceInTransaction signals the nonce in transaction is lower than the account's nonce
var ErrLowerNonceInTransaction = errors.New("lower nonce in transaction")

// ErrInsufficientFunds signals the funds are insufficient
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrNilValue signals the value is nil
var ErrNilValue = errors.New("nil value")

// ErrNilBlockchain signals that an operation has been attempted to or with a nil blockchain
var ErrNilBlockChain = errors.New("nil block chain")

// ErrNilBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilBlockBody = errors.New("nil block body")

// ErrNilBlockHeader signals that an operation has been attempted to or with a nil block header
var ErrNilBlockHeader = errors.New("nil block header")

// ErrHigherNonceInBlock signals the nonce in block is higher than next expected nounce
var ErrHigherNonceInBlock = errors.New("higher nonce in block, waiting for the right one")

// ErrLowerNonceInBlock signals the nonce in block is lower than next expected nounce
var ErrLowerNonceInBlock = errors.New("lower nonce in block")

// ErrWrongNonceInBlock signals the nonce in block is different than expected nounce
var ErrWrongNonceInBlock = errors.New("wrong nonce in block")

// ErrInvalidBlockHash signals the hash of the block is not matching with the previous one
var ErrInvalidBlockHash = errors.New("invalid block hash")

// ErrInvalidBlockSignature signals the signature of the block is not valid
var ErrInvalidBlockSignature = errors.New("invalid block signature")

// ErrMissingTransaction signals that one transaction is missing
var ErrMissingTransaction = errors.New("missing transaction")

// ErrMarshalWithoutSuccess signals that marshal some data was not done with success
var ErrMarshalWithoutSuccess = errors.New("marshal without success")

// ErrPersistWithoutSuccess signals that persist some data was not done with success
var ErrPersistWithoutSuccess = errors.New("persist without success")

// ErrRootStateMissmatch signals that persist some data was not done with success
var ErrRootStateMissmatch = errors.New("Root state does not match")

// ErrAccountStateDirty signals that the accounts were modified before starting the current modifications
var ErrAccountStateDirty = errors.New("AccountState was dirty before starting to change")

// ErrMissingHeader signals that header of the block is missing
var ErrMissingHeader = errors.New("missing header")

// ErrMissingBody signals that body of the block is missing
var ErrMissingBody = errors.New("missing body")

// ErrNilBlockExecutor signals that an operation has been attempted to or with a nil BlockExecutor implementation
var ErrNilBlockExecutor = errors.New("nil BlockExecutor")

// ErrNilBlockPool signals that an operation has been attempted to or with a nil BlockPool
var ErrNilBlockPool = errors.New("nil BlockPool")

// ErrNilRound signals that an operation has been attempted to or with a nil Round
var ErrNilRound = errors.New("nil Round")
