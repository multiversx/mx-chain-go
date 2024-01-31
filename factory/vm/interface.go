package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

// VmContainerMetaCreator can create vm container and factory for metachain
type VmContainerMetaCreator interface {
	CreateVmContainerFactoryMeta(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
	IsInterfaceNil() bool
}

// VmContainerShardCreator can create vm container and factory for shard
type VmContainerShardCreator interface {
	CreateVmContainerFactoryShard(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
	IsInterfaceNil() bool
}
