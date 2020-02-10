package vm

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SystemSmartContract interface defines the function a system smart contract should have
type SystemSmartContract interface {
	Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
	IsInterfaceNil() bool
}

// SystemSCContainerFactory defines the functionality to create a system smart contract container
type SystemSCContainerFactory interface {
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
	Transfer(destination []byte, sender []byte, value *big.Int, input []byte) error
	GetBalance(addr []byte) *big.Int
	SetStorage(key []byte, value []byte)
	GetStorage(key []byte) []byte
	Finish(value []byte)
	BlockChainHook() vmcommon.BlockchainHook
	CryptoHook() vmcommon.CryptoHook

	IsInterfaceNil() bool
}

// ContextHandler defines the methods needed to execute system smart contracts
type ContextHandler interface {
	SystemEI

	SetSystemSCContainer(scContainer SystemSCContainer) error
	CreateVMOutput() *vmcommon.VMOutput
	CleanCache()
	SetSCAddress(addr []byte)
	AddCode(addr []byte, code []byte)
	AddTxValueToSmartContract(value *big.Int, scAddress []byte)
}

// MessageSignVerifier is used to verify if message was signed with given public key
type MessageSignVerifier interface {
	Verify(message []byte, signedMessage []byte, pubKey []byte) error
	IsInterfaceNil() bool
}

// ValidatorSettingsHandler defines the functionality which is needed for validators' settings
type ValidatorSettingsHandler interface {
	UnBondPeriod() uint64
	StakeValue() *big.Int
	MinStepValue() *big.Int
	TotalSupply() *big.Int
	NumNodes() uint32
	AuctionEnableNonce() uint64
	StakeEnableNonce() uint64
	IsInterfaceNil() bool
}

// ArgumentsParser defines the functionality to parse transaction data into arguments and code for smart contracts
type ArgumentsParser interface {
	GetArguments() ([][]byte, error)
	GetFunction() (string, error)
	ParseData(data string) error
	IsInterfaceNil() bool
}
