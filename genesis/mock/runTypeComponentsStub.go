package mock

import (
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	testFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	BlockChainHookHandlerFactory  hooks.BlockChainHookHandlerCreator
	TransactionCoordinatorFactory coordinator.TransactionCoordinatorCreator
	SCResultsPreProcessorFactory  preprocess.SmartContractResultPreProcessorCreator
	SCProcessorFactory            scrCommon.SCProcessorCreator
	AccountCreator                state.AccountFactory
	ShardCoordinatorFactory       sharding.ShardCoordinatorFactory
	TxPreProcessorFactory         preprocess.TxPreProcessorCreator
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		BlockChainHookHandlerFactory:  &testFactory.BlockChainHookHandlerFactoryMock{},
		TransactionCoordinatorFactory: &testFactory.TransactionCoordinatorFactoryMock{},
		SCResultsPreProcessorFactory:  &testFactory.SmartContractResultPreProcessorFactoryMock{},
		SCProcessorFactory:            &testFactory.SCProcessorFactoryMock{},
		AccountCreator:                &stateMock.AccountsFactoryStub{},
		ShardCoordinatorFactory:       sharding.NewMultiShardCoordinatorFactory(),
		TxPreProcessorFactory:         preprocess.NewTxPreProcessorCreator(),
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
