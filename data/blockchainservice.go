package data

type BlockChainService interface {
	AddBlock(BlockChain, Block)
}
