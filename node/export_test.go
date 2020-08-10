package node

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

func (n *Node) ComputeTransactionStatus(tx data.TransactionHandler, isInPool bool) core.TransactionStatus {
	return n.computeTransactionStatus(tx, isInPool)
}
