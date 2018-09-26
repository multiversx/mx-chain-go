package data

import (
	block "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	blockchain "github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type Blocker interface {
	CalculateHash(*block.Block) string
	PrintImpl()
	Print(*block.Block)
}

type BlockChainer interface {
	AddBlock(*blockchain.BlockChain, block.Block)
	GetCurrentBlock(*blockchain.BlockChain) *block.Block
	Print(*blockchain.BlockChain)
}
