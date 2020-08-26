package node

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

func (n *Node) ComputeTransactionStatus(tx data.TransactionHandler, isInPool bool) core.TransactionStatus {
	return n.computeTransactionStatus(tx, isInPool)
}

func PutHistoryFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *fullHistory.MiniblockMetadata) *transaction.ApiTransactionResult {
	return putHistoryFieldsInTransaction(tx, miniblockMetadata)
}
