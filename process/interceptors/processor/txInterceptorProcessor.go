package processor

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TxInterceptorProcessor is the interceptor used for transaction-like structs that satisfy
type TxInterceptorProcessor struct {
	shardedDataCache dataRetriever.ShardedDataCacherNotifier
}

// NewTxInterceptorProcessor creates a new TxInterceptorProcessor instance
func NewTxInterceptorProcessor(argument *ArgTxInterceptorProcessor) (*TxInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if argument.ShardedDataCache == nil {
		return nil, process.ErrNilDataPoolHolder
	}

	return &TxInterceptorProcessor{
		shardedDataCache: argument.ShardedDataCache,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (txip TxInterceptorProcessor) Validate(data process.InterceptedData) error {
	//TODO implement tx checking logic here (will be done in a future PR)
	return nil
}

// Save will save the received data into the cacher
func (txip TxInterceptorProcessor) Save(data process.InterceptedData) error {
	interceptedTxHandler, ok := data.(InterceptedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	cacherIdentifier := process.ShardCacherIdentifier(interceptedTxHandler.SndShard(), interceptedTxHandler.RcvShard())
	txip.shardedDataCache.AddData(
		interceptedTxHandler.Hash(),
		interceptedTxHandler.UnderlyingTransaction(),
		cacherIdentifier,
	)
	return nil
}
