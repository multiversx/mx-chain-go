package factory

import (
	"errors"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/factory"
	processDisabled "github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"time"
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

type SovereignTransactionCoordinatorFactory struct {
}

func (tcf *SovereignTransactionCoordinatorFactory) CreateTransactionCoordinator(argsTransactionCoordinator coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	transactionCoordinator, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	if err != nil {
		return nil, err
	}

	return coordinator.NewSovereignChainTransactionCoordinator(transactionCoordinator)
}

type SovereignResolverRequestHandler struct {
}

func (rrh *SovereignResolverRequestHandler) CreateResolverRequestHandler(resolverRequestArgs factory.ResolverRequestArgs) (process.RequestHandler, error) {
	requestHandler, err := requestHandlers.NewResolverRequestHandler(
		resolverRequestArgs.RequestersFinder,
		resolverRequestArgs.RequestedItemsHandler,
		resolverRequestArgs.WhiteListHandler,
		common.MaxTxsToRequest,
		resolverRequestArgs.ShardId,
		time.Second,
	)
	if err != nil {
		return nil, err
	}

	return requestHandlers.NewSovereignResolverRequestHandler(requestHandler)
}

type SovereignScheduledTxsExecutionFactory struct {
}

func (stxef *SovereignScheduledTxsExecutionFactory) CreateScheduledTxsExecutionHandler(_ factory.ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error) {
	return &processDisabled.ScheduledTxsExecutionHandler{}, nil
}
