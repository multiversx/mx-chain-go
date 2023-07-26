package peer

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

// ValidatorStatisticsProcessorCreator is an interface for creating validator statistics processors
type ValidatorStatisticsProcessorCreator interface {
	CreateValidatorStatisticsProcessor(args ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error)
	IsInterfaceNil() bool
}

type sovereignValidatorStatisticsProcessorFactory struct {
	validatorStatisticsProcessorCreator ValidatorStatisticsProcessorCreator
}

// NewSovereignValidatorStatisticsProcessorFactory creates a new validator statistics processor factory for sovereign chain
func NewSovereignValidatorStatisticsProcessorFactory(validatorStatisticsProcessorCreator ValidatorStatisticsProcessorCreator) (*sovereignValidatorStatisticsProcessorFactory, error) {
	if check.IfNil(validatorStatisticsProcessorCreator) {
		return nil, process.ErrNilValidatorStatisticsProcessorCreator
	}
	return &sovereignValidatorStatisticsProcessorFactory{}, nil
}

// CreateValidatorStatisticsProcessor creates a new validator statistics processor
func (vsf *sovereignValidatorStatisticsProcessorFactory) CreateValidatorStatisticsProcessor(args ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error) {
	return NewValidatorStatisticsProcessor(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (vsf *sovereignValidatorStatisticsProcessorFactory) IsInterfaceNil() bool {
	return vsf == nil
}
