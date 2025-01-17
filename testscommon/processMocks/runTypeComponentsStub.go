package processMocks

import (
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	SCResultsPreProcessorFactory preprocess.SmartContractResultPreProcessorCreator
	TxPreProcessorFactory        preprocess.TxPreProcessorCreator
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		SCResultsPreProcessorFactory: preprocess.NewSmartContractResultPreProcessorFactory(),
		TxPreProcessorFactory:        preprocess.NewTxPreProcessorCreator(),
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

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
