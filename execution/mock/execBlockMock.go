package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type ExecBlockMock struct {
	ProcessBlockCalled func(blockChain *blockchain.BlockChain, header *block.Header, body *block.Block) error
}

func (ebm *ExecBlockMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.Block) error {
	return ebm.ProcessBlockCalled(blockChain, header, body)
}
