package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

func (n *Node) SetMessenger(mes p2p.Messenger) {
	n.messenger = mes
}

func (n *Node) Interceptors() []process.Interceptor {
	return n.interceptors
}

func (n *Node) Resolvers() []process.Resolver {
	return n.resolvers
}
