package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (n *Node) SetMessenger(mes p2p.Messenger) {
	n.messenger = mes
}

func (n *Node) BroadcastBlock(blockBody *block.TxBlockBody, header *block.Header) error {
	return n.broadcastBlock(blockBody, header)
}
