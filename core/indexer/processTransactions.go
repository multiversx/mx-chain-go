package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type txDatabaseProcessor struct {
	*commonProcessor
	txLogsProcessor process.TransactionLogProcessorDatabase
	hasher          hashing.Hasher
	marshalizer     marshal.Marshalizer
}

func newTxDatabaseProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter state.PubkeyConverter,
	validatorPubkeyConverter state.PubkeyConverter,
) *txDatabaseProcessor {
	return &txDatabaseProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
		commonProcessor: &commonProcessor{
			addressPubkeyConverter:   addressPubkeyConverter,
			validatorPubkeyConverter: validatorPubkeyConverter,
		},
	}
}

func (tdp *txDatabaseProcessor) prepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardId uint32,
) []*Transaction {
	transactions, rewardsTxs := tdp.groupNormalTxsAndRewards(body, txPool, header, selfShardId)
	receipts, scResults := groupReceiptAndSmartContractResults(body, txPool)

	for _, rec := range receipts {
		tx, ok := transactions[string(rec.TxHash)]
		if !ok {
			continue
		}
		tx.ReceiptValue = rec.Value.String()
	}

	for _, scResult := range scResults {
		tx, ok := transactions[string(scResult.OriginalTxHash)]
		if !ok {
			continue
		}
		tx.SmartContractResults = append(tx.SmartContractResults, tdp.commonProcessor.convertScResultInDatabaseScr(scResult))
	}

	for hash, tx := range transactions {
		log, ok := tdp.txLogsProcessor.GetLogFromRAM([]byte(hash))
		if !ok {
			continue
		}

		tx.Log = prepareTxLog(log)
	}

	return append(convertMapTxsToSlice(transactions), rewardsTxs...)
}

func prepareTxLog(log data.LogHandler) TxLog {
	scAddr := string(log.GetAddress())
	events := log.GetLogEvents()

	txLogEvents := make([]*Event, len(events))
	for i, event := range events {
		txLogEvents[i].Address = string(event.GetAddress())
		txLogEvents[i].Data = string(event.GetData())
		txLogEvents[i].Identifier = string(event.GetIdentifier())

		topics := event.GetTopics()
		txLogEvents[i].Topics = make([]string, len(topics))
		for j, topic := range topics {
			txLogEvents[i].Topics[j] = string(topic)
		}
	}

	return TxLog{
		Address: scAddr,
		Events:  txLogEvents,
	}
}

func convertMapTxsToSlice(txs map[string]*Transaction) []*Transaction {
	transactions := make([]*Transaction, len(txs))
	i := 0
	for _, tx := range txs {
		transactions[i] = tx
		i++
	}
	return transactions
}

func (tdp *txDatabaseProcessor) groupNormalTxsAndRewards(
	body *block.Body,
	txPool map[string]data.TransactionHandler,
	header data.HeaderHandler,
	selfShardId uint32,
) (
	map[string]*Transaction,
	[]*Transaction,
) {
	transactions := make(map[string]*Transaction)
	rewardsTxs := make([]*Transaction, 0)

	blockHash, err := core.CalculateHash(tdp.marshalizer, tdp.hasher, body)
	if err != nil {
		return transactions, rewardsTxs
	}

	for _, mb := range body.MiniBlocks {
		mbHash, err := core.CalculateHash(tdp.marshalizer, tdp.hasher, mb)
		if err != nil {
			continue
		}

		mbTxStatus := "Pending"
		if selfShardId == mb.ReceiverShardID {
			mbTxStatus = "Success"
		}

		switch mb.Type {
		case block.TxBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, blockHash, mb, header, mbTxStatus)
				transactions[hash] = dbTx
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				dbTx := tdp.commonProcessor.buildRewardTransaction(rtx, []byte(hash), mbHash, blockHash, mb, header, mbTxStatus)
				rewardsTxs = append(rewardsTxs, dbTx)
			}
		default:
			continue
		}
	}

	return transactions, rewardsTxs
}

func groupReceiptAndSmartContractResults(body *block.Body, txPool map[string]data.TransactionHandler) (
	map[string]*receipt.Receipt,
	map[string]*smartContractResult.SmartContractResult,
) {
	receipts := make(map[string]*receipt.Receipt)
	scResults := make(map[string]*smartContractResult.SmartContractResult)
	for _, mb := range body.MiniBlocks {
		switch mb.Type {
		case block.ReceiptBlock:
			tRecs := getTransactionsReceipts(txPool, mb.TxHashes)
			for hash, tRec := range tRecs {
				receipts[hash] = tRec
			}
		case block.SmartContractResultBlock:
			scsRes := getSmartContractResults(txPool, mb.TxHashes)
			for hash, scr := range scsRes {
				scResults[hash] = scr
			}
		default:
			continue
		}
	}

	return receipts, scResults
}

func getTransactionsReceipts(
	txPool map[string]data.TransactionHandler,
	txHashes [][]byte,
) map[string]*receipt.Receipt {
	receiptsMap := make(map[string]*receipt.Receipt)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		rec, ok := txHandler.(*receipt.Receipt)
		if !ok {
			continue
		}

		receiptsMap[string(txHash)] = rec
	}

	return receiptsMap
}

func getSmartContractResults(
	txPool map[string]data.TransactionHandler,
	txHashes [][]byte,
) map[string]*smartContractResult.SmartContractResult {
	scrsResults := make(map[string]*smartContractResult.SmartContractResult)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		scr, ok := txHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		scrsResults[string(txHash)] = scr
	}
	return scrsResults
}

func getTransactions(txPool map[string]data.TransactionHandler,
	txHashes [][]byte,
) map[string]*transaction.Transaction {
	transactions := make(map[string]*transaction.Transaction)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		tx, ok := txHandler.(*transaction.Transaction)
		if !ok {
			continue
		}
		transactions[string(txHash)] = tx
	}
	return transactions
}

func getRewardsTransaction(txPool map[string]data.TransactionHandler,
	txHashes [][]byte,
) map[string]*rewardTx.RewardTx {
	rewardsTxs := make(map[string]*rewardTx.RewardTx)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		reward, ok := txHandler.(*rewardTx.RewardTx)
		if !ok {
			continue
		}
		rewardsTxs[string(txHash)] = reward
	}
	return rewardsTxs
}
