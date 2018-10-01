package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/davecgh/go-spew/spew"
)

type BlockChain struct {
	blocks []block.Block
}

func New(blocks []block.Block) BlockChain {

	blockChain := BlockChain{blocks}
	return blockChain
}

func (bc *BlockChain) SetBlocks(blocks []block.Block) {
	bc.blocks = blocks
}

func (bc *BlockChain) GetBlocks() []block.Block {
	return bc.blocks
}

// impl

type BlockChainImpl struct {
}

func (BlockChainImpl) AddBlock(blockchain *BlockChain, block block.Block) {
	blockchain.blocks = append(blockchain.blocks, block)
}

func (BlockChainImpl) GetCurrentBlock(blockChain *BlockChain) *block.Block {
	if blockChain == nil || len(blockChain.blocks) == 0 {
		return nil
	}

	return &blockChain.blocks[len(blockChain.blocks)-1]
}

func (BlockChainImpl) Print(blockChain *BlockChain) {
	spew.Dump(blockChain)
}
