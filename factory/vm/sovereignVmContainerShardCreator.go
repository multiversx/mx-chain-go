package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

type sovereignVmContainerShardFactory struct {
	blockChainHookHandlerCreator hooks.BlockChainHookHandlerCreator
	vmContainerMetaFactory       VmContainerCreator
	vmContainerShardFactory      VmContainerCreator
}

// NewSovereignVmContainerShardFactory creates a new sovereign vm container shard factory
func NewSovereignVmContainerShardFactory(bhhc hooks.BlockChainHookHandlerCreator, vcm VmContainerCreator, vcs VmContainerCreator) (*sovereignVmContainerShardFactory, error) {
	if bhhc == nil {
		return nil, process.ErrNilBlockChainHook
	}

	if vcm == nil {
		return nil, ErrNilVmContainerMetaCreator
	}

	if vcs == nil {
		return nil, ErrNilVmContainerShardCreator
	}

	return &sovereignVmContainerShardFactory{
		blockChainHookHandlerCreator: bhhc,
		vmContainerMetaFactory:       vcm,
		vmContainerShardFactory:      vcs,
	}, nil
}

// CreateVmContainerFactory will create a new instance of sovereign vm container and factory for shard
func (svcsf *sovereignVmContainerShardFactory) CreateVmContainerFactory(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	vmContainer, vmFactory, err := svcsf.vmContainerShardFactory.CreateVmContainerFactory(argsHook, args)
	if err != nil {
		return nil, nil, err
	}

	metaStorage := argsHook.ConfigSCStorage
	metaStorage.DB.FilePath = metaStorage.DB.FilePath + "_meta"
	argsHook.ConfigSCStorage = metaStorage

	vmContainerMeta, _, err := svcsf.vmContainerMetaFactory.CreateVmContainerFactory(argsHook, args)
	if err != nil {
		return nil, nil, err
	}

	vmMeta, err := vmContainerMeta.Get(factory.SystemVirtualMachine)
	if err != nil {
		return nil, nil, err
	}

	err = vmContainer.Add(factory.SystemVirtualMachine, vmMeta)
	if err != nil {
		return nil, nil, err
	}

	return vmContainer, vmFactory, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (svcsf *sovereignVmContainerShardFactory) IsInterfaceNil() bool {
	return svcsf == nil
}
