package vm

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type ExecuteArguments struct {
	Sender   []byte
	Value    *big.Int
	Function string
	Args     []*big.Int
}

// SystemSmartContract interface defines the function a system smart contract should have
type SystemSmartContract interface {
	Execute(args *ExecuteArguments) vmcommon.ReturnCode
	ValueOf(key interface{}) interface{}
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
	AddMultiple(keys [][]byte, scs []SystemSmartContract) error
	Replace(key []byte, val SystemSmartContract) error
	Remove(key []byte)
	Len() int
	Keys() [][]byte
	IsInterfaceNil() bool
}

// SystemEI defines the environment interface system smart contract can use
type SystemEI interface {
	Transfer(destination []byte, sender []byte, value *big.Int, input []byte) error
	BlockHash(number int64) []byte
	GetVMInput() vmcommon.VMInput
	GetBalance(addr []byte) *big.Int
	SetStorage(addr []byte, key []byte, value []byte)
	GetStorage(addr []byte, key []byte) []byte
	GetSCAddress() []byte
	SelfDestruct(addr []byte, beneficiary []byte)

	CreateVMOutput() *vmcommon.VMOutput
	CleanCache()
	SetContractCallInput(input *vmcommon.ContractCallInput)

	IsInterfaceNil() bool
}

// PeerChangesEI defines the environment interface system smart contract can use to write peer changes
type PeerChangesEI interface {
	GetPeerState()
	SetPeerState()

	CleanCache()
	CreatePeerChangesOutput()
	IsInterfaceNil() bool
}
