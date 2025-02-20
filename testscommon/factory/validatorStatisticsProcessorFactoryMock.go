package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
)

// ValidatorStatisticsProcessorFactoryMock -
type ValidatorStatisticsProcessorFactoryMock struct {
	CreateValidatorStatisticsProcessorCalled func(args peer.ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error)
}

// CreateValidatorStatisticsProcessor -
func (v *ValidatorStatisticsProcessorFactoryMock) CreateValidatorStatisticsProcessor(args peer.ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error) {
	if v.CreateValidatorStatisticsProcessorCalled != nil {
		return v.CreateValidatorStatisticsProcessorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (v *ValidatorStatisticsProcessorFactoryMock) IsInterfaceNil() bool {
	return v == nil
}
