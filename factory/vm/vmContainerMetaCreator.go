package vm

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

type vmContainerMetaFactory struct {
	blockChainHookHandlerCreator hooks.BlockChainHookHandlerCreator
}

// NewVmContainerMetaFactory creates a new vm container meta factory
func NewVmContainerMetaFactory(bhhc hooks.BlockChainHookHandlerCreator) (*vmContainerMetaFactory, error) {
	return &vmContainerMetaFactory{
		blockChainHookHandlerCreator: bhhc,
	}, nil
}

// CreateVmContainerFactoryMeta will create a new vm container and factory for metachain
func (vcmf *vmContainerMetaFactory) CreateVmContainerFactoryMeta(argsHook hooks.ArgBlockChainHook, args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	blockChainHookImpl, err := vcmf.blockChainHookHandlerCreator.CreateBlockChainHookHandler(argsHook)
	if err != nil {
		return nil, nil, err
	}

	argsNewVmFactory := metachain.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          args.PubkeyConv,
		Economics:           args.Economics,
		MessageSignVerifier: args.MessageSignVerifier,
		GasSchedule:         args.GasSchedule,
		NodesConfigProvider: args.NodesConfigProvider,
		Hasher:              args.Hasher,
		Marshalizer:         args.Marshalizer,
		SystemSCConfig:      args.SystemSCConfig,
		ValidatorAccountsDB: args.ValidatorAccountsDB,
		UserAccountsDB:      args.UserAccountsDB,
		ChanceComputer:      args.ChanceComputer,
		ShardCoordinator:    args.ShardCoordinator,
		EnableEpochsHandler: args.EnableEpochsHandler,
	}
	vmFactory, err := metachain.NewVMContainerFactory(argsNewVmFactory)
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
func (vcmf *vmContainerMetaFactory) IsInterfaceNil() bool {
	return vcmf == nil
}
