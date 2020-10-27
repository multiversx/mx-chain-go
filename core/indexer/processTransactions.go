package indexer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const (
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to ESDT contract
	minimumNumberOfSmartContractResults = 2
)

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
	shardCoordinator sharding.Coordinator,
	isInImportMode bool,
) *txDatabaseProcessor {
	return &txDatabaseProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
		commonProcessor: &commonProcessor{
			addressPubkeyConverter:   addressPubkeyConverter,
			validatorPubkeyConverter: validatorPubkeyConverter,
		},
		txLogsProcessor:  disabled.NewNilTxLogsProcessor(),
		shardCoordinator: shardCoordinator,
		isInImportMode:   isInImportMode,
	}
}

func (tdp *txDatabaseProcessor) prepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
) []*Transaction {
	transactions, rewardsTxs := tdp.groupNormalTxsAndRewards(body, txPool, header, selfShardID)
	receipts := groupReceipts(txPool)
	scResults := groupSmartContractResults(txPool)

	transactions = tdp.setTransactionSearchOrder(transactions, selfShardID)
	for _, rec := range receipts {
		tx, ok := transactions[string(rec.TxHash)]
		if !ok {
			continue
		}

		gasUsed := big.NewInt(0).SetUint64(tx.GasPrice)
		gasUsed.Mul(gasUsed, big.NewInt(0).SetUint64(tx.GasLimit))
		gasUsed.Sub(gasUsed, rec.Value)
		gasUsed.Div(gasUsed, big.NewInt(0).SetUint64(tx.GasPrice))

		tx.GasUsed = gasUsed.Uint64()
	}

	countScResults := make(map[string]int)
	for scHash, scResult := range scResults {
		tx, ok := transactions[string(scResult.OriginalTxHash)]
		if !ok {
			continue
		}

		tx = tdp.addScResultInfoInTx(scHash, scResult, tx)
		countScResults[string(scResult.OriginalTxHash)]++
		delete(scResults, scHash)

		// append child smart contract results
		scrs := findAllChildScrResults(scHash, scResults)
		for scHash, sc := range scrs {
			tx = tdp.addScResultInfoInTx(scHash, sc, tx)
			countScResults[string(scResult.OriginalTxHash)]++
		}
	}

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

	return append(convertMapTxsToSlice(transactions), rewardsTxs...)
}

func isScResultSuccessful(scResultData []byte) bool {
	okReturnDataNewVersion := []byte("@" + hex.EncodeToString([]byte(vmcommon.Ok.String())))
	okReturnDataOldVersion := []byte("@" + vmcommon.Ok.String()) // backwards compatible
	return bytes.Contains(scResultData, okReturnDataNewVersion) || bytes.Contains(scResultData, okReturnDataOldVersion)
}

func findAllChildScrResults(hash string, scrs map[string]*smartContractResult.SmartContractResult) map[string]*smartContractResult.SmartContractResult {
	scrResults := make(map[string]*smartContractResult.SmartContractResult, 0)
	for scrHash, scr := range scrs {
		if string(scr.OriginalTxHash) == hash {
			scrResults[scrHash] = scr
			delete(scrs, scrHash)
		}
	}

	return scrResults
}

func (tdp *txDatabaseProcessor) addScResultInfoInTx(scHash string, scr *smartContractResult.SmartContractResult, tx *Transaction) *Transaction {
	dbScResult := tdp.commonProcessor.convertScResultInDatabaseScr(scHash, scr)
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)

	if isSCRForSenderWithGasUsed(dbScResult, tx) {
		gasUsed := tx.GasLimit - scr.GasLimit
		tx.GasUsed = gasUsed
	}

	return tx
}

func isSCRForSenderWithGasUsed(dbScResult ScResult, tx *Transaction) bool {
	isForSender := dbScResult.Receiver == tx.Sender
	isWithGasLimit := dbScResult.GasLimit != 0
	isFromCurrentTx := dbScResult.PreTxHash == tx.Hash

	return isFromCurrentTx && isForSender && isWithGasLimit
}

func (tdp *txDatabaseProcessor) prepareTxLog(log data.LogHandler) TxLog {
	scAddr := tdp.addressPubkeyConverter.Encode(log.GetAddress())
	events := log.GetLogEvents()

	txLogEvents := make([]Event, len(events))
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
	selfShardID uint32,
) (
	map[string]*Transaction,
	[]*Transaction,
) {
	transactions := make(map[string]*Transaction)
	rewardsTxs := make([]*Transaction, 0)

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
				if tdp.shouldIndex(tx) {
					dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, mbTxStatus)
					transactions[hash] = dbTx
				}

				delete(txPool, hash)
			}
		case block.InvalidBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, transaction.TxStatusInvalid.String())
				transactions[hash] = dbTx
				delete(txPool, hash)
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				if tdp.shouldIndex(rtx) {
					dbTx := tdp.commonProcessor.buildRewardTransaction(rtx, []byte(hash), mbHash, mb, header, mbTxStatus)
					rewardsTxs = append(rewardsTxs, dbTx)
				}
				delete(txPool, hash)
			}
		default:
			continue
		}
	}

	return transactions, rewardsTxs
}

func (tdp *txDatabaseProcessor) shouldIndex(txHandler data.TransactionHandler) bool {
	if !tdp.isInImportMode {
		return true
	}

	destShardId := tdp.shardCoordinator.ComputeId(txHandler.GetRcvAddr())
	return destShardId == tdp.shardCoordinator.SelfId()
}

func (tdp *txDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*Transaction, shardId uint32) map[string]*Transaction {
	currentOrder := 0
	shardIdentifier := tdp.createShardIdentifier(shardId)

	for _, tx := range transactions {
		stringOrder := fmt.Sprintf("%d%d", shardIdentifier, currentOrder)
		order, err := strconv.ParseUint(stringOrder, 10, 32)
		if err != nil {
			order = 0
			log.Debug("processTransactions.setTransactionSearchOrder", "could not set uint32 search order", err.Error())
		}
		tx.SearchOrder = uint32(order)
		currentOrder++
	}

	return transactions
}

func (tdp *txDatabaseProcessor) createShardIdentifier(shardId uint32) uint32 {
	shardIdentifier := shardId + 2
	if shardId == core.MetachainShardId {
		shardIdentifier = 1
	}

	return shardIdentifier
}

func groupSmartContractResults(txPool map[string]data.TransactionHandler) map[string]*smartContractResult.SmartContractResult {
	scResults := make(map[string]*smartContractResult.SmartContractResult, 0)
	for hash, tx := range txPool {
		scResult, ok := tx.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		scResults[hash] = scResult
	}

	return scResults
}

func groupReceipts(txPool map[string]data.TransactionHandler) []*receipt.Receipt {
	receipts := make([]*receipt.Receipt, 0)
	for hash, tx := range txPool {
		rec, ok := tx.(*receipt.Receipt)
		if !ok {
			continue
		}

		receipts = append(receipts, rec)
		delete(txPool, hash)
	}

	return receipts
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
