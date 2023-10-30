package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// SmartContractResultPreProcessorFactoryMock -
type SmartContractResultPreProcessorFactoryMock struct {
	CreateSmartContractResultPreProcessorCalled func(args preprocess.SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error)
}

// CreateSmartContractResultPreProcessor -
func (s *SmartContractResultPreProcessorFactoryMock) CreateSmartContractResultPreProcessor(args preprocess.SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error) {
	if s.CreateSmartContractResultPreProcessorCalled != nil {
		return s.CreateSmartContractResultPreProcessorCalled(args)
	}
	return &testscommon.PreProcessorMock{}, nil
}

// IsInterfaceNil -
func (s *SmartContractResultPreProcessorFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
