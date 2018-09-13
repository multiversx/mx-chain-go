package data

type BlockChain struct {
	Blocks []Block
}

func (bc BlockChain) AddBlock(b Block) {
	bc.Blocks = append(bc.Blocks, b)
}
