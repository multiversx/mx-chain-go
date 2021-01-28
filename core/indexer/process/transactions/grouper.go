package transactions

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

type txGrouper struct {
	selfShardID    uint32
	txBuilder      *txDBBuilder
	calculateHash  func(object interface{}) ([]byte, error)
	isInImportMode bool
}

func newTxGrouper(
	txBuilder *txDBBuilder,
	calculateHash func(object interface{}) ([]byte, error),
	isInImportMode bool,
	selfShardID uint32,
) *txGrouper {
	return &txGrouper{
		txBuilder:      txBuilder,
		calculateHash:  calculateHash,
		selfShardID:    selfShardID,
		isInImportMode: isInImportMode,
	}
}

func (tg *txGrouper) groupNormalTxsAndRewards(
	body *block.Body,
	txPool map[string]data.TransactionHandler,
	header data.HeaderHandler,
	alteredAddresses map[string]*types.AlteredAccount,
) (
	map[string]*types.Transaction,
	[]*types.Transaction,
) {
	transactions := make(map[string]*types.Transaction)
	rewardsTxs := make([]*types.Transaction, 0)

	for _, mb := range body.MiniBlocks {
		mbHash, err := tg.calculateHash(mb)
		if err != nil {
			continue
		}

		mbStatus := computeStatus(tg.selfShardID, mb.ReceiverShardID)

		switch mb.Type {
		case block.TxBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tg.txBuilder.buildTransaction(tx, []byte(hash), mbHash, mb, header, mbStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, tg.selfShardID, false)
				if tg.shouldIndex(mb.ReceiverShardID) {
					transactions[hash] = dbTx
				}
				delete(txPool, hash)
			}
		case block.InvalidBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tg.txBuilder.buildTransaction(tx, []byte(hash), mbHash, mb, header, transaction.TxStatusInvalid.String())
				addToAlteredAddresses(dbTx, alteredAddresses, mb, tg.selfShardID, false)

				dbTx.GasUsed = dbTx.GasLimit
				fee := tg.txBuilder.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, dbTx.GasUsed)
				dbTx.Fee = fee.String()

				transactions[hash] = dbTx
				delete(txPool, hash)
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				dbTx := tg.txBuilder.buildRewardTransaction(rtx, []byte(hash), mbHash, mb, header, mbStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, tg.selfShardID, true)
				if tg.shouldIndex(mb.ReceiverShardID) {
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

func (tg *txGrouper) shouldIndex(destinationShardID uint32) bool {
	if !tg.isInImportMode {
		return true
	}

	return tg.selfShardID == destinationShardID
}

func (tg *txGrouper) groupReceipts(header data.HeaderHandler, txPool map[string]data.TransactionHandler) []*types.Receipt {
	receipts := make(map[string]*receipt.Receipt)
	for hash, tx := range txPool {
		rec, ok := tx.(*receipt.Receipt)
		if !ok {
			continue
		}

		receipts[hash] = rec
		delete(txPool, hash)
	}

	dbReceipts := make([]*types.Receipt, 0)
	for recHash, rec := range receipts {
		dbReceipts = append(dbReceipts, tg.txBuilder.convertReceiptInDatabaseReceipt(recHash, rec, header))
	}

	return dbReceipts
}

func computeStatus(selfShardID uint32, receiverShardID uint32) string {
	if selfShardID == receiverShardID {
		return transaction.TxStatusSuccess.String()
	}

	return transaction.TxStatusPending.String()
}

func groupSmartContractResults(txPool map[string]data.TransactionHandler) map[string]*smartContractResult.SmartContractResult {
	scResults := make(map[string]*smartContractResult.SmartContractResult)
	for hash, tx := range txPool {
		scResult, ok := tx.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		scResults[hash] = scResult
	}

	return scResults
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

func convertMapTxsToSlice(txs map[string]*types.Transaction) []*types.Transaction {
	transactions := make([]*types.Transaction, len(txs))
	i := 0
	for _, tx := range txs {
		transactions[i] = tx
		i++
	}
	return transactions
}
