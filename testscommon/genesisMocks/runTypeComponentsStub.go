package genesisMocks

import (
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
	"github.com/multiversx/mx-chain-go/process/factory/sovereign"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/vmContext"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	BlockChainHookHandlerFactory              hooks.BlockChainHookHandlerCreator
	TransactionCoordinatorFactory             coordinator.TransactionCoordinatorCreator
	SCResultsPreProcessorFactory              preprocess.SmartContractResultPreProcessorCreator
	SCProcessorFactory                        scrCommon.SCProcessorCreator
	AccountParser                             genesis.AccountsParser
	AccountCreator                            state.AccountFactory
	VMContextCreatorHandler                   systemSmartContracts.VMContextCreatorHandler
	ShardCoordinatorFactory                   sharding.ShardCoordinatorFactory
	TxPreProcessorFactory                     preprocess.TxPreProcessorCreator
	VmContainerShardFactory                   factoryVm.VmContainerCreator
	VmContainerMetaFactory                    factoryVm.VmContainerCreator
	PreProcessorsContainerFactoryCreatorField data.PreProcessorsContainerFactoryCreator
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	blockChainHookHandlerFactory, _ := hooks.NewBlockChainHookFactory()
	transactionCoordinatorFactory, _ := coordinator.NewShardTransactionCoordinatorFactory()
	scResultsPreProcessorCreator, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	scProcessorCreator := processProxy.NewSCProcessProxyFactory()
	accountsCreator, _ := factory.NewAccountCreator(factory.ArgsAccountCreator{
		Hasher:              &hashingMocks.HasherMock{},
		Marshaller:          &marshallerMock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	})
	vmContainerShard := factoryVm.NewVmContainerShardFactory()
	vmContainerMeta, _ := factoryVm.NewVmContainerMetaFactory(systemSmartContracts.NewVMContextCreator())

	return &RunTypeComponentsStub{
		BlockChainHookHandlerFactory:              blockChainHookHandlerFactory,
		TransactionCoordinatorFactory:             transactionCoordinatorFactory,
		SCResultsPreProcessorFactory:              scResultsPreProcessorCreator,
		SCProcessorFactory:                        scProcessorCreator,
		AccountParser:                             &AccountsParserStub{},
		AccountCreator:                            accountsCreator,
		VMContextCreatorHandler:                   systemSmartContracts.NewVMContextCreator(),
		ShardCoordinatorFactory:                   sharding.NewMultiShardCoordinatorFactory(),
		TxPreProcessorFactory:                     preprocess.NewTxPreProcessorCreator(),
		VmContainerShardFactory:                   vmContainerShard,
		VmContainerMetaFactory:                    vmContainerMeta,
		PreProcessorsContainerFactoryCreatorField: shard.NewPreProcessorContainerFactoryCreator(),
	}
}

// NewSovereignRunTypeComponentsStub -
func NewSovereignRunTypeComponentsStub() *RunTypeComponentsStub {
	runTypeComponents := NewRunTypeComponentsStub()

	blockChainHookHandlerFactory, _ := hooks.NewSovereignBlockChainHookFactory(runTypeComponents.BlockChainHookHandlerFactory)
	transactionCoordinatorFactory, _ := coordinator.NewSovereignTransactionCoordinatorFactory(runTypeComponents.TransactionCoordinatorFactory)
	scResultsPreProcessorCreator, _ := preprocess.NewSovereignSmartContractResultPreProcessorFactory(runTypeComponents.SCResultsPreProcessorFactory)
	scProcessorCreator, _ := processorV2.NewSovereignSCProcessFactory(runTypeComponents.SCProcessorFactory)
	accountsCreator, _ := factory.NewSovereignAccountCreator(factory.ArgsSovereignAccountCreator{
		ArgsAccountCreator: factory.ArgsAccountCreator{
			Hasher:              &hashingMocks.HasherMock{},
			Marshaller:          &marshallerMock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		},
		BaseTokenID: "WEGLD-bd4d79",
	})

	oneShardVM := systemSmartContracts.NewOneShardSystemVMEEICreator()
	vmMetaFactory, _ := factoryVm.NewVmContainerMetaFactory(oneShardVM)
	sovVMContainerShardFactory, _ := factoryVm.NewSovereignVmContainerShardFactory(vmMetaFactory, runTypeComponents.VmContainerShardFactory)
	sovVMContainerMeta, _ := factoryVm.NewVmContainerMetaFactory(oneShardVM)

	return &RunTypeComponentsStub{
		BlockChainHookHandlerFactory:              blockChainHookHandlerFactory,
		TransactionCoordinatorFactory:             transactionCoordinatorFactory,
		SCResultsPreProcessorFactory:              scResultsPreProcessorCreator,
		SCProcessorFactory:                        scProcessorCreator,
		AccountParser:                             &AccountsParserStub{},
		AccountCreator:                            accountsCreator,
		VMContextCreatorHandler:                   &vmContext.VMContextCreatorStub{},
		ShardCoordinatorFactory:                   sharding.NewSovereignShardCoordinatorFactory(),
		TxPreProcessorFactory:                     preprocess.NewSovereignTxPreProcessorCreator(),
		VmContainerShardFactory:                   sovVMContainerShardFactory,
		VmContainerMetaFactory:                    sovVMContainerMeta,
		PreProcessorsContainerFactoryCreatorField: sovereign.NewSovereignPreProcessorContainerFactoryCreator(),
	}
}

// BlockChainHookHandlerCreator -
func (r *RunTypeComponentsStub) BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator {
	return r.BlockChainHookHandlerFactory
}

// TransactionCoordinatorCreator -
func (r *RunTypeComponentsStub) TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator {
	return r.TransactionCoordinatorFactory
}

// SCProcessorCreator -
func (r *RunTypeComponentsStub) SCProcessorCreator() scrCommon.SCProcessorCreator {
	return r.SCProcessorFactory
}

// SCResultsPreProcessorCreator -
func (r *RunTypeComponentsStub) SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator {
	return r.SCResultsPreProcessorFactory
}

// AccountsParser -
func (r *RunTypeComponentsStub) AccountsParser() genesis.AccountsParser {
	return r.AccountParser
}

// AccountsCreator -
func (r *RunTypeComponentsStub) AccountsCreator() state.AccountFactory {
	return r.AccountCreator
}

// VMContextCreator -
func (r *RunTypeComponentsStub) VMContextCreator() systemSmartContracts.VMContextCreatorHandler {
	return r.VMContextCreatorHandler
}

// ShardCoordinatorCreator -
func (r *RunTypeComponentsStub) ShardCoordinatorCreator() sharding.ShardCoordinatorFactory {
	return r.ShardCoordinatorFactory
}

// TxPreProcessorCreator -
func (r *RunTypeComponentsStub) TxPreProcessorCreator() preprocess.TxPreProcessorCreator {
	return r.TxPreProcessorFactory
}

// VmContainerShardFactoryCreator -
func (r *RunTypeComponentsStub) VmContainerShardFactoryCreator() factoryVm.VmContainerCreator {
	return r.VmContainerShardFactory
}

// VmContainerMetaFactoryCreator -
func (r *RunTypeComponentsStub) VmContainerMetaFactoryCreator() factoryVm.VmContainerCreator {
	return r.VmContainerMetaFactory
}

// PreProcessorsContainerFactoryCreator -
func (r *RunTypeComponentsStub) PreProcessorsContainerFactoryCreator() data.PreProcessorsContainerFactoryCreator {
	return r.PreProcessorsContainerFactoryCreatorField
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
