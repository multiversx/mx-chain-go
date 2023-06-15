package factory

import (
	"errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

// SovereignBlockChainHookFactory - factory for sovereign run
type SovereignBlockChainHookFactory struct {
}

func (bhf *SovereignBlockChainHookFactory) CreateBlockChainHook(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := hooks.NewBlockChainHookImpl(args)
	return hooks.NewSovereignBlockChainHook(bh)
}

type SovereignBlockProcessorFactory struct {
}

func (s SovereignBlockProcessorFactory) CreateBlockProcessor(argumentsBaseProcessor factory.ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	argShardProcessor := block.ArgShardProcessor{
		ArgBaseProcessor: argumentsBaseProcessor,
	}

	shardProcessor, err := block.NewShardProcessor(argShardProcessor)
	if err != nil {
		return nil, errors.New("could not create shard block processor: " + err.Error())
	}

	scbp, err := block.NewSovereignChainBlockProcessor(
		shardProcessor,
		argumentsBaseProcessor.ValidatorStatisticsProcessor,
	)

	return scbp, nil
}
