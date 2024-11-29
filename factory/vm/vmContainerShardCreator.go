package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
)

type vmContainerShardFactory struct {
}

// NewVmContainerShardFactory creates a new vm container shard factory
func NewVmContainerShardFactory() *vmContainerShardFactory {
	return &vmContainerShardFactory{}
}

// CreateVmContainerFactory will create a new vm container and factoy for shard
func (vcsf *vmContainerShardFactory) CreateVmContainerFactory(args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	argsNewVmFactory := shard.ArgVMContainerFactory{
		BlockChainHook:      args.BlockChainHook,
		BuiltInFunctions:    args.BuiltInFunctions,
		Config:              args.Config,
		BlockGasLimit:       args.BlockGasLimit,
		GasSchedule:         args.GasSchedule,
		EpochNotifier:       args.EpochNotifier,
		EnableEpochsHandler: args.EnableEpochsHandler,
		WasmVMChangeLocker:  args.WasmVMChangeLocker,
		ESDTTransferParser:  args.ESDTTransferParser,
		Hasher:              args.Hasher,
		PubKeyConverter:     args.PubkeyConv,
	}
	vmFactory, err := shard.NewVMContainerFactory(argsNewVmFactory)
	if err != nil {
		return nil, nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	return vmContainer, vmFactory, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (vcsf *vmContainerShardFactory) IsInterfaceNil() bool {
	return vcsf == nil
}
