package metachain

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	systemVMFactory "github.com/ElrondNetwork/elrond-go/vm/factory"
	systemVMProcess "github.com/ElrondNetwork/elrond-go/vm/process"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/ElrondNetwork/elrond-vm-common"
)

type vmContainerFactory struct {
	blockChainHookImpl *hooks.BlockChainHookImpl
	cryptoHook         vmcommon.CryptoHook
	systemContracts    vm.SystemSCContainer
	economics          *economics.EconomicsData
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	argBlockChainHook hooks.ArgBlockChainHook,
	economics *economics.EconomicsData,
) (*vmContainerFactory, error) {
	if economics == nil {
		return nil, process.ErrNilEconomicsData
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(argBlockChainHook)
	if err != nil {
		return nil, err
	}
	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactory{
		blockChainHookImpl: blockChainHookImpl,
		cryptoHook:         cryptoHook,
		economics:          economics,
	}, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createSystemVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.SystemVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createSystemVM() (vmcommon.VMExecutionHandler, error) {
	systemEI, err := systemSmartContracts.NewVMContext(vmf.blockChainHookImpl, vmf.cryptoHook)
	if err != nil {
		return nil, err
	}

	scFactory, err := systemVMFactory.NewSystemSCFactory(systemEI, vmf.economics)
	if err != nil {
		return nil, err
	}

	vmf.systemContracts, err = scFactory.Create()
	if err != nil {
		return nil, err
	}

	systemVM, err := systemVMProcess.NewSystemVM(systemEI, vmf.systemContracts, factory.SystemVirtualMachine)
	if err != nil {
		return nil, err
	}

	return systemVM, nil
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// SystemSmartContractContainer return the created system smart contracts
func (vmf *vmContainerFactory) SystemSmartContractContainer() vm.SystemSCContainer {
	return vmf.systemContracts
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	if vmf == nil {
		return true
	}
	return false
}
