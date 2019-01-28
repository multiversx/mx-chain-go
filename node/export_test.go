package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (n *Node) SetMessenger(mes p2p.Messenger) {
	n.messenger = mes
}
