package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/state/syncer"
)

type sovereignVmContainerShardFactory struct {
	blockChainHookHandlerCreator hooks.BlockChainHookHandlerCreator
}

func NewSovereignVmContainerShardFactory(bhhc hooks.BlockChainHookHandlerCreator) (*sovereignVmContainerShardFactory, error) {
	return &sovereignVmContainerShardFactory{
		blockChainHookHandlerCreator: bhhc,
	}, nil
}

func (svcsf *sovereignVmContainerShardFactory) CreateVmContainerFactoryShard(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	blockChainHookImpl, err := svcsf.blockChainHookHandlerCreator.CreateBlockChainHookHandler(argsHook)
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

	sovereignFactory, err := NewSovereignVmContainerMetaFactory(svcsf.blockChainHookHandlerCreator)
	if err != nil {
		return nil, nil, err
	}

	metaStorage := argsHook.ConfigSCStorage
	metaStorage.DB.FilePath = metaStorage.DB.FilePath + "_meta"
	argsHook.ConfigSCStorage = metaStorage
	argsHook.Counter = counters.NewDisabledCounter()
	argsHook.MissingTrieNodesNotifier = syncer.NewMissingTrieNodesNotifier()

	vmContainerMeta, _, err := sovereignFactory.CreateVmContainerFactoryMeta(argsHook, args)
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

func (svcsf *sovereignVmContainerShardFactory) IsInterfaceNil() bool {
	return svcsf == nil
}
