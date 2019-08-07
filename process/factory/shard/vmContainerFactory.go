package shard

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
)

type vmContainerFactory struct {
	accounts         state.AccountsAdapter
	addressConverter state.AddressConverter
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	accounts state.AccountsAdapter,
	addressConverter state.AddressConverter,
) (*vmContainerFactory, error) {
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if addressConverter == nil {
		return nil, process.ErrNilAddressConverter
	}

	return &vmContainerFactory{
		accounts:         accounts,
		addressConverter: addressConverter,
	}, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachineContainer, error) {
	container := containers.NewVirtualMachineContainer()

	vmAccountsDB, err := hooks.NewVMAccountsDB(vmf.accounts, vmf.addressConverter)
	if err != nil {
		return nil, err
	}

	cryptoHook := hooks.NewVMCryptoHook()

	vm, err := vmf.createIeleVM(vmAccountsDB, cryptoHook)
	if err != nil {
		return nil, err
	}

	err = container.Add([]byte(factory.IELEVirtualMachine), vm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createIeleVM(
	blockChainHook vmcommon.BlockchainHook,
	crypto vmcommon.CryptoHook,
) (vmcommon.VMExecutionHandler, error) {

	ieleVM := endpoint.NewElrondIeleVM(blockChainHook, crypto, endpoint.ElrondTestnet)
	return ieleVM, nil
}
