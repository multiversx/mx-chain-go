package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
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

func (n *Node) DisplayLogDataAndComputeNewNoncePrevHash(
	sposWrk *spos.SPOSConsensusWorker,
	hdr *block.Header,
	txBlock *block.TxBlockBody,
	prevHash []byte) (uint64, []byte, error) {

	return n.displayLogDataAndComputeNewNoncePrevHash(sposWrk, hdr, txBlock, prevHash)
}
