package firehose

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/firehose"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type txPool struct {
	transactions        map[string]*firehose.TxInfo
	smartContractResult map[string]*firehose.SCRInfo
	rewards             map[string]*firehose.RewardInfo
	receipts            map[string]*receipt.Receipt
	invalidTxs          map[string]*firehose.TxInfo
	logs                map[string]*transaction.Log
}

func getTxPool(transactionsPool *outportcore.Pool) (*txPool, error) {
	if transactionsPool == nil {
		return nil, errNilTxPool
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

func getFirehoseFeeInfo(txHandler data.TransactionHandlerWithGasUsedAndFee) *firehose.FeeInfo {
	return &firehose.FeeInfo{
		GasUsed:        txHandler.GetGasUsed(),
		Fee:            txHandler.GetFee(),
		InitialPaidFee: txHandler.GetInitialPaidFee(),
	}
}

func getTxs(txs map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*firehose.TxInfo, error) {
	ret := make(map[string]*firehose.TxInfo, len(txs))

	for txHash, txHandler := range txs {
		tx, castOk := txHandler.GetTxHandler().(*transaction.Transaction)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastTransaction, txHash)
		}

		ret[txHash] = &firehose.TxInfo{
			Transaction:    tx,
			FeeInfo:        getFirehoseFeeInfo(txHandler),
			ExecutionOrder: uint32(txHandler.GetExecutionOrder()),
		}
	}

	return ret, nil
}

func getScrs(scrs map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*firehose.SCRInfo, error) {
	ret := make(map[string]*firehose.SCRInfo, len(scrs))

	for scrHash, txHandler := range scrs {
		tx, castOk := txHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastSCR, scrHash)
		}

		ret[scrHash] = &firehose.SCRInfo{
			SmartContractResult: tx,
			FeeInfo:             getFirehoseFeeInfo(txHandler),
			ExecutionOrder:      uint32(txHandler.GetExecutionOrder()),
		}
	}

	return ret, nil
}

func getRewards(rewards map[string]data.TransactionHandlerWithGasUsedAndFee) (map[string]*firehose.RewardInfo, error) {
	ret := make(map[string]*firehose.RewardInfo, len(rewards))

	for hash, txHandler := range rewards {
		reward, castOk := txHandler.GetTxHandler().(*rewardTx.RewardTx)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastReward, hash)
		}

		ret[hash] = &firehose.RewardInfo{
			Reward:         reward,
			ExecutionOrder: uint32(txHandler.GetExecutionOrder()),
		}
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

func getLogs(logs []*data.LogData) (map[string]*transaction.Log, error) {
	ret := make(map[string]*transaction.Log, len(logs))

	for _, logHandler := range logs {
		eventHandlers := logHandler.GetLogEvents()
		events, err := getEvents(eventHandlers)
		if err != nil {
			return nil, fmt.Errorf("%w, hash: %s", err, logHandler.TxHash)
		}

		ret[logHandler.TxHash] = &transaction.Log{
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
