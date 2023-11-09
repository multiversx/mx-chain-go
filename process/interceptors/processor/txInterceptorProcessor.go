package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.InterceptorProcessor = (*TxInterceptorProcessor)(nil)
var txLog = logger.GetOrCreate("process/interceptors/processor/txlog")

// TxInterceptorProcessor is the processor used when intercepting transactions
// (smart contract results, receipts, transaction) structs which satisfy TransactionHandler interface.
type TxInterceptorProcessor struct {
	shardedPool process.ShardedPool
	txValidator process.TxValidator
}

// NewTxInterceptorProcessor creates a new TxInterceptorProcessor instance
func NewTxInterceptorProcessor(argument *ArgTxInterceptorProcessor) (*TxInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.ShardedDataCache) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(argument.TxValidator) {
		return nil, process.ErrNilTxValidator
	}

	return &TxInterceptorProcessor{
		shardedPool: argument.ShardedDataCache,
		txValidator: argument.TxValidator,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (txip *TxInterceptorProcessor) Validate(data process.InterceptedData, _ core.PeerID) error {
	interceptedTx, ok := data.(process.InterceptedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return txip.txValidator.CheckTxValidity(interceptedTx)
}

// Save will save the received data into the cacher
func (txip *TxInterceptorProcessor) Save(data process.InterceptedData, peerOriginator core.PeerID, _ string) error {
	interceptedTx, ok := data.(process.InterceptedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err := txip.txValidator.CheckTxWhiteList(data)
	if err != nil {
		log.Trace(
			"TxInterceptorProcessor.Save: not whitelisted cross transactions will not be added in pool",
			"nonce", interceptedTx.Nonce(),
			"sender address", interceptedTx.SenderAddress(),
			"sender shard", interceptedTx.SenderShardId(),
			"receiver shard", interceptedTx.ReceiverShardId(),
		)
		return nil
	}

	txLog.Trace("received transaction", "pid", peerOriginator.Pretty(), "hash", data.Hash())
	cacherIdentifier := process.ShardCacherIdentifier(interceptedTx.SenderShardId(), interceptedTx.ReceiverShardId())
	txip.shardedPool.AddData(
		data.Hash(),
		interceptedTx.Transaction(),
		interceptedTx.Transaction().Size(),
		cacherIdentifier,
	)

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming transactions
func (txip *TxInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("txInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (txip *TxInterceptorProcessor) IsInterfaceNil() bool {
	return txip == nil
}
