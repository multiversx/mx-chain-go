package factory

import (
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

type VmContainerMetaFactoryMock struct {
	CreateVmContainerFactoryMetaCalled func(argsHook hooks.ArgBlockChainHook, args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
}

func (v *VmContainerMetaFactoryMock) CreateVmContainerFactoryMeta(argsHook hooks.ArgBlockChainHook, args factoryVm.ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	if v.CreateVmContainerFactoryMetaCalled != nil {
		return v.CreateVmContainerFactoryMeta(argsHook, args)
	}
	return &mock.VMContainerMock{}, &mock.VmMachinesContainerFactoryMock{}, nil
}

func (v *VmContainerMetaFactoryMock) IsInterfaceNil() bool {
	return v == nil
}
