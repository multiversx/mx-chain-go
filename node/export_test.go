package node

import (
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// PutMiniblockFieldsInTransaction -
func PutMiniblockFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *dblookupext.MiniblockMetadata) *transaction.ApiTransactionResult {
	return putMiniblockFieldsInTransaction(tx, miniblockMetadata)
}

// PutResultsInTransaction -
func (n *Node) PutResultsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	n.putResultsInTransaction(hash, tx, epoch)
}

func (n *Node) PrepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	return n.prepareUnsignedTx(tx)
}
