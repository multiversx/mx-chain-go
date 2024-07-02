package factory

import (
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
)

// VMContainerFactoryMock -
type VMContainerFactoryMock struct {
	CreateVmContainerFactoryShardCalled func(args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
}

// CreateVmContainerFactory -
func (v *VMContainerFactoryMock) CreateVmContainerFactory(args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	if v.CreateVmContainerFactoryShardCalled != nil {
		return v.CreateVmContainerFactoryShardCalled(args)
	}
	return &mock.VMContainerMock{}, &mock.VmMachinesContainerFactoryMock{}, nil
}

// IsInterfaceNil -
func (v *VMContainerFactoryMock) IsInterfaceNil() bool {
	return v == nil
}
