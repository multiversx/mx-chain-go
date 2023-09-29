package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/mock"
)

// SmartContractResultPreProcessorFactoryStub -
type SmartContractResultPreProcessorFactoryStub struct {
}

// CreateSmartContractResultPreProcessor -
func (s *SmartContractResultPreProcessorFactoryStub) CreateSmartContractResultPreProcessor(_ preprocess.SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error) {
	return &mock.PreProcessorMock{}, nil
}

// IsInterfaceNil -
func (s *SmartContractResultPreProcessorFactoryStub) IsInterfaceNil() bool {
	return s == nil
}
