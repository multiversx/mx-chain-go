package node

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (n *Node) CreateConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	return n.createConsensusTopic(messageProcessor)
}

func PutHistoryFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *fullHistory.MiniblockMetadata) *transaction.ApiTransactionResult {
	return putHistoryFieldsInTransaction(tx, miniblockMetadata)
}
