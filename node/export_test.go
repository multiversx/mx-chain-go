package node

import (
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/factory"
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

// AddClosableComponents -
func (n *Node) AddClosableComponents(components ...factory.Closer) {
	n.closableComponents = append(n.closableComponents, components...)
}

// ComputeTimestampForRound -
func (n *Node) ComputeTimestampForRound(round uint64) int64 {
	return n.computeTimestampForRound(round)
}
