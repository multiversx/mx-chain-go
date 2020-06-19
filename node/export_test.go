package node

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (n *Node) CreateConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	return n.createConsensusTopic(messageProcessor)
}

func (n *Node) ComputeTransactionStatus(tx data.TransactionHandler, isInPool bool) core.TransactionStatus {
	return n.computeTransactionStatus(tx, isInPool)
}
