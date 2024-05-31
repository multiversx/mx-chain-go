package factory

import (
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

// VMContainerShardFactoryMock -
type VMContainerShardFactoryMock struct {
	CreateVmContainerFactoryShardCalled func(argsHook hooks.ArgBlockChainHook, args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
}

// CreateVmContainerFactory -
func (v *VMContainerShardFactoryMock) CreateVmContainerFactory(argsHook hooks.ArgBlockChainHook, args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	if v.CreateVmContainerFactoryShardCalled != nil {
		return v.CreateVmContainerFactoryShardCalled(argsHook, args)
	}
	return &mock.VMContainerMock{}, &mock.VmMachinesContainerFactoryMock{}, nil
}

// IsInterfaceNil -
func (v *VMContainerShardFactoryMock) IsInterfaceNil() bool {
	return v == nil
}
