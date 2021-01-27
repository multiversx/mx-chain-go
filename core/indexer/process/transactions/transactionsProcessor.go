package transactions

import (
	"encoding/hex"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to ESDT contract
	minimumNumberOfSmartContractResults = 2
)

var log = logger.GetOrCreate("indexer/process/transactions")

type txDatabaseProcessor struct {
	txFeeCalculator    process.TransactionFeeCalculator
	txBuilder          *txDBBuilder
	txGrouper          *txGrouper
	txLogsProcessor    process.TransactionLogProcessorDatabase
	saveTxsLogsEnabled bool
}

// NewTransactionsProcessor will create a new instance of transactions database processor
func NewTransactionsProcessor(
	addressPubkeyConverter core.PubkeyConverter,
	txFeeCalculator process.TransactionFeeCalculator,
	isInImportMode bool,
	shardCoordinator sharding.Coordinator,
	saveTxsLogsEnabled bool,
	calculateHash func(object interface{}) ([]byte, error),
	txLogsProcessor process.TransactionLogProcessorDatabase,
) *txDatabaseProcessor {
	txBuilder := newTransactionDBBuilder(addressPubkeyConverter, shardCoordinator, txFeeCalculator)
	txDBGrouper := newTxGrouper(txBuilder, calculateHash, isInImportMode, shardCoordinator.SelfId())

	if isInImportMode {
		log.Warn("the node is in import mode! Cross shard transactions and rewards where destination shard is " +
			"not the current node's shard won't be indexed in Elastic Search")
	}

	return &txDatabaseProcessor{
		txFeeCalculator:    txFeeCalculator,
		txBuilder:          txBuilder,
		txGrouper:          txDBGrouper,
		saveTxsLogsEnabled: saveTxsLogsEnabled,
		txLogsProcessor:    txLogsProcessor,
	}
}

// PrepareTransactionsForDatabase will prepare transactions for database
func (tdp *txDatabaseProcessor) PrepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
) ([]*types.Transaction, []*types.ScResult, []*types.Receipt, map[string]*types.AlteredAccount) {
	alteredAddresses := make(map[string]*types.AlteredAccount)
	transactions, rewardsTxs := tdp.txGrouper.groupNormalTxsAndRewards(body, txPool, header, alteredAddresses)

	transactions = tdp.setTransactionSearchOrder(transactions)

	dbReceipts := tdp.txGrouper.groupReceipts(header, txPool)

	dbSCResults, countScResults := tdp.iterateSCRSAndConvert(txPool, header, transactions)

	tdp.txBuilder.addScrsReceiverToAlteredAccounts(alteredAddresses, dbSCResults)

	tdp.setStatusOfTxsWithSCRS(transactions, countScResults)

	tdp.addTxsLogsIfNeeded(transactions)

	txsSlice := append(convertMapTxsToSlice(transactions), rewardsTxs...)

	return txsSlice, dbSCResults, dbReceipts, alteredAddresses
}

func (tdp *txDatabaseProcessor) setStatusOfTxsWithSCRS(
	transactions map[string]*types.Transaction,
	countScResults map[string]int,
) {
	for hash, nrScResult := range countScResults {
		tx, ok := transactions[hash]
		if !ok {
			continue
		}

		tx.HasSCR = true

		if isRelayedTx(tx) {
			tx.GasUsed = tx.GasLimit
			fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
			tx.Fee = fee.String()

			continue
		}

		if nrScResult < minimumNumberOfSmartContractResults {
			if len(tx.SmartContractResults) > 0 {
				scResultData := tx.SmartContractResults[0].Data
				if isScResultSuccessful(scResultData) {
					// ESDT contract calls generate just one smart contract result
					continue
				}
			}

			tx.Status = transaction.TxStatusFail.String()

			tx.GasUsed = tx.GasLimit
			fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
			tx.Fee = fee.String()
		}
	}
}

func (tdp *txDatabaseProcessor) iterateSCRSAndConvert(
	txPool map[string]data.TransactionHandler,
	header data.HeaderHandler,
	transactions map[string]*types.Transaction,
) ([]*types.ScResult, map[string]int) {
	//we can not iterate smart contract results directly on the miniblocks contained in the block body
	// as some miniblocks might be missing. Example: intra-shard miniblock that holds smart contract results

	scResults := groupSmartContractResults(txPool)

	dbSCResults := make([]*types.ScResult, 0)
	countScResults := make(map[string]int)
	for scHash, scResult := range scResults {
		dbScResult := tdp.txBuilder.convertScResultInDatabaseScr(scHash, scResult, header)
		dbSCResults = append(dbSCResults, dbScResult)

		tx, ok := transactions[string(scResult.OriginalTxHash)]
		if !ok {
			continue
		}

		tx = tdp.addScResultInfoInTx(dbScResult, tx)
		countScResults[string(scResult.OriginalTxHash)]++
		delete(scResults, scHash)

		// append child smart contract results
		scrs := findAllChildScrResults(scHash, scResults)
		for childScHash, sc := range scrs {
			childDBScResult := tdp.txBuilder.convertScResultInDatabaseScr(childScHash, sc, header)

			tx = tdp.addScResultInfoInTx(childDBScResult, tx)
			countScResults[string(scResult.OriginalTxHash)]++
		}
	}

	return dbSCResults, countScResults
}

func (tdp *txDatabaseProcessor) addTxsLogsIfNeeded(txs map[string]*types.Transaction) {
	defer tdp.txLogsProcessor.Clean()

	if !tdp.saveTxsLogsEnabled {
		return
	}

	for hash, tx := range txs {
		txLog, ok := tdp.txLogsProcessor.GetLogFromCache([]byte(hash))
		if !ok {
			continue
		}

		tx.Logs = tdp.prepareTxLog(txLog)
	}
}

func (tdp *txDatabaseProcessor) addScResultInfoInTx(dbScResult *types.ScResult, tx *types.Transaction) *types.Transaction {
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)

	if isSCRForSenderWithRefund(dbScResult, tx) {
		refundValue := stringValueToBigInt(dbScResult.Value)
		gasUsed, fee := tdp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
	}

	return tx
}

func (tdp *txDatabaseProcessor) prepareTxLog(log data.LogHandler) *types.TxLog {
	scAddr := tdp.txBuilder.addressPubkeyConverter.Encode(log.GetAddress())
	events := log.GetLogEvents()

	txLogEvents := make([]types.Event, len(events))
	for i, event := range events {
		txLogEvents[i].Address = hex.EncodeToString(event.GetAddress())
		txLogEvents[i].Data = hex.EncodeToString(event.GetData())
		txLogEvents[i].Identifier = hex.EncodeToString(event.GetIdentifier())

		topics := event.GetTopics()
		txLogEvents[i].Topics = make([]string, len(topics))
		for j, topic := range topics {
			txLogEvents[i].Topics[j] = hex.EncodeToString(topic)
		}
	}

	return &types.TxLog{
		Address: scAddr,
		Events:  txLogEvents,
	}
}

func (tdp *txDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*types.Transaction) map[string]*types.Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}

// SetTxLogsProcessor -
func (tdp *txDatabaseProcessor) SetTxLogsProcessor(txLogProcessor process.TransactionLogProcessorDatabase) {
	tdp.txLogsProcessor = txLogProcessor
}
