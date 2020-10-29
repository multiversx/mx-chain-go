package indexer

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
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
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const (
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to ESDT contract
	minimumNumberOfSmartContractResults = 2
)

type txDatabaseProcessor struct {
	*commonProcessor
	txLogsProcessor process.TransactionLogProcessorDatabase
	hasher          hashing.Hasher
	marshalizer     marshal.Marshalizer
	isInImportMode  bool
}

func newTxDatabaseProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
	feeConfig *config.FeeSettings,
	isInImportMode bool,
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
		},
		txLogsProcessor: disabled.NewNilTxLogsProcessor(),
		isInImportMode:  isInImportMode,
	}
}

func (tdp *txDatabaseProcessor) prepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
) ([]*Transaction, map[string]struct{}) {
	transactions, rewardsTxs, alteredAddresses := tdp.groupNormalTxsAndRewards(body, txPool, header, selfShardID)
	receipts := groupReceipts(body, txPool)
	scResults := groupSmartContractResults(body, txPool)
	tdp.addScrsReceiverToAlteredAccounts(body, alteredAddresses, scResults, selfShardID)

	transactions = tdp.setTransactionSearchOrder(transactions)
	for _, rec := range receipts {
		tx, ok := transactions[string(rec.TxHash)]
		if !ok {
			continue
		}

		tx.GasUsed = getGasUsedFromReceipt(rec, tx)
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
		for childScHash, sc := range scrs {
			tx = tdp.addScResultInfoInTx(childScHash, sc, tx)
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

	return append(convertMapTxsToSlice(transactions), rewardsTxs...), alteredAddresses
}

func (tdp *txDatabaseProcessor) addScrsReceiverToAlteredAccounts(
	body *block.Body,
	alteredAddress map[string]struct{},
	scrs map[string]*smartContractResult.SmartContractResult,
	selfShardID uint32,
) {
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}
		if mb.ReceiverShardID != selfShardID {
			continue
		}

		for _, txHash := range mb.TxHashes {
			scr, ok := scrs[string(txHash)]
			if !ok {
				log.Warn("internal error: missing already computed scr when altering address",
					"hash", txHash)
			}

			encodedReceiverAddress := tdp.addressPubkeyConverter.Encode(scr.RcvAddr)
			alteredAddress[encodedReceiverAddress] = struct{}{}
		}
	}
}

func getGasUsedFromReceipt(rec *receipt.Receipt, tx *Transaction) uint64 {
	if rec.Data != nil && string(rec.Data) == processTransaction.RefundGasMessage {
		// in this gas receipt contains the refunded value
		gasUsed := big.NewInt(0).SetUint64(tx.GasPrice)
		gasUsed.Mul(gasUsed, big.NewInt(0).SetUint64(tx.GasLimit))
		gasUsed.Sub(gasUsed, rec.Value)
		gasUsed.Div(gasUsed, big.NewInt(0).SetUint64(tx.GasPrice))

		return gasUsed.Uint64()
	}

	gasUsed := big.NewInt(0)
	gasUsed = gasUsed.Div(rec.Value, big.NewInt(0).SetUint64(tx.GasPrice))

	return gasUsed.Uint64()
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
	map[string]struct{},
) {
	alteredAddresses := make(map[string]struct{})
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
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, mbTxStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, false)
				if tdp.shouldIndex(selfShardID, mb.ReceiverShardID) {
					transactions[hash] = dbTx
				}
			}
		case block.InvalidBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, transaction.TxStatusInvalid.String())
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, false)
				transactions[hash] = dbTx
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				dbTx := tdp.commonProcessor.buildRewardTransaction(rtx, []byte(hash), mbHash, mb, header, mbTxStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, true)
				if tdp.shouldIndex(selfShardID, mb.ReceiverShardID) {
					rewardsTxs = append(rewardsTxs, dbTx)
				}
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

func (tdp *txDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*Transaction) map[string]*Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}

func addToAlteredAddresses(
	tx *Transaction,
	alteredAddresses map[string]struct{},
	miniBlock *block.MiniBlock,
	selfShardID uint32,
	isRewardTx bool,
) {
	if selfShardID == miniBlock.SenderShardID && !isRewardTx {
		alteredAddresses[tx.Sender] = struct{}{}
	}
	if selfShardID == miniBlock.ReceiverShardID || miniBlock.ReceiverShardID == core.AllShardId {
		alteredAddresses[tx.Receiver] = struct{}{}
	}
}

func groupSmartContractResults(body *block.Body, txPool map[string]data.TransactionHandler) map[string]*smartContractResult.SmartContractResult {
	scResults := make(map[string]*smartContractResult.SmartContractResult, 0)
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}

		for _, txHash := range mb.TxHashes {
			txHandler, ok := txPool[string(txHash)]
			if !ok {
				log.Warn("internal error: missing scr", "hash", txHash)
				continue
			}

			scResult, ok := txHandler.(*smartContractResult.SmartContractResult)
			if !ok {
				log.Warn("internal error: wrong type assertion for scr", "hash", txHash)
				continue
			}

			scResults[string(txHash)] = scResult
		}
	}

	return scResults
}

func groupReceipts(body *block.Body, txPool map[string]data.TransactionHandler) []*receipt.Receipt {
	receipts := make([]*receipt.Receipt, 0)
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.ReceiptBlock {
			continue
		}

		for _, txHash := range mb.TxHashes {
			txHandler, ok := txPool[string(txHash)]
			if !ok {
				log.Warn("internal error: missing receipt", "hash", txHash)
				continue
			}

			rtx, ok := txHandler.(*receipt.Receipt)
			if !ok {
				log.Warn("internal error: wrong type assertion for receipt", "hash", txHash)
				continue
			}

			receipts = append(receipts, rtx)
		}
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
