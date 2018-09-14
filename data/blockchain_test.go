package data

import (
	"testing"
	"time"
)

func TestBlockChain(t *testing.T) {

	var bcsi BlockChainServiceImpl
	var blockChain BlockChain
	var block Block

	block = NewBlock(0, time.Now().String(), "", "", "", "")
	bcsi.AddBlock(&blockChain, block)

	block = NewBlock(1, time.Now().String(), "", "", "", "")
	bcsi.AddBlock(&blockChain, block)

	block = NewBlock(2, time.Now().String(), "", "", "", "")
	bcsi.AddBlock(&blockChain, block)

	if len(blockChain.blocks) != 3 {
		t.Fatal("Error: blocks count not match")
	}

	if bcsi.GetCurrentBlock(&blockChain).nonce != 2 {
		t.Fatal("Error: nonce not match")
	}

	bcsi.Print(&blockChain)
}
