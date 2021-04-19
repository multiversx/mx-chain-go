package vmcommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/esdt"
)

// FunctionNames (alias) is a map of function names
type FunctionNames = map[string]struct{}

// BlockchainHook is the interface for VM blockchain callbacks
type BlockchainHook interface {
	// NewAddress yields the address of a new SC account, when one such account is created.
	// The result should only depend on the creator address and nonce.
	// Returning an empty address lets the VM decide what the new address should be.
	NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)

	// GetStorageData should yield the storage value for a certain account and index.
	// Should return an empty byte array if the key is missing from the account storage,
	// or if account does not exist.
	GetStorageData(accountAddress []byte, index []byte) ([]byte, error)

	// GetBlockhash returns the hash of the block with the asked nonce if available
	GetBlockhash(nonce uint64) ([]byte, error)

	// LastNonce returns the nonce from from the last committed block
	LastNonce() uint64

	// LastRound returns the round from the last committed block
	LastRound() uint64

	// LastTimeStamp returns the timeStamp from the last committed block
	LastTimeStamp() uint64

	// LastRandomSeed returns the random seed from the last committed block
	LastRandomSeed() []byte

	// LastEpoch returns the epoch from the last committed block
	LastEpoch() uint32

	// GetStateRootHash returns the state root hash from the last committed block
	GetStateRootHash() []byte

	// CurrentNonce returns the nonce from the current block
	CurrentNonce() uint64

	// CurrentRound returns the round from the current block
	CurrentRound() uint64

	// CurrentTimeStamp return the timestamp from the current block
	CurrentTimeStamp() uint64

	// CurrentRandomSeed returns the random seed from the current header
	CurrentRandomSeed() []byte

	// CurrentEpoch returns the current epoch
	CurrentEpoch() uint32

	// ProcessBuiltInFunction will process the builtIn function for the created input
	ProcessBuiltInFunction(input *ContractCallInput) (*VMOutput, error)

	// GetBuiltinFunctionNames returns the names of protocol built-in functions
	GetBuiltinFunctionNames() FunctionNames

	// GetAllState returns the full state of the account, all the key-value saved
	GetAllState(address []byte) (map[string][]byte, error)

	// GetUserAccount returns a user account
	GetUserAccount(address []byte) (UserAccountHandler, error)

	// GetCode returns the code for the given account
	GetCode(UserAccountHandler) []byte

	// GetShardOfAddress returns the shard ID of a given address
	GetShardOfAddress(address []byte) uint32

	// IsSmartContract returns whether the address points to a smart contract
	IsSmartContract(address []byte) bool

	// IsPayable checks weather the provided address can receive ERD or not
	IsPayable(address []byte) (bool, error)

	// SaveCompiledCode saves to cache and storage the compiled code
	SaveCompiledCode(codeHash []byte, code []byte)

	// GetCompiledCode returns the compiled code if it finds in the cache or storage
	GetCompiledCode(codeHash []byte) (bool, []byte)

	// ClearCompiledCodes clears the cache and storage of compiled codes
	ClearCompiledCodes()

	// GetESDTToken loads the ESDT digital token for the given key
	GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// UserAccountHandler defines a user account
type UserAccountHandler interface {
	AddressBytes() []byte
	GetNonce() uint64
	GetCodeMetadata() []byte
	GetCodeHash() []byte
	GetRootHash() []byte
	GetBalance() *big.Int
	GetDeveloperReward() *big.Int
	GetOwnerAddress() []byte
	GetUserName() []byte
	IsInterfaceNil() bool
}

// VMExecutionHandler interface for any Elrond VM endpoint
type VMExecutionHandler interface {
	// RunSmartContractCreate computes how a smart contract creation should be performed
	RunSmartContractCreate(input *ContractCreateInput) (*VMOutput, error)

	// RunSmartContractCall computes the result of a smart contract call and how the system must change after the execution
	RunSmartContractCall(input *ContractCallInput) (*VMOutput, error)

	// GasScheduleChange sets a new gas schedule for the VM
	GasScheduleChange(newGasSchedule map[string]map[string]uint64)

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// CryptoHook interface for VM krypto functions
type CryptoHook interface {
	// Sha256 cryptographic function
	Sha256(data []byte) ([]byte, error)

	// Keccak256 cryptographic function
	Keccak256(data []byte) ([]byte, error)

	// Ripemd160 cryptographic function
	Ripemd160(data []byte) ([]byte, error)

	// Ecrecover calculates the corresponding Ethereum address for the public key which created the given signature
	// https://ewasm.readthedocs.io/en/mkdocs/system_contracts/
	Ecrecover(hash []byte, recoveryID []byte, r []byte, s []byte) ([]byte, error)

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}
