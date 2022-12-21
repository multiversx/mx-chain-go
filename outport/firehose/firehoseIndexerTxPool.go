package firehose

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

func getTxPool(transactionsPool *outportcore.Pool) (*txPool, error) {
	if transactionsPool == nil {
		return &txPool{}, nil
	}

	txs, err := getTxs(transactionsPool.Txs)
	if err != nil {
		return nil, err
	}
	scrs, err := getScrs(transactionsPool.Scrs)
	if err != nil {
		return nil, err
	}
	rewards, err := getRewards(transactionsPool.Rewards)
	if err != nil {
		return nil, err
	}
	receipts, err := getReceipts(transactionsPool.Receipts)
	if err != nil {
		return nil, err
	}
	logs, err := getLogs(transactionsPool.Logs)
	if err != nil {
		return nil, err
	}
	invalidTxs, err := getTxs(transactionsPool.Invalid)
	if err != nil {
		return nil, err
	}

	return &txPool{
		transactions:        txs,
		smartContractResult: scrs,
		rewards:             rewards,
		receipts:            receipts,
		invalidTxs:          invalidTxs,
		logs:                logs,
	}, nil
}

func getTxs(txs map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*firehose.TxWithFee, error) {
	ret := make(map[string]*firehose.TxWithFee, len(txs))

	for txHash, txHandler := range txs {
		tx, castOk := txHandler.GetTxHandler().(*transaction.Transaction)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastTransaction, txHash)
		}

		ret[txHash] = &firehose.TxWithFee{
			Transaction: tx,
			FeeInfo:     getFirehoseFreeInfo(txHandler),
		}
	}

	return ret, nil
}

func getScrs(scrs map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*firehose.SCRWithFee, error) {
	ret := make(map[string]*firehose.SCRWithFee, len(scrs))

	for scrHash, txHandler := range scrs {
		tx, castOk := txHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastSCR, scrHash)
		}

		ret[scrHash] = &firehose.SCRWithFee{
			SmartContractResult: tx,
			FeeInfo:             getFirehoseFreeInfo(txHandler),
		}
	}

	return ret, nil
}

func getRewards(rewards map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*rewardTx.RewardTx, error) {
	ret := make(map[string]*rewardTx.RewardTx, len(rewards))
	for hash, txHandler := range rewards {
		tx, castOk := txHandler.GetTxHandler().(*rewardTx.RewardTx)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastRewards, hash)
		}

		ret[hash] = tx
	}
	return ret, nil
}

func getReceipts(receipts map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*receipt.Receipt, error) {
	ret := make(map[string]*receipt.Receipt, len(receipts))
	for hash, receiptHandler := range receipts {
		tx, castOk := receiptHandler.GetTxHandler().(*receipt.Receipt)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastReceipt, hash)
		}

		ret[hash] = tx
	}
	return ret, nil
}

func getLogs(logs []*data.LogData) ([]*transaction.Log, error) {
	ret := make([]*transaction.Log, len(logs))
	for idx, logHandler := range logs {
		eventHandlers := logHandler.GetLogEvents()
		events, err := getEvents(eventHandlers)
		if err != nil {
			return nil, fmt.Errorf("%w, hash: %s", err, logHandler.TxHash)
		}

		ret[idx] = &transaction.Log{
			Address: logHandler.GetAddress(),
			Events:  events,
		}
	}
	return ret, nil
}

func getEvents(eventHandlers []data.EventHandler) ([]*transaction.Event, error) {
	events := make([]*transaction.Event, len(eventHandlers))

	for idx, eventHandler := range eventHandlers {
		event, castOk := eventHandler.(*transaction.Event)
		if !castOk {
			return nil, errCannotCastEvent
		}

		events[idx] = event
	}

	return events, nil
}

func getFirehoseFreeInfo(txHandler data.TransactionHandlerWithGasUsedAndFee) *firehose.FeeInfo {
	return &firehose.FeeInfo{
		GasUsed:        txHandler.GetGasUsed(),
		Fee:            txHandler.GetFee(),
		InitialPaidFee: txHandler.GetInitialPaidFee(),
	}
}
