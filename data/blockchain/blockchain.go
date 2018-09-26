package data

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/davecgh/go-spew/spew"
)

type BlockChain struct {
	blocks []data.Block
}

func New(blocks []data.Block) BlockChain {

	blockChain := BlockChain{blocks}
	return blockChain
}

func (bc *BlockChain) SetBlocks(blocks []data.Block) {
	bc.blocks = blocks
}

func (bc *BlockChain) GetBlocks() []data.Block {
	return bc.blocks
}

// impl

type BlockChainImpl struct {
}

func (BlockChainImpl) AddBlock(blockchain *BlockChain, block data.Block) {
	blockchain.blocks = append(blockchain.blocks, block)
}

func (BlockChainImpl) GetCurrentBlock(blockChain *BlockChain) *data.Block {
	if blockChain == nil || len(blockChain.blocks) == 0 {
		return nil
	}

	return &blockChain.blocks[len(blockChain.blocks)-1]
}

func (BlockChainImpl) Print(blockChain *BlockChain) {
	spew.Dump(blockChain)
}
