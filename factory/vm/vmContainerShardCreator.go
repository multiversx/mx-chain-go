package vm

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

type vmContainerShardFactory struct {
	blockChainHookHandlerCreator hooks.BlockChainHookHandlerCreator
}

// NewVmContainerShardFactory creates a new vm container shard factory
func NewVmContainerShardFactory(bhhc hooks.BlockChainHookHandlerCreator) (*vmContainerShardFactory, error) {
	if check.IfNil(bhhc) {
		return nil, process.ErrNilBlockChainHook
	}

	return &vmContainerShardFactory{
		blockChainHookHandlerCreator: bhhc,
	}, nil
}

// CreateVmContainerFactory will create a new vm container and factoy for shard
func (vcsf *vmContainerShardFactory) CreateVmContainerFactory(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	blockChainHookImpl, err := vcsf.blockChainHookHandlerCreator.CreateBlockChainHookHandler(argsHook)
	if err != nil {
		return nil, nil, err
	}

	argsNewVmFactory := shard.ArgVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		BuiltInFunctions:    args.BuiltInFunctions,
		Config:              args.Config,
		BlockGasLimit:       args.BlockGasLimit,
		GasSchedule:         args.GasSchedule,
		EpochNotifier:       args.EpochNotifier,
		EnableEpochsHandler: args.EnableEpochsHandler,
		WasmVMChangeLocker:  args.WasmVMChangeLocker,
		ESDTTransferParser:  args.ESDTTransferParser,
		Hasher:              args.Hasher,
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
