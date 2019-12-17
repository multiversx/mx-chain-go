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

// ErrNilBlockchainHook signals that blockchain hook is nil
var ErrNilBlockchainHook = errors.New("blockchain hook is nil")

// ErrNilCryptoHook signals that crypto hook is nil
var ErrNilCryptoHook = errors.New("crypto hook is nil")

// ErrNilOrEmptyKey signals that key is nil or empty
var ErrNilOrEmptyKey = errors.New("nil or empty key")

// ErrNilInitialStakeValue signals that nil initial stake value was provided
var ErrNilInitialStakeValue = errors.New("initial stake value is nil")

// ErrNilEconomicsData signals that nil economics data has been provided
var ErrNilEconomicsData = errors.New("nil economics data")

// ErrNegativeInitialStakeValue signals that a negative initial stake value was provided
var ErrNegativeInitialStakeValue = errors.New("initial stake value is negative")

// ErrNotEnoughQualifiedNodes signals that there are insufficient number of qualified nodes
var ErrNotEnoughQualifiedNodes = errors.New("not enough qualified nodes")

// ErrBLSPublicKeyMissmatch signals that public keys do not match
var ErrBLSPublicKeyMissmatch = errors.New("public key missmatch")

// ErrKeyAlreadyRegistered signals that bls key is already registered
var ErrKeyAlreadyRegistered = errors.New("bls key already registered")

// ErrBLSKeyIsNotValid signals that bls key is invalid
var ErrBLSKeyIsNotValid = errors.New("bls key is invalid")

// ErrBLSKeyIsNotStaked signals that bls key is not staked
var ErrBLSKeyIsNotStaked = errors.New("bls key is not staked")

// ErrNotEnoughArgumentsToStake signals that the arguments provided are not enough
var ErrNotEnoughArgumentsToStake = errors.New("not enough arguments to stake")

// ErrStillInUnBoundPeriod signals that bls key is in unbond period
var ErrStillInUnBoundPeriod = errors.New("bls key is in unbond period")
