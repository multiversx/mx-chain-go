package data

import "github.com/davecgh/go-spew/spew"

type BlockChainServiceImpl struct {
}

func (BlockChainServiceImpl) AddBlock(blockchain *BlockChain, block Block) {
	blockchain.blocks = append(blockchain.blocks, block)
}

func (BlockChainServiceImpl) GetCurrentBlock(blockChain *BlockChain) Block {
	if blockChain == nil || len(blockChain.blocks) == 0 {
		return NewBlock(-1, "", "", "", "", "")
	}

	return blockChain.blocks[len(blockChain.blocks)-1]
}

func (BlockChainServiceImpl) Print(blockChain *BlockChain) {
	spew.Dump(blockChain)
}
