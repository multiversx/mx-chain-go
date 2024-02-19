package factory

import (
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

type VMContainerMetaFactoryMock struct {
	CreateVmContainerFactoryMetaCalled func(argsHook hooks.ArgBlockChainHook, args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
}

func (v *VMContainerMetaFactoryMock) CreateVmContainerFactory(argsHook hooks.ArgBlockChainHook, args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	if v.CreateVmContainerFactoryMetaCalled != nil {
		return v.CreateVmContainerFactoryMetaCalled(argsHook, args)
	}
	return &mock.VMContainerMock{}, &mock.VmMachinesContainerFactoryMock{}, nil
}

func (v *VMContainerMetaFactoryMock) IsInterfaceNil() bool {
	return v == nil
}
