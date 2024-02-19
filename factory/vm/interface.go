package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

// VmContainerCreator can create vm container and factory for metachain/shard/sovereign
type VmContainerCreator interface {
	CreateVmContainerFactory(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error)
	IsInterfaceNil() bool
}
