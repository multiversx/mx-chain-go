package vm

import (
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignVmContainerShardFactory struct {
	vmContainerMetaFactory  VmContainerCreator
	vmContainerShardFactory VmContainerCreator
}

// NewSovereignVmContainerShardFactory creates a new sovereign vm container shard factory
func NewSovereignVmContainerShardFactory(
	vmContainerMeta VmContainerCreator,
	vmContainerShard VmContainerCreator,
) (*sovereignVmContainerShardFactory, error) {
	if check.IfNil(vmContainerMeta) {
		return nil, ErrNilVmContainerMetaCreator
	}
	if check.IfNil(vmContainerShard) {
		return nil, ErrNilVmContainerShardCreator
	}

	return &sovereignVmContainerShardFactory{
		vmContainerMetaFactory:  vmContainerMeta,
		vmContainerShardFactory: vmContainerShard,
	}, nil
}

// CreateVmContainerFactory will create a new instance of sovereign vm container and factory for shard
func (svcsf *sovereignVmContainerShardFactory) CreateVmContainerFactory(args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	vmContainer, vmFactory, err := svcsf.vmContainerShardFactory.CreateVmContainerFactory(args)
	if err != nil {
		return nil, nil, err
	}

	vmContainerMeta, _, err := svcsf.vmContainerMetaFactory.CreateVmContainerFactory(args)
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

	err = vmFactory.BlockChainHookImpl().SetVMContainer(vmContainer)
	if err != nil {
		return nil, nil, err
	}

	return vmContainer, vmFactory, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (svcsf *sovereignVmContainerShardFactory) IsInterfaceNil() bool {
	return svcsf == nil
}
