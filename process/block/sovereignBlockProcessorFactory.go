package block

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	mxErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignBlockProcessorFactory struct {
	shardBlockProcessorFactory BlockProcessorCreator
}

// NewSovereignBlockProcessorFactory creates a new sovereign block processor factory
func NewSovereignBlockProcessorFactory(sbpf BlockProcessorCreator) (*sovereignBlockProcessorFactory, error) {
	if check.IfNil(sbpf) {
		return nil, mxErrors.ErrNilBlockProcessorFactory
	}

	return &sovereignBlockProcessorFactory{
		shardBlockProcessorFactory: sbpf,
	}, nil
}

// CreateBlockProcessor creates a new sovereign block processor for the chain run type sovereign
func (s *sovereignBlockProcessorFactory) CreateBlockProcessor(argumentsBaseProcessor ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	sp, err := s.shardBlockProcessorFactory.CreateBlockProcessor(argumentsBaseProcessor)
	if err != nil {
		return nil, errors.New("could not create shard block processor: " + err.Error())
	}

	shardProc, ok := sp.(*shardProcessor)
	if !ok {
		return nil, mxErrors.ErrWrongTypeAssertion
	}

	scbp, err := NewSovereignChainBlockProcessor(
		shardProc,
		argumentsBaseProcessor.ValidatorStatisticsProcessor,
	)

	return scbp, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sovereignBlockProcessorFactory) IsInterfaceNil() bool {
	return s == nil
}
