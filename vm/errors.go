package vm

import "errors"

// ErrNilBlockchainHook signals that nil blockchain hook was provided
var ErrNilBlockchainHook = errors.New("nil blockchain hook")

// ErrNilCryptoHook signals that nil crypto hook was provided
var ErrNilCryptoHook = errors.New("nil crypto hook")

// ErrUnknownSystemSmartContract signals that there is no system smart contract on the provided address
var ErrUnknownSystemSmartContract = errors.New("missing system smart contract on selected address")
