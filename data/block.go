package data

import (
	"fmt"
)

type Block struct {
	nonce     int
	timeStamp string
	signature string
	hash      string
	prevHash  string
	metaData  string
}

func NewBlock(nonce int, timeStamp string, signature string, hash string, prevhash string, metaData string) Block {

	b := Block{nonce, timeStamp, signature, hash, prevhash, metaData}
	return b
}

func (b *Block) SetNonce(nonce int) {
	b.nonce = nonce
}

func (b Block) GetNonce() int {
	return b.nonce
}

func (b *Block) SetTimeStamp(timeStamp string) {
	b.timeStamp = timeStamp
}

func (b Block) GetTimeStamp() string {
	return b.timeStamp
}

func (b *Block) SetSignature(signature string) {
	b.signature = signature
}

func (b Block) GetSignature() string {
	return b.signature
}

func (b *Block) SetHash(hash string) {
	b.hash = hash
}

func (b Block) GetHash() string {
	return b.hash
}

func (b *Block) SetPrevHash(prevHash string) {
	b.prevHash = prevHash
}

func (b Block) GetPrevHash() string {
	return b.prevHash
}

func (b *Block) SetMetaData(metaData string) {
	b.metaData = metaData
}

func (b Block) GetMetaData() string {
	return b.metaData
}

func (b Block) Print() {
	fmt.Println(b)
}
