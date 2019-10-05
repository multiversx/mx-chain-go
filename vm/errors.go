package vm

import "errors"

// ErrUnknownSystemSmartContract signals that there is no system smart contract on the provided address
var ErrUnknownSystemSmartContract = errors.New("missing system smart contract on selected address")

// ErrNilSystemEnvironmentInterface signals that a nil system environment interface was provided
var ErrNilSystemEnvironmentInterface = errors.New("system environment interface is nil")

// ErrNilSystemContractsContainer signals that the provided system contract container is nil
var ErrNilSystemContractsContainer = errors.New("system contract container is nil")

// ErrNilVMType signals that the provided vm type is nil
var ErrNilVMType = errors.New("vm type is nil")

// ErrCannotCreateNewSystemSmartContract signals that creation of new system smart contracts is not available
var ErrCannotCreateNewSystemSmartContract = errors.New("creation of new system smart contract is not enabled")

// ErrInputArgsIsNil signals that input arguments are nil for system smart contract
var ErrInputArgsIsNil = errors.New("input system smart contract arguments are nil")

// ErrInputCallValueIsNil signals that input call value is nil for system smart contract
var ErrInputCallValueIsNil = errors.New("input value for system smart contract is nil")

// ErrInputFunctionIsNil signals that input function is nil for system smart contract
var ErrInputFunctionIsNil = errors.New("input function for system smart contract is nil")

// ErrInputCallerAddrIsNil signals that input caller address is nil for system smart contract
var ErrInputCallerAddrIsNil = errors.New("input called address for system smart contract is nil")

// ErrInputRecipientAddrIsNil signals that input recipient address for system smart contract is nil
var ErrInputRecipientAddrIsNil = errors.New("input recipient address for system smart contract is nil")

// ErrInputGasProvidedIsNil signals that input gas provided value is nil for system smart contracts
var ErrInputGasProvidedIsNil = errors.New("input gas provided for system smart contract is nil")

// ErrInputGasPriceIsNil signals that input gas price value is nil for system smart contracts
var ErrInputGasPriceIsNil = errors.New("input gas price for system smart contract is nil")
