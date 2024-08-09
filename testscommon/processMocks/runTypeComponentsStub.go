package processMocks

import (
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	SCResultsPreProcessorFactory preprocess.SmartContractResultPreProcessorCreator
	TxPreProcessorFactory        preprocess.TxPreProcessorCreator
	RewardsTxPreProcFactoryField preprocess.RewardsTxPreProcFactory
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	scrPreProcFactory, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	return &RunTypeComponentsStub{
		SCResultsPreProcessorFactory: scrPreProcFactory,
		TxPreProcessorFactory:        preprocess.NewTxPreProcessorCreator(),
		RewardsTxPreProcFactoryField: preprocess.NewRewardsTxPreProcFactory(),
	}
}

// SCResultsPreProcessorCreator -
func (r *RunTypeComponentsStub) SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator {
	return r.SCResultsPreProcessorFactory
}

// TxPreProcessorCreator -
func (r *RunTypeComponentsStub) TxPreProcessorCreator() preprocess.TxPreProcessorCreator {
	return r.TxPreProcessorFactory
}

// RewardsTxPreProcFactory -
func (r *RunTypeComponentsStub) RewardsTxPreProcFactory() preprocess.RewardsTxPreProcFactory {
	return r.RewardsTxPreProcFactoryField
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
