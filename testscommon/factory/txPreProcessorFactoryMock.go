package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/mock"
)

// TxPreProcessorFactoryMock -
type TxPreProcessorFactoryMock struct {
	CreateTxPreProcessorCalled func(args preprocess.ArgsTransactionPreProcessor) (process.PreProcessor, error)
}

// CreateTxPreProcessor -
func (t *TxPreProcessorFactoryMock) CreateTxPreProcessor(args preprocess.ArgsTransactionPreProcessor) (process.PreProcessor, error) {
	if t.CreateTxPreProcessorCalled != nil {
		return t.CreateTxPreProcessor(args)
	}
	return &mock.PreProcessorMock{}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (t *TxPreProcessorFactoryMock) IsInterfaceNil() bool {
	return t == nil
}
