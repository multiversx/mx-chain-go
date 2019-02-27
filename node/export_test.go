package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (n *Node) SetMessenger(mes p2p.Messenger) {
	n.messenger = mes
}

func (n *Node) BroadcastTxBlockBody(blockBody *block.TxBlockBody) error {
	return n.broadcastTxBlockBody(blockBody)
}

func (n *Node) BroadcastHeader(header *block.Header) error {
	return n.broadcastHeader(header)
}
