package filters

import (
	"bytes"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
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
