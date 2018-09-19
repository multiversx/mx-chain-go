package data

type IBlockChainService interface {
	AddBlock(*BlockChain, Block)
	GetCurrentBlock(*BlockChain) *Block
	Print(*BlockChain)
}
