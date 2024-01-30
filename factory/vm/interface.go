package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

type VmContainerMetaCreator interface {
	CreateVmContainerFactoryMeta(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
	IsInterfaceNil() bool
}

type VmContainerShardCreator interface {
	CreateVmContainerFactoryShard(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
	IsInterfaceNil() bool
}
