package block

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/davecgh/go-spew/spew"
	"strconv"
)

type Block struct {
	Nonce     int
	TimeStamp string
	Signature string
	Hash      string
	PrevHash  string
	MetaData  string
}

func New(nonce int, timeStamp string, signature string, hash string, prevhash string, metaData string) Block {
	b := Block{nonce, timeStamp, signature, hash, prevhash, metaData}
	return b
}

func (b *Block) ResetBlock() {
	*b = New(-1, "", "", "", "", "")
}

func (b *Block) CalculateHash() string {
	message := strconv.Itoa(b.Nonce) + b.TimeStamp + b.MetaData + b.PrevHash
	var h hashing.Sha256
	hash := h.Compute(message)
	return hex.EncodeToString(hash)
}

func (b *Block) Print() {
	spew.Dump(b)
}
