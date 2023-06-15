package processing

import (
	"errors"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"time"
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

type TransactionCoordinatorFactory struct {
}

func (tcf *TransactionCoordinatorFactory) CreateTransactionCoordinator(argsTransactionCoordinator coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	return coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
}

type ResolverRequestHandler struct {
}

func (rrh *ResolverRequestHandler) CreateResolverRequestHandler(resolverRequestArgs factory.ResolverRequestArgs) (process.RequestHandler, error) {
	return requestHandlers.NewResolverRequestHandler(
		resolverRequestArgs.RequestersFinder,
		resolverRequestArgs.RequestedItemsHandler,
		resolverRequestArgs.WhiteListHandler,
		common.MaxTxsToRequest,
		resolverRequestArgs.ShardId,
		time.Second,
	)
}
