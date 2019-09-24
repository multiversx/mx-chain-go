package processor

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TxInterceptorProcessor is the processor used when intercepting transactions
// (smart contract results, receipts, transaction) structs which satisfy TransactionHandler interface.
type TxInterceptorProcessor struct {
	shardedDataCache dataRetriever.ShardedDataCacherNotifier
	txValidator      process.TxValidator
}

// NewTxInterceptorProcessor creates a new TxInterceptorProcessor instance
func NewTxInterceptorProcessor(argument *ArgTxInterceptorProcessor) (*TxInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.IfNil(argument.ShardedDataCache) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(argument.TxValidator) {
		return nil, process.ErrNilTxValidator
	}

	return &TxInterceptorProcessor{
		shardedDataCache: argument.ShardedDataCache,
		txValidator:      argument.TxValidator,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (txip *TxInterceptorProcessor) Validate(data process.InterceptedData) error {
	interceptedTx, ok := data.(InterceptedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	//TODO change the IsTxValidForProcessing to output error
	isTxValid := txip.txValidator.IsTxValidForProcessing(interceptedTx)
	if !isTxValid {
		return process.ErrTxNotValid
	}

	return nil
}

// Save will save the received data into the cacher
func (txip *TxInterceptorProcessor) Save(data process.InterceptedData) error {
	interceptedTx, ok := data.(InterceptedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	cacherIdentifier := process.ShardCacherIdentifier(interceptedTx.SenderShardId(), interceptedTx.ReceiverShardId())
	txip.shardedDataCache.AddData(
		interceptedTx.Hash(),
		interceptedTx.Transaction(),
		cacherIdentifier,
	)
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txip *TxInterceptorProcessor) IsInterfaceNil() bool {
	if txip == nil {
		return true
	}
	return false
}
