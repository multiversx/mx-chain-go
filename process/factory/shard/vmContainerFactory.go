package shard

import (
	vm "github.com/ElrondNetwork/elrond-go/core/libLocator"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
	"github.com/ElrondNetwork/hera/evmc/bindings/go/evmc"
)

type vmContainerFactory struct {
	accounts         state.AccountsAdapter
	addressConverter state.AddressConverter
	vmAccountsDB     *hooks.VMAccountsDB
	cryptoHook       vmcommon.CryptoHook
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	accounts state.AccountsAdapter,
	addressConverter state.AddressConverter,
) (*vmContainerFactory, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addressConverter == nil || addressConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}

	vmAccountsDB, err := hooks.NewVMAccountsDB(accounts, addressConverter)
	if err != nil {
		return nil, err
	}
	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactory{
		accounts:         accounts,
		addressConverter: addressConverter,
		vmAccountsDB:     vmAccountsDB,
		cryptoHook:       cryptoHook,
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

	currVm, err = vmf.createHeraBinaryenVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.HeraWBinaryenVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	currVm, err = vmf.createHeraWABTVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.HeraWABTVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	currVm, err = vmf.createHeraWAVMVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.HeraWAVMVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createIeleVM() (vmcommon.VMExecutionHandler, error) {
	ieleVM := endpoint.NewElrondIeleVM(factory.IELEVirtualMachine, endpoint.ElrondTestnet, vmf.vmAccountsDB, vmf.cryptoHook)
	return ieleVM, nil
}

func (vmf *vmContainerFactory) createHeraBinaryenVM() (vmcommon.VMExecutionHandler, error) {
	config := vm.WASMLibLocation() + ",engine=binaryen"
	wasmVM, err := evmc.NewWASMInstance(config, vmf.vmAccountsDB, vmf.cryptoHook, factory.HeraWABTVirtualMachine)
	return wasmVM, err
}

func (vmf *vmContainerFactory) createHeraWABTVM() (vmcommon.VMExecutionHandler, error) {
	config := vm.WASMLibLocation() + ",engine=wabt"
	wasmVM, err := evmc.NewWASMInstance(config, vmf.vmAccountsDB, vmf.cryptoHook, factory.HeraWABTVirtualMachine)
	return wasmVM, err
}

func (vmf *vmContainerFactory) createHeraWAVMVM() (vmcommon.VMExecutionHandler, error) {
	config := vm.WASMLibLocation() + ",engine=wavm"
	wasmVM, err := evmc.NewWASMInstance(config, vmf.vmAccountsDB, vmf.cryptoHook, factory.HeraWABTVirtualMachine)
	return wasmVM, err
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
