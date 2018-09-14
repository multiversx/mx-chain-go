package data

import (
	_ "github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

func TestBlock(t *testing.T) {

	block := NewBlock(0, time.Now().String(), "", "", "", "Test")
	block.SetHash(BlockServiceImpl{}.CalculateHash(block))

	if block.GetHash() == "" {
		t.Fatal("Hash was not set")
	}

	//	spew.Dump(block)
	block.Print()
}
