package peer

import "github.com/multiversx/mx-chain-go/process"

type validatorStatisticsProcessorFactory struct {
}

// NewValidatorStatisticsProcessorFactory creates a new validator statistics processor factory for normal chain
func NewValidatorStatisticsProcessorFactory() (*validatorStatisticsProcessorFactory, error) {
	return &validatorStatisticsProcessorFactory{}, nil
}

// CreateValidatorStatisticsProcessor creates a new validator statistics processor
func (vsf *validatorStatisticsProcessorFactory) CreateValidatorStatisticsProcessor(args ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error) {
	return NewValidatorStatisticsProcessor(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (vsf *validatorStatisticsProcessorFactory) IsInterfaceNil() bool {
	return vsf == nil
}
