package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
)

// ValidatorStatisticsProcessorFactoryStub -
type ValidatorStatisticsProcessorFactoryStub struct {
	CreateValidatorStatisticsProcessorCalled func(args peer.ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error)
}

// NewValidatorStatisticsProcessorFactoryStub -
func NewValidatorStatisticsProcessorFactoryStub() *ValidatorStatisticsProcessorFactoryStub {
	return &ValidatorStatisticsProcessorFactoryStub{}
}

// CreateValidatorStatisticsProcessor -
func (v *ValidatorStatisticsProcessorFactoryStub) CreateValidatorStatisticsProcessor(args peer.ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error) {
	if v.CreateValidatorStatisticsProcessorCalled != nil {
		return v.CreateValidatorStatisticsProcessorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (v *ValidatorStatisticsProcessorFactoryStub) IsInterfaceNil() bool {
	return false
}
