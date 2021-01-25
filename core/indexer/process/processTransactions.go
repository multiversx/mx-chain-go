package process

import (
	"encoding/hex"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to ESDT contract
	minimumNumberOfSmartContractResults = 2
)

var log = logger.GetOrCreate("indexer/process")

type txDatabaseProcessor struct {
	*commonProcessor
	txLogsProcessor    process.TransactionLogProcessorDatabase
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
	isInImportMode     bool
	saveTxsLogsEnabled bool
}

func newTxDatabaseProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
	txFeeCalculator process.TransactionFeeCalculator,
	isInImportMode bool,
	shardCoordinator sharding.Coordinator,
	saveTxsLogsEnabled bool,
) *txDatabaseProcessor {
	return &txDatabaseProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
		commonProcessor: &commonProcessor{
			addressPubkeyConverter:   addressPubkeyConverter,
			validatorPubkeyConverter: validatorPubkeyConverter,
			txFeeCalculator:          txFeeCalculator,
			shardCoordinator:         shardCoordinator,
			esdtProc:                 newEsdtTransactionHandler(),
		},
		txLogsProcessor:    disabled.NewNilTxLogsProcessor(),
		isInImportMode:     isInImportMode,
		saveTxsLogsEnabled: saveTxsLogsEnabled,
	}
}

// PrepareTransactionsForDatabase will prepare transactions for database
func (tdp *txDatabaseProcessor) prepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
) ([]*types.Transaction, []*types.ScResult, []*types.Receipt, map[string]*accounts.AlteredAccount) {
	transactions, rewardsTxs, alteredAddresses := tdp.groupNormalTxsAndRewards(body, txPool, header, selfShardID)
	//we can not iterate smart contract results directly on the miniblocks contained in the block body
	// as some miniblocks might be missing. Example: intra-shard miniblock that holds smart contract results
	receipts := groupReceipts(txPool)
	scResults := groupSmartContractResults(txPool)

	dbReceipts := make([]*types.Receipt, 0)
	transactions = tdp.setTransactionSearchOrder(transactions)
	for recHash, rec := range receipts {
		dbReceipts = append(dbReceipts, tdp.commonProcessor.convertReceiptInDatabaseReceipt(recHash, rec, header))
	}

	dbSCResults := make([]*types.ScResult, 0)
	countScResults := make(map[string]int)
	for scHash, scResult := range scResults {
		dbScResult := tdp.commonProcessor.convertScResultInDatabaseScr(scHash, scResult, header)
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
			childDBScResult := tdp.commonProcessor.convertScResultInDatabaseScr(childScHash, sc, header)

			tx = tdp.addScResultInfoInTx(childDBScResult, tx)
			countScResults[string(scResult.OriginalTxHash)]++
		}
	}

	tdp.addScrsReceiverToAlteredAccounts(alteredAddresses, dbSCResults)

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

	tdp.addTxsLogsIfNeeded(transactions)

	txsSlice := append(convertMapTxsToSlice(transactions), rewardsTxs...)

	return txsSlice, dbSCResults, dbReceipts, alteredAddresses
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

func (tdp *txDatabaseProcessor) addScrsReceiverToAlteredAccounts(
	alteredAddress map[string]*accounts.AlteredAccount,
	scrs []*types.ScResult,
) {
	for _, scr := range scrs {
		receiverAddr, _ := tdp.addressPubkeyConverter.Decode(scr.Receiver)
		shardID := tdp.shardCoordinator.ComputeId(receiverAddr)
		if shardID != tdp.shardCoordinator.SelfId() {
			continue
		}

		egldBalanceNotChanged := scr.Value == "" || scr.Value == "0"
		esdtBalanceNotChanged := scr.EsdtValue == "" || scr.EsdtValue == "0"
		if egldBalanceNotChanged && esdtBalanceNotChanged {
			// the smart contract results that dont't alter the balance of the receiver address should be ignored
			continue
		}
		encodedReceiverAddress := scr.Receiver
		alteredAddress[encodedReceiverAddress] = &accounts.AlteredAccount{
			IsESDTOperation: scr.EsdtTokenIdentifier != "" && scr.EsdtValue != "",
			TokenIdentifier: scr.EsdtTokenIdentifier,
		}
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
	scAddr := tdp.addressPubkeyConverter.Encode(log.GetAddress())
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

func (tdp *txDatabaseProcessor) groupNormalTxsAndRewards(
	body *block.Body,
	txPool map[string]data.TransactionHandler,
	header data.HeaderHandler,
	selfShardID uint32,
) (
	map[string]*types.Transaction,
	[]*types.Transaction,
	map[string]*accounts.AlteredAccount,
) {
	alteredAddresses := make(map[string]*accounts.AlteredAccount)
	transactions := make(map[string]*types.Transaction)
	rewardsTxs := make([]*types.Transaction, 0)

	for _, mb := range body.MiniBlocks {
		mbHash, err := core.CalculateHash(tdp.marshalizer, tdp.hasher, mb)
		if err != nil {
			continue
		}

		mbTxStatus := transaction.TxStatusPending.String()
		if selfShardID == mb.ReceiverShardID {
			mbTxStatus = transaction.TxStatusSuccess.String()
		}

		switch mb.Type {
		case block.TxBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, mbTxStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, false)
				if tdp.shouldIndex(selfShardID, mb.ReceiverShardID) {
					transactions[hash] = dbTx
				}
				delete(txPool, hash)
			}
		case block.InvalidBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, transaction.TxStatusInvalid.String())
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, false)

				dbTx.GasUsed = dbTx.GasLimit
				fee := tdp.commonProcessor.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, dbTx.GasUsed)
				dbTx.Fee = fee.String()

				transactions[hash] = dbTx
				delete(txPool, hash)
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				dbTx := tdp.commonProcessor.buildRewardTransaction(rtx, []byte(hash), mbHash, mb, header, mbTxStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, true)
				if tdp.shouldIndex(selfShardID, mb.ReceiverShardID) {
					rewardsTxs = append(rewardsTxs, dbTx)
				}
				delete(txPool, hash)
			}
		default:
			continue
		}
	}

	return transactions, rewardsTxs, alteredAddresses
}

func (tdp *txDatabaseProcessor) shouldIndex(selfShardID uint32, destinationShardID uint32) bool {
	if !tdp.isInImportMode {
		return true
	}

	return selfShardID == destinationShardID
}

func (tdp *txDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*types.Transaction) map[string]*types.Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}
