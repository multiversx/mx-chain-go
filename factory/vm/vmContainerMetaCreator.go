package vm

import (
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

type vmContainerMetaFactory struct {
	vmContextCreatorHandler systemSmartContracts.VMContextCreatorHandler
}

// NewVmContainerMetaFactory creates a new vm container meta factory
func NewVmContainerMetaFactory(
	vmContextCreator systemSmartContracts.VMContextCreatorHandler,
) (*vmContainerMetaFactory, error) {
	if check.IfNil(vmContextCreator) {
		return nil, errors.ErrNilVMContextCreator
	}

	return &vmContainerMetaFactory{
		vmContextCreatorHandler: vmContextCreator,
	}, nil
}

// CreateVmContainerFactory will create a new vm container and factory for metachain
func (vcmf *vmContainerMetaFactory) CreateVmContainerFactory(args ArgsVmContainerFactory) (process.VirtualMachinesContainer, process.VirtualMachinesContainerFactory, error) {
	argsNewVmFactory := metachain.ArgsNewVMContainerFactory{
		BlockChainHook:          args.BlockChainHook,
		PubkeyConv:              args.PubkeyConv,
		Economics:               args.Economics,
		MessageSignVerifier:     args.MessageSignVerifier,
		GasSchedule:             args.GasSchedule,
		NodesConfigProvider:     args.NodesConfigProvider,
		Hasher:                  args.Hasher,
		Marshalizer:             args.Marshalizer,
		SystemSCConfig:          args.SystemSCConfig,
		ValidatorAccountsDB:     args.ValidatorAccountsDB,
		UserAccountsDB:          args.UserAccountsDB,
		ChanceComputer:          args.ChanceComputer,
		ShardCoordinator:        args.ShardCoordinator,
		EnableEpochsHandler:     args.EnableEpochsHandler,
		NodesCoordinator:        args.NodesCoordinator,
		VMContextCreatorHandler: vcmf.vmContextCreatorHandler,
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
