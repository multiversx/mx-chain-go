package data

import (
	_ "github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

func TestBlock(t *testing.T) {

	var bsi BlockServiceImpl

	block := NewBlock(0, time.Now().String(), "", "", "", "Test")
	hash := bsi.CalculateHash(block)
	block.SetHash(hash)

	if block.GetHash() == "" {
		t.Fatal("Hash was not set")
	}

	//	spew.Dump(block)
	block.Print()
}
