package data

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/davecgh/go-spew/spew"
	"strconv"
	"testing"
	"time"
)

func TestBlock(t *testing.T) {

	//	block := NewBlock(0, time.Now().String(), "", "", "", "Test")
	block := Block{0, time.Now().String(), "", "", "", "Test"}
	x := CalculateBlockHash(block)
	block.SetHash(x)

	if block.GetNonce() != 0 {
		t.Fatal("Eroare nonce")
	}

	spew.Dump(block)
	block.Print()
}

func CalculateBlockHash(block Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetTimeStamp() + block.GetMetaData() + block.GetPrevHash()
	h := sha256.New()
	h.Write([]byte(message))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
