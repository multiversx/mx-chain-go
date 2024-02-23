package block

import (
	"errors"

	"github.com/multiversx/mx-chain-go/process"
)

type shardBlockProcessorFactory struct {
}

// NewShardBlockProcessorFactory creates a new shard block processor factory
func NewShardBlockProcessorFactory() (*shardBlockProcessorFactory, error) {
	return &shardBlockProcessorFactory{}, nil
}

// CreateBlockProcessor creates a new shard block processor for the chain run type normal
func (s *shardBlockProcessorFactory) CreateBlockProcessor(argumentsBaseProcessor ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	argShardProcessor := ArgShardProcessor{
		ArgBaseProcessor: argumentsBaseProcessor,
	}

	sp, err := NewShardProcessor(argShardProcessor)
	if err != nil {
		return nil, errors.New("could not create shard block processor: " + err.Error())
	}

	return sp, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardBlockProcessorFactory) IsInterfaceNil() bool {
	return s == nil
}
