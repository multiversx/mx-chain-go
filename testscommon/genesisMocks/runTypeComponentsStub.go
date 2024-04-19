package genesisMocks

import (
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
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
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	BlockChainHookHandlerFactory  hooks.BlockChainHookHandlerCreator
	TransactionCoordinatorFactory coordinator.TransactionCoordinatorCreator
	SCResultsPreProcessorFactory  preprocess.SmartContractResultPreProcessorCreator
	SCProcessorFactory            scrCommon.SCProcessorCreator
	AccountParser                 genesis.AccountsParser
	AccountCreator                state.AccountFactory
	ShardCoordinatorFactory       sharding.ShardCoordinatorFactory
	TxPreProcessorFactory         preprocess.TxPreProcessorCreator
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

	return &RunTypeComponentsStub{
		BlockChainHookHandlerFactory:  blockChainHookHandlerFactory,
		TransactionCoordinatorFactory: transactionCoordinatorFactory,
		SCResultsPreProcessorFactory:  scResultsPreProcessorCreator,
		SCProcessorFactory:            scProcessorCreator,
		AccountParser:                 &AccountsParserStub{},
		AccountCreator:                accountsCreator,
		ShardCoordinatorFactory:       sharding.NewMultiShardCoordinatorFactory(),
		TxPreProcessorFactory:         preprocess.NewTxPreProcessorCreator(),
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

	return &RunTypeComponentsStub{
		BlockChainHookHandlerFactory:  blockChainHookHandlerFactory,
		TransactionCoordinatorFactory: transactionCoordinatorFactory,
		SCResultsPreProcessorFactory:  scResultsPreProcessorCreator,
		SCProcessorFactory:            scProcessorCreator,
		AccountParser:                 &AccountsParserStub{},
		AccountCreator:                accountsCreator,
		ShardCoordinatorFactory:       sharding.NewSovereignShardCoordinatorFactory(),
		TxPreProcessorFactory:         preprocess.NewSovereignTxPreProcessorCreator(),
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

// ShardCoordinatorCreator -
func (r *RunTypeComponentsStub) ShardCoordinatorCreator() sharding.ShardCoordinatorFactory {
	return r.ShardCoordinatorFactory
}

// TxPreProcessorCreator -
func (r *RunTypeComponentsStub) TxPreProcessorCreator() preprocess.TxPreProcessorCreator {
	return r.TxPreProcessorFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
