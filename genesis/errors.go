package genesis

import "errors"

// ErrNilEntireSupply signals that the provided entire supply is nil
var ErrNilEntireSupply = errors.New("nil entire supply")

// ErrInvalidEntireSupply signals that the provided entire supply is invalid
var ErrInvalidEntireSupply = errors.New("invalid entire supply")

// ErrEntireSupplyMismatch signals that the provided entire supply mismatches the computed one
var ErrEntireSupplyMismatch = errors.New("entire supply mismatch")

// ErrEmptyAddress signals that an empty address was found in genesis file
var ErrEmptyAddress = errors.New("empty address")

// ErrInvalidAddress signals that an invalid address was found
var ErrInvalidAddress = errors.New("invalid address")

// ErrInvalidSupplyString signals that the supply string is not a valid number
var ErrInvalidSupplyString = errors.New("invalid supply string")

// ErrInvalidBalanceString signals that the balance string is not a valid number
var ErrInvalidBalanceString = errors.New("invalid balance string")

// ErrInvalidStakingBalanceString signals that the staking balance string is not a valid number
var ErrInvalidStakingBalanceString = errors.New("invalid staking balance string")

// ErrInvalidDelegationValueString signals that the delegation balance string is not a valid number
var ErrInvalidDelegationValueString = errors.New("invalid delegation value string")

// ErrEmptyDelegationAddress signals that the delegation address is empty
var ErrEmptyDelegationAddress = errors.New("empty delegation address")

// ErrInvalidDelegationAddress signals that the delegation address is invalid
var ErrInvalidDelegationAddress = errors.New("invalid delegation address")

// ErrInvalidSupply signals that the supply field is invalid
var ErrInvalidSupply = errors.New("invalid supply")

// ErrInvalidBalance signals that the balance field is invalid
var ErrInvalidBalance = errors.New("invalid balance")

// ErrInvalidStakingBalance signals that the staking balance field is invalid
var ErrInvalidStakingBalance = errors.New("invalid staking balance")

// ErrInvalidDelegationValue signals that the delegation value field is invalid
var ErrInvalidDelegationValue = errors.New("invalid delegation value")

// ErrSupplyMismatch signals that the supply value provided is not valid when summing the other fields
var ErrSupplyMismatch = errors.New("supply value mismatch")

// ErrDuplicateAddress signals that the same address was found more than one time
var ErrDuplicateAddress = errors.New("duplicate address")

// ErrAddressIsSmartContract signals that provided address is of type smart contract
var ErrAddressIsSmartContract = errors.New("address is a smart contract")

// ErrNilShardCoordinator signals that the provided shard coordinator is nil
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilPubkeyConverter signals that the provided public key converter is nil
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilGenesisParser signals that the provided genesis parser is nil
var ErrNilGenesisParser = errors.New("nil genesis parser")
