package shard

import (
	arwen "github.com/ElrondNetwork/arwen-wasm-vm/arwen/context"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
)

type vmContainerFactory struct {
	accounts         state.AccountsAdapter
	addressConverter state.AddressConverter
	blockChainHookImpl *hooks.BlockChainHookImpl
	cryptoHook       vmcommon.CryptoHook
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	accounts state.AccountsAdapter,
	addressConverter state.AddressConverter,
	blockGasLimit uint64,
	gasSchedule map[string]uint64,
	argBlockChainHook hooks.ArgBlockChainHook,
) (*vmContainerFactory, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addressConverter == nil || addressConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if gasSchedule == nil {
		return nil, process.ErrNilGasSchedule
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(argBlockChainHook)
	if err != nil {
		return nil, err
	}
	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactory{
		accounts:         accounts,
		addressConverter: addressConverter,
		blockChainHookImpl *hooks.BlockChainHookImpl,
		cryptoHook:       cryptoHook,
		blockGasLimit:    blockGasLimit,
		gasSchedule:      gasSchedule,
	}, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createIeleVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.IELEVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	currVm, err = vmf.createArwenVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.ArwenVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createIeleVM() (vmcommon.VMExecutionHandler, error) {
	ieleVM := endpoint.NewElrondIeleVM(factory.IELEVirtualMachine, endpoint.ElrondTestnet, vmf.blockChainHookImpl, vmf.cryptoHook)
	return ieleVM, nil
}

func (vmf *vmContainerFactory) createArwenVM() (vmcommon.VMExecutionHandler, error) {
	arwenVM, err := arwen.NewArwenVM(vmf.vmAccountsDB, vmf.cryptoHook, factory.ArwenVirtualMachine, vmf.blockGasLimit, vmf.gasSchedule)
	return arwenVM, err
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	if vmf == nil {
		return true
	}
	return false
}
