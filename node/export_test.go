package node

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

func PutHistoryFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *fullHistory.MiniblockMetadata) *transaction.ApiTransactionResult {
	return putHistoryFieldsInTransaction(tx, miniblockMetadata)
}
