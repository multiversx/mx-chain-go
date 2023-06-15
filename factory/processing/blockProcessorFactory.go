package processing

import (
	"errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
)

type ShardBlockProcessorFactory struct {
}

func (s ShardBlockProcessorFactory) CreateBlockProcessor(argumentsBaseProcessor factory.ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	argShardProcessor := block.ArgShardProcessor{
		ArgBaseProcessor: argumentsBaseProcessor,
	}

	shardProcessor, err := block.NewShardProcessor(argShardProcessor)
	if err != nil {
		return nil, errors.New("could not create shard block processor: " + err.Error())
	}

	return shardProcessor, nil
}
