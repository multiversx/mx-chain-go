package data

type BlockChain struct {
	blocks []Block
}

func NewBlockChain(blocks []Block) BlockChain {

	blockChain := BlockChain{blocks}
	return blockChain
}

func (bc *BlockChain) SetBlocks(blocks []Block) {
	bc.blocks = blocks
}

func (bc *BlockChain) GetBlocks() []Block {
	return bc.blocks
}
