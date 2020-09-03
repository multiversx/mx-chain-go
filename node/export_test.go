package node

import (
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (n *Node) CreateConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	return n.createConsensusTopic(messageProcessor)
}

func PutMiniblockFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *dblookupext.MiniblockMetadata) *transaction.ApiTransactionResult {
	return putMiniblockFieldsInTransaction(tx, miniblockMetadata)
}
