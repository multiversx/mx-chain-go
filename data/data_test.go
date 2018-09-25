package data

import (
	block "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	blockchain "github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"testing"
	"time"
)

func TestBlock(t *testing.T) {

	bs := GetBlockerService()

	block := block.New(0, time.Now().String(), "", "", "", "Test")
	hash := bs.CalculateHash(&block)
	block.SetHash(hash)

	if block.GetHash() == "" {
		t.Fatal("Hash was not set")
	}

	bs.Print(&block)
}

func TestAddNewBlocksToBlockChain(t *testing.T) {

	bs := GetBlockerService()
	bcs := GetBlockChainerService()

	var blockChain blockchain.BlockChain
	var b block.Block
	var currentBlock *block.Block

	b = block.New(0, time.Now().String(), "", "", "", "")
	b.SetPrevHash("")
	b.SetHash(bs.CalculateHash(&b))
	bcs.AddBlock(&blockChain, b)

	b = block.New(1, time.Now().String(), "", "", "", "")
	currentBlock = bcs.GetCurrentBlock(&blockChain)
	b.SetPrevHash(currentBlock.GetHash())
	b.SetHash(bs.CalculateHash(&b))
	bcs.AddBlock(&blockChain, b)

	b = block.New(2, time.Now().String(), "", "", "", "")
	currentBlock = bcs.GetCurrentBlock(&blockChain)
	b.SetPrevHash(currentBlock.GetHash())
	b.SetHash(bs.CalculateHash(&b))
	bcs.AddBlock(&blockChain, b)

	if len(blockChain.GetBlocks()) != 3 {
		t.Fatal("Error: blocks count not match")
	}

	if bcs.GetCurrentBlock(&blockChain).GetNonce() != 2 {
		t.Fatal("Error: nonce not match")
	}

	bcs.Print(&blockChain)
}

func TestAddSameModifiedBlockToBlockChain(t *testing.T) {

	bs := GetBlockerService()
	bcs := GetBlockChainerService()

	var blockChain blockchain.BlockChain
	var b block.Block

	b = block.New(0, time.Now().String(), "", "", "", "")
	b.SetPrevHash("")
	b.SetHash(bs.CalculateHash(&b))
	bcs.AddBlock(&blockChain, b)

	b.SetNonce(1)
	b.SetPrevHash(b.GetHash())
	b.SetHash(bs.CalculateHash(&b))
	bcs.AddBlock(&blockChain, b)

	b.SetNonce(2)
	b.SetPrevHash(b.GetHash())
	b.SetHash(bs.CalculateHash(&b))
	bcs.AddBlock(&blockChain, b)

	if len(blockChain.GetBlocks()) != 3 {
		t.Fatal("Error: blocks count not match")
	}

	if bcs.GetCurrentBlock(&blockChain).GetNonce() != 2 {
		t.Fatal("Error: last block nonce not match")
	}

	if blockChain.GetBlocks()[0].GetNonce() != 0 {
		t.Fatal("Error: first block nonce not match")
	}

	bcs.Print(&blockChain)
}

func TestServiceProvider(t *testing.T) {

	PutService("Blocker", block.BlockImpl1{})
	bs := GetBlockerService()

	switch bs.(type) {
	case block.BlockImpl1:
		bs.PrintImpl()
	default:
		t.Fatal("Service implementation selecation error")
	}

	PutService("Blocker", block.BlockImpl2{})
	bs = GetBlockerService()

	switch bs.(type) {
	case block.BlockImpl2:
		bs.PrintImpl()
	default:
		t.Fatal("Service implementation selecation error")
	}
}
