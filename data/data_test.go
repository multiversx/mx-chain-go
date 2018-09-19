package data

import (
	"testing"
	"time"
)

func TestBlock(t *testing.T) {

	var bsi BlockServiceImpl

	block := NewBlock(0, time.Now().String(), "", "", "", "Test")
	hash := bsi.CalculateHash(&block)
	block.SetHash(hash)

	if block.GetHash() == "" {
		t.Fatal("Hash was not set")
	}

	bsi.Print(&block)
}

func TestAddNewBlocksToBlockChain(t *testing.T) {

	var bcsi BlockChainServiceImpl
	var blockChain BlockChain
	var block Block
	var currentBlock *Block

	block = NewBlock(0, time.Now().String(), "", "", "", "")
	block.SetPrevHash("")
	block.SetHash(BlockServiceImpl{}.CalculateHash(&block))
	bcsi.AddBlock(&blockChain, block)

	block = NewBlock(1, time.Now().String(), "", "", "", "")
	currentBlock = bcsi.GetCurrentBlock(&blockChain)
	block.SetPrevHash(currentBlock.GetHash())
	block.SetHash(BlockServiceImpl{}.CalculateHash(&block))
	bcsi.AddBlock(&blockChain, block)

	block = NewBlock(2, time.Now().String(), "", "", "", "")
	currentBlock = bcsi.GetCurrentBlock(&blockChain)
	block.SetPrevHash(currentBlock.GetHash())
	block.SetHash(BlockServiceImpl{}.CalculateHash(&block))
	bcsi.AddBlock(&blockChain, block)

	if len(blockChain.blocks) != 3 {
		t.Fatal("Error: blocks count not match")
	}

	if bcsi.GetCurrentBlock(&blockChain).nonce != 2 {
		t.Fatal("Error: nonce not match")
	}

	bcsi.Print(&blockChain)
}

func TestAddSameModifiedBlockToBlockChain(t *testing.T) {

	var bcsi BlockChainServiceImpl
	var blockChain BlockChain
	var block Block

	block = NewBlock(0, time.Now().String(), "", "", "", "")
	block.SetPrevHash("")
	block.SetHash(BlockServiceImpl{}.CalculateHash(&block))
	bcsi.AddBlock(&blockChain, block)

	block.SetNonce(1)
	block.SetPrevHash(block.GetHash())
	block.SetHash(BlockServiceImpl{}.CalculateHash(&block))
	bcsi.AddBlock(&blockChain, block)

	block.SetNonce(2)
	block.SetPrevHash(block.GetHash())
	block.SetHash(BlockServiceImpl{}.CalculateHash(&block))
	bcsi.AddBlock(&blockChain, block)

	if len(blockChain.blocks) != 3 {
		t.Fatal("Error: blocks count not match")
	}

	if bcsi.GetCurrentBlock(&blockChain).nonce != 2 {
		t.Fatal("Error: last block nonce not match")
	}

	if blockChain.GetBlocks()[0].GetNonce() != 0 {
		t.Fatal("Error: first block nonce not match")
	}

	bcsi.Print(&blockChain)
}
