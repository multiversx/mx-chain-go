package process

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
)

func convertMapTxsToSlice(txs map[string]*types.Transaction) []*types.Transaction {
	transactions := make([]*types.Transaction, len(txs))
	i := 0
	for _, tx := range txs {
		transactions[i] = tx
		i++
	}
	return transactions
}

func addToAlteredAddresses(
	tx *types.Transaction,
	alteredAddresses map[string]*accounts.AlteredAccount,
	miniBlock *block.MiniBlock,
	selfShardID uint32,
	isRewardTx bool,
) {
	isESDTTx := tx.EsdtTokenIdentifier != "" && tx.EsdtValue != ""

	if selfShardID == miniBlock.SenderShardID && !isRewardTx {
		alteredAddresses[tx.Sender] = &accounts.AlteredAccount{
			IsSender:        true,
			IsESDTOperation: isESDTTx,
			TokenIdentifier: tx.EsdtTokenIdentifier,
		}
	}

	if tx.Status == transaction.TxStatusInvalid.String() {
		// ignore receiver if we have an invalid transaction
		return
	}

	if selfShardID == miniBlock.ReceiverShardID || miniBlock.ReceiverShardID == core.AllShardId {
		alteredAddresses[tx.Receiver] = &accounts.AlteredAccount{
			IsSender:        false,
			IsESDTOperation: isESDTTx,
			TokenIdentifier: tx.EsdtTokenIdentifier,
		}
	}
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

func groupReceipts(txPool map[string]data.TransactionHandler) map[string]*receipt.Receipt {
	receipts := make(map[string]*receipt.Receipt, 0)
	for hash, tx := range txPool {
		rec, ok := tx.(*receipt.Receipt)
		if !ok {
			continue
		}

		receipts[hash] = rec
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

func getGasUsedFromReceipt(rec *receipt.Receipt, tx *types.Transaction) uint64 {
	if rec.Data != nil && string(rec.Data) == processTransaction.RefundGasMessage {
		// this receipts contains the gas that needs to be refunded
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
	scrResults := make(map[string]*smartContractResult.SmartContractResult)
	for scrHash, scr := range scrs {
		if string(scr.OriginalTxHash) == hash {
			scrResults[scrHash] = scr
			delete(scrs, scrHash)
		}
	}

	return scrResults
}

func computeTxGasUsedField(dbScResult *types.ScResult, tx *types.Transaction) uint64 {
	fee := big.NewInt(0).SetUint64(tx.GasPrice)
	fee.Mul(fee, big.NewInt(0).SetUint64(tx.GasLimit))

	refundValue, ok := big.NewInt(0).SetString(dbScResult.Value, 10)
	if !ok {
		log.Warn("indexer.computeTxGasUsedField() cannot cast value from string to big.Int")
		return fee.Uint64()
	}

	diff := fee.Sub(fee, refundValue)
	gasUsedBig := diff.Div(diff, big.NewInt(0).SetUint64(tx.GasPrice))

	return gasUsedBig.Uint64()
}

func isSCRForSenderWithRefund(dbScResult *types.ScResult, tx *types.Transaction) bool {
	isForSender := dbScResult.Receiver == tx.Sender
	isRightNonce := dbScResult.Nonce == tx.Nonce+1
	isFromCurrentTx := dbScResult.PreTxHash == tx.Hash
	isScrDataOk := isDataOk(dbScResult.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isDataOk(data []byte) bool {
	okEncoded := hex.EncodeToString([]byte("ok"))
	dataFieldStr := "@" + okEncoded

	return strings.HasPrefix(string(data), dataFieldStr)
}

func stringValueToBigInt(strValue string) *big.Int {
	value, ok := big.NewInt(0).SetString(strValue, 10)
	if !ok {
		return big.NewInt(0)
	}

	return value
}
