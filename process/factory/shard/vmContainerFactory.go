package shard

import (
	arwen "github.com/ElrondNetwork/arwen-wasm-vm/arwen/context"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
)

type vmContainerFactory struct {
	blockChainHookImpl *hooks.BlockChainHookImpl
	cryptoHook         vmcommon.CryptoHook
	blockGasLimit      uint64
	gasSchedule        map[string]uint64
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	blockGasLimit uint64,
	gasSchedule map[string]uint64,
	argBlockChainHook hooks.ArgBlockChainHook,
) (*vmContainerFactory, error) {
	if gasSchedule == nil {
		return nil, process.ErrNilGasSchedule
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(argBlockChainHook)
	if err != nil {
		return nil, err
	}
	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactory{
		blockChainHookImpl: blockChainHookImpl,
		cryptoHook:         cryptoHook,
		blockGasLimit:      blockGasLimit,
		gasSchedule:        gasSchedule,
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
	arwenVM, err := arwen.NewArwenVM(vmf.blockChainHookImpl, vmf.cryptoHook, factory.ArwenVirtualMachine, vmf.blockGasLimit, vmf.gasSchedule)
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
