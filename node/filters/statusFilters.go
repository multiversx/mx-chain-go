package filters

import (
	"bytes"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type statusFilters struct {
	selfShardID uint32
}

// NewStatusFilters will create a new instance of a statusFilters
func NewStatusFilters(selfShardID uint32) *statusFilters {
	return &statusFilters{
		selfShardID: selfShardID,
	}
}

// SetStatusIfIsFailedESDTTransfer will set the status if the provided transaction if a failed ESDT transfer
func (sf *statusFilters) SetStatusIfIsFailedESDTTransfer(tx *transaction.ApiTransactionResult) {
	if len(tx.SmartContractResults) < 1 {
		return
	}

	isCrossShardTxDestMe := tx.SourceShard != tx.DestinationShard && sf.selfShardID == tx.DestinationShard
	if !isCrossShardTxDestMe {
		return
	}

	if !isESDTTransfer(tx) {
		return
	}

	for _, scr := range tx.SmartContractResults {
		setStatusBasedOnSCRDataAndNonce(tx, []byte(scr.Data), scr.Nonce)
	}
}

func (sf *statusFilters) SetStatusIfFailedMoveBalanceWithError(tx *transaction.ApiTransactionResult) {
	if len(tx.SmartContractResults) < 1 {
		return
	}
	if !isMoveBalanceWithValue(tx) {
		return
	}

	if hasMirroredRefundWithSameValue(tx, apiTxValue(tx)) {
		tx.Status = transaction.TxStatusFail
	}
}

func hasMirroredRefundWithSameValue(tx *transaction.ApiTransactionResult, txValue *big.Int) bool {
	for _, scr := range tx.SmartContractResults {
		if scr == nil || scr.Value == nil {
			continue
		}
		if scr.Value.Cmp(txValue) != 0 {
			continue
		}
		if scr.SndAddr == tx.Receiver && scr.RcvAddr == tx.Sender {
			return true
		}
	}
	return false
}

// ApplyStatusFilters will apply status filters on the provided miniblocks
func (sf *statusFilters) ApplyStatusFilters(miniblocks []*api.MiniBlock) {
	for _, mb := range miniblocks {
		if mb.Type != block.TxBlock.String() {
			continue
		}

		isNotCrossShardDestinationMe := mb.SourceShard == mb.DestinationShard || mb.DestinationShard != sf.selfShardID
		if isNotCrossShardDestinationMe {
			continue
		}

		iterateMiniblockTxsForESDTTransfer(mb, miniblocks)
		iterateMiniblockTxsForFailedMoveBalance(mb, miniblocks)
	}
}

func iterateMiniblockTxsForESDTTransfer(miniblock *api.MiniBlock, miniblocks []*api.MiniBlock) {
	for _, tx := range miniblock.Transactions {
		if !isESDTTransfer(tx) {
			continue
		}

		searchUnsignedTransaction(tx, miniblocks)
	}
}

func searchUnsignedTransaction(tx *transaction.ApiTransactionResult, miniblocks []*api.MiniBlock) {
	for _, mb := range miniblocks {
		if mb.Type != block.SmartContractResultBlock.String() {
			continue
		}

		shouldCheckTransaction := mb.DestinationShard == tx.SourceShard && mb.SourceShard == tx.DestinationShard
		if shouldCheckTransaction {
			tryToSetStatusOfESDTTransfer(tx, mb)
		}
	}
}

func tryToSetStatusOfESDTTransfer(tx *transaction.ApiTransactionResult, miniblock *api.MiniBlock) {
	for _, unsignedTx := range miniblock.Transactions {
		if unsignedTx.OriginalTransactionHash != tx.Hash {
			continue
		}

		setStatusBasedOnSCRDataAndNonce(tx, unsignedTx.Data, unsignedTx.Nonce)
	}
}

func iterateMiniblockTxsForFailedMoveBalance(miniblock *api.MiniBlock, miniblocks []*api.MiniBlock) {
	for _, tx := range miniblock.Transactions {
		if !isMoveBalanceWithValue(tx) {
			continue
		}

		searchMirroredRefundSCR(tx, miniblocks)
	}
}

func isMoveBalanceWithValue(tx *transaction.ApiTransactionResult) bool {
	if tx == nil || tx.Tx == nil {
		return false
	}
	txValue := tx.Tx.GetValue()
	if txValue == nil || txValue.Cmp(big.NewInt(0)) <= 0 {
		return false
	}

	return !core.IsSmartContractAddress(tx.Tx.GetSndAddr()) && !core.IsSmartContractAddress(tx.Tx.GetRcvAddr())
}

func searchMirroredRefundSCR(tx *transaction.ApiTransactionResult, miniblocks []*api.MiniBlock) {
	for _, mb := range miniblocks {
		if mb.Type != block.SmartContractResultBlock.String() {
			continue
		}

		shouldCheckTransaction := mb.DestinationShard == tx.SourceShard && mb.SourceShard == tx.DestinationShard
		if shouldCheckTransaction {
			tryToSetStatusOfFailedMoveBalance(tx, mb)
		}
	}
}

func tryToSetStatusOfFailedMoveBalance(tx *transaction.ApiTransactionResult, miniblock *api.MiniBlock) {
	for _, scr := range miniblock.Transactions {
		if scr.OriginalTransactionHash != tx.Hash {
			continue
		}

		if isMirroredRefundWithSameValue(scr, tx) {
			tx.Status = transaction.TxStatusFail
			return
		}
	}
}

func isMirroredRefundWithSameValue(scr, tx *transaction.ApiTransactionResult) bool {
	if scr.Sender != tx.Receiver || scr.Receiver != tx.Sender {
		return false
	}

	scrValue := apiTxValue(scr)
	txValue := apiTxValue(tx)
	if scrValue == nil || txValue == nil {
		return false
	}

	return scrValue.Cmp(txValue) == 0
}

func apiTxValue(tx *transaction.ApiTransactionResult) *big.Int {
	if check.IfNil(tx.Tx) {
		return big.NewInt(0)
	}

	return tx.Tx.GetValue()
}

func setStatusBasedOnSCRDataAndNonce(tx *transaction.ApiTransactionResult, scrDataField []byte, scrNonce uint64) {
	isSCRWithRefund := bytes.HasPrefix(scrDataField, tx.Data) && scrNonce == tx.Nonce
	if isSCRWithRefund {
		tx.Status = transaction.TxStatusFail
		return
	}
}

func isESDTTransfer(tx *transaction.ApiTransactionResult) bool {
	return strings.HasPrefix(string(tx.Data), core.BuiltInFunctionESDTTransfer)
}
