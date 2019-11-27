package processor

import (
	"encoding/hex"
	"fmt"

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

	debugTx(interceptedTx, "validate")
	errTxValidation := txip.txValidator.CheckTxValidity(interceptedTx)
	if errTxValidation != nil {
		log.Debug("Intercepted transaction", "error", errTxValidation)
		return errTxValidation
	}
	log.Trace("Intercepted transaction", "valid", true)
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
		data.Hash(),
		interceptedTx.Transaction(),
		cacherIdentifier,
	)

	debugTx(interceptedTx, "save")
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txip *TxInterceptorProcessor) IsInterfaceNil() bool {
	if txip == nil {
		return true
	}
	return false
}

func debugTx(tx InterceptedTransactionHandler, verb string) {
	if tx.Transaction() == nil {
		log.Debug("Intercepted transaction", "nil transaction")
	}

	sender := "nil"
	if tx.SenderAddress() != nil {
		if tx.SenderAddress().Bytes() != nil {
			senderAddress := hex.EncodeToString(tx.SenderAddress().Bytes())
			sender = fmt.Sprintf("S%d - %s (%d)", tx.SenderShardId(), senderAddress, tx.Nonce())
		}
	}

	receiver := "nil"
	if tx.Transaction().GetRecvAddress() != nil {
		receiverAddress := hex.EncodeToString(tx.Transaction().GetRecvAddress())
		receiver = fmt.Sprintf("R%d - %s", tx.ReceiverShardId(), receiverAddress)
	}

	data := fmt.Sprintf("Data: %s", tx.Transaction().GetData())
	log.Trace("Intercepted transaction "+verb, "sender", sender)
	log.Trace("Intercepted transaction "+verb, "receiver", receiver)
	log.Trace("Intercepted transaction "+verb, "data", data)
}
