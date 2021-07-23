package vm

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SystemSmartContract interface defines the function a system smart contract should have
type SystemSmartContract interface {
	Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
	CanUseContract() bool
	SetNewGasCost(gasCost GasCost)
	IsInterfaceNil() bool
}

// SystemSCContainerFactory defines the functionality to create a system smart contract container
type SystemSCContainerFactory interface {
	CreateForGenesis() (SystemSCContainer, error)
	Create() (SystemSCContainer, error)
	IsInterfaceNil() bool
}

// SystemSCContainer defines a system smart contract holder data type with basic functionality
type SystemSCContainer interface {
	Get(key []byte) (SystemSmartContract, error)
	Add(key []byte, val SystemSmartContract) error
	Replace(key []byte, val SystemSmartContract) error
	Remove(key []byte)
	Len() int
	Keys() [][]byte
	IsInterfaceNil() bool
}

// SystemEI defines the environment interface system smart contract can use
type SystemEI interface {
	ExecuteOnDestContext(destination []byte, sender []byte, value *big.Int, input []byte) (*vmcommon.VMOutput, error)
	DeploySystemSC(baseContract []byte, newAddress []byte, ownerAddress []byte, initFunction string, value *big.Int, input [][]byte) (vmcommon.ReturnCode, error)
	Transfer(destination []byte, sender []byte, value *big.Int, input []byte, gasLimit uint64) error
	SendGlobalSettingToAll(sender []byte, input []byte)
	GetBalance(addr []byte) *big.Int
	SetStorage(key []byte, value []byte)
	SetStorageForAddress(address []byte, key []byte, value []byte)
	AddReturnMessage(msg string)
	GetStorage(key []byte) []byte
	GetStorageFromAddress(address []byte, key []byte) []byte
	Finish(value []byte)
	UseGas(gasToConsume uint64) error
	GasLeft() uint64
	BlockChainHook() BlockchainHook
	CryptoHook() vmcommon.CryptoHook
	IsValidator(blsKey []byte) bool
	StatusFromValidatorStatistics(blsKey []byte) string
	CanUnJail(blsKey []byte) bool
	IsBadRating(blsKey []byte) bool
	CleanStorageUpdates()

	IsInterfaceNil() bool
}

// EconomicsHandler defines the methods to get data from the economics component
type EconomicsHandler interface {
	GenesisTotalSupply() *big.Int
	IsInterfaceNil() bool
}

// ContextHandler defines the methods needed to execute system smart contracts
type ContextHandler interface {
	SystemEI

	GetContract(address []byte) (SystemSmartContract, error)
	SetSystemSCContainer(scContainer SystemSCContainer) error
	CreateVMOutput() *vmcommon.VMOutput
	CleanCache()
	SetSCAddress(addr []byte)
	AddCode(addr []byte, code []byte)
	AddTxValueToSmartContract(value *big.Int, scAddress []byte)
	SetGasProvided(gasProvided uint64)
}

// MessageSignVerifier is used to verify if message was signed with given public key
type MessageSignVerifier interface {
	Verify(message []byte, signedMessage []byte, pubKey []byte) error
	IsInterfaceNil() bool
}

// ArgumentsParser defines the functionality to parse transaction data into arguments and code for smart contracts
type ArgumentsParser interface {
	ParseData(data string) (string, [][]byte, error)
	IsInterfaceNil() bool
}

// NodesConfigProvider defines the functionality which is needed for nodes config in system smart contracts
type NodesConfigProvider interface {
	MinNumberOfNodes() uint32
	MinNumberOfNodesWithHysteresis() uint32
	IsInterfaceNil() bool
}

// EpochNotifier can notify upon an epoch change and provide the current epoch
type EpochNotifier interface {
	RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler)
	CurrentEpoch() uint32
	CheckEpoch(header data.HeaderHandler)
	IsInterfaceNil() bool
}

// BlockchainHook is the interface for VM blockchain callbacks
type BlockchainHook interface {
	GetStorageData(accountAddress []byte, index []byte) ([]byte, error)
	CurrentNonce() uint64
	CurrentRound() uint64
	CurrentEpoch() uint32
	GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error)
	GetCode(account vmcommon.UserAccountHandler) []byte
	GetShardOfAddress(address []byte) uint32
	IsSmartContract(address []byte) bool
	IsPayable(address []byte) (bool, error)
	NumberOfShards() uint32
	CurrentRandomSeed() []byte
	Close() error
	GetSnapshot() int
	RevertToSnapshot(snapshot int) error
}
