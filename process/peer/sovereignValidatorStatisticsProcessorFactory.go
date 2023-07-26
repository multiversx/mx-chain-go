package peer

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignValidatorStatisticsProcessorFactory struct {
	validatorStatisticsProcessorCreator ValidatorStatisticsProcessorCreator
}

// NewSovereignValidatorStatisticsProcessorFactory creates a new validator statistics processor factory for sovereign chain
func NewSovereignValidatorStatisticsProcessorFactory(validatorStatisticsProcessorCreator ValidatorStatisticsProcessorCreator) (*sovereignValidatorStatisticsProcessorFactory, error) {
	if check.IfNil(validatorStatisticsProcessorCreator) {
		return nil, process.ErrNilValidatorStatisticsProcessorCreator
	}
	return &sovereignValidatorStatisticsProcessorFactory{
		validatorStatisticsProcessorCreator: validatorStatisticsProcessorCreator,
	}, nil
}

// CreateValidatorStatisticsProcessor creates a new validator statistics processor
func (vsf *sovereignValidatorStatisticsProcessorFactory) CreateValidatorStatisticsProcessor(args ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error) {
	vsp, err := vsf.validatorStatisticsProcessorCreator.CreateValidatorStatisticsProcessor(args)
	if err != nil {
		return nil, err
	}
	shardVsp, ok := vsp.(*validatorStatistics)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}
	return NewSovereignChainValidatorStatisticsProcessor(shardVsp)
}

// IsInterfaceNil returns true if there is no value under the interface
func (vsf *sovereignValidatorStatisticsProcessorFactory) IsInterfaceNil() bool {
	return vsf == nil
}
