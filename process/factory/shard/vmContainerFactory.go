package shard

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
)

type vmContainerFactory struct {
	accounts         state.AccountsAdapter
	addressConverter state.AddressConverter
	vmAccountsDB     *hooks.VMAccountsDB
	cryptoHook       vmcommon.CryptoHook
	shardCoordinator sharding.Coordinator
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	accounts state.AccountsAdapter,
	addressConverter state.AddressConverter,
	shardCoordinator sharding.Coordinator,
) (*vmContainerFactory, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addressConverter == nil || addressConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	vmAccountsDB, err := hooks.NewVMAccountsDB(accounts, addressConverter, shardCoordinator)
	if err != nil {
		return nil, err
	}
	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactory{
		accounts:         accounts,
		addressConverter: addressConverter,
		vmAccountsDB:     vmAccountsDB,
		cryptoHook:       cryptoHook,
		shardCoordinator: shardCoordinator,
	}, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	vm, err := vmf.createIeleVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.IELEVirtualMachine, vm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createIeleVM() (vmcommon.VMExecutionHandler, error) {
	ieleVM := endpoint.NewElrondIeleVM(factory.IELEVirtualMachine, endpoint.ElrondTestnet, vmf.vmAccountsDB, vmf.cryptoHook)
	return ieleVM, nil
}

// VMAccountsDB returns the created vmAccountsDB
func (vmf *vmContainerFactory) VMAccountsDB() *hooks.VMAccountsDB {
	return vmf.vmAccountsDB
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	if vmf == nil {
		return true
	}
	return false
}
