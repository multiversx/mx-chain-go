package data

type BlockChainServiceImpl struct {
}

func (BlockChainServiceImpl) AddBlock(blockchain BlockChain, block Block) {
	blockchain.Blocks = append(blockchain.Blocks, block)
}
