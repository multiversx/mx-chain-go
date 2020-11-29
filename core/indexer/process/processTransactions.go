package process

import (
	"encoding/hex"
	"strconv"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
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
	txLogsProcessor  process.TransactionLogProcessorDatabase
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	isInImportMode   bool
	shardCoordinator sharding.Coordinator
}

func newTxDatabaseProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
	feeConfig *config.FeeSettings,
	isInImportMode bool,
	shardCoordinator sharding.Coordinator,
) *txDatabaseProcessor {
	// this should never return error because is tested when economics file is created
	minGasLimit, _ := strconv.ParseUint(feeConfig.MinGasLimit, 10, 64)
	gasPerDataByte, _ := strconv.ParseUint(feeConfig.GasPerDataByte, 10, 64)

	return &txDatabaseProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
		commonProcessor: &commonProcessor{
			addressPubkeyConverter:   addressPubkeyConverter,
			validatorPubkeyConverter: validatorPubkeyConverter,
			minGasLimit:              minGasLimit,
			gasPerDataByte:           gasPerDataByte,
			esdtProc:                 newEsdtTransactionHandler(),
		},
		txLogsProcessor:  disabled.NewNilTxLogsProcessor(),
		isInImportMode:   isInImportMode,
		shardCoordinator: shardCoordinator,
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

		tx, ok := transactions[string(rec.TxHash)]
		if !ok {
			continue
		}

		tx.GasUsed = getGasUsedFromReceipt(rec, tx)
	}

	dbSCResults := make([]*types.ScResult, 0)
	countScResults := make(map[string]int)
	for scHash, scResult := range scResults {
		dbScResult := tdp.commonProcessor.convertScResultInDatabaseScr(scHash, scResult)
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
			childDBScResult := tdp.commonProcessor.convertScResultInDatabaseScr(childScHash, sc)

			tx = tdp.addScResultInfoInTx(childDBScResult, tx)
			countScResults[string(scResult.OriginalTxHash)]++
		}
	}

	tdp.addScrsReceiverToAlteredAccounts(alteredAddresses, dbSCResults)

	for hash, nrScResult := range countScResults {
		if nrScResult < minimumNumberOfSmartContractResults {
			if len(transactions[hash].SmartContractResults) > 0 {
				scResultData := transactions[hash].SmartContractResults[0].Data
				if isScResultSuccessful(scResultData) {
					// ESDT contract calls generate just one smart contract result
					continue
				}
			}

			if strings.Contains(string(transactions[hash].Data), "relayedTx") {
				continue
			}

			transactions[hash].Status = transaction.TxStatusFail.String()
		}
	}

	// TODO for the moment do not save logs in database
	// uncomment this when transaction logs need to be saved in database
	//for hash, tx := range transactions {
	//	txLog, ok := tdp.txLogsProcessor.GetLogFromCache([]byte(hash))
	//	if !ok {
	//		continue
	//	}
	//
	//	tx.Log = tdp.prepareTxLog(txLog)
	//}

	tdp.txLogsProcessor.Clean()

	txsSlice := append(convertMapTxsToSlice(transactions), rewardsTxs...)

	return txsSlice, dbSCResults, dbReceipts, alteredAddresses
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
			IsESDTSender:    false,
			IsESDTOperation: scr.EsdtTokenIdentifier != "" && scr.EsdtValue != "",
			TokenIdentifier: scr.EsdtTokenIdentifier,
		}
	}
}

func (tdp *txDatabaseProcessor) addScResultInfoInTx(dbScResults *types.ScResult, tx *types.Transaction) *types.Transaction {
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResults)

	if isSCRForSenderWithGasUsed(dbScResults, tx) {
		gasUsed := tx.GasLimit - dbScResults.GasLimit
		tx.GasUsed = gasUsed
	}

	return tx
}

func (tdp *txDatabaseProcessor) prepareTxLog(log data.LogHandler) types.TxLog {
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

	return types.TxLog{
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
