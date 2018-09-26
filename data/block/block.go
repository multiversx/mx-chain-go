package data

import (
	"fmt"
	sha256 "github.com/ElrondNetwork/elrond-go-sandbox/hasher/sha256"
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

func (b *Block) SetNonce(nonce int) {
	b.Nonce = nonce
}

func (b *Block) GetNonce() int {
	return b.Nonce
}

func (b *Block) SetTimeStamp(timeStamp string) {
	b.TimeStamp = timeStamp
}

func (b *Block) GetTimeStamp() string {
	return b.TimeStamp
}

func (b *Block) SetSignature(signature string) {
	b.Signature = signature
}

func (b *Block) GetSignature() string {
	return b.Signature
}

func (b *Block) SetHash(hash string) {
	b.Hash = hash
}

func (b *Block) GetHash() string {
	return b.Hash
}

func (b *Block) SetPrevHash(prevHash string) {
	b.PrevHash = prevHash
}

func (b *Block) GetPrevHash() string {
	return b.PrevHash
}

func (b *Block) SetMetaData(metaData string) {
	b.MetaData = metaData
}

func (b *Block) GetMetaData() string {
	return b.MetaData
}

// impl1

type BlockImpl1 struct {
}

func (BlockImpl1) CalculateHash(block *Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetTimeStamp() + block.GetMetaData() + block.GetPrevHash()
	var h sha256.Sha256Impl
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bi BlockImpl1) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bi)
}

func (BlockImpl1) Print(block *Block) {
	spew.Dump(block)
}

// impl2

type BlockImpl2 struct {
}

func (BlockImpl2) CalculateHash(block *Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetMetaData() + block.GetPrevHash()
	var h sha256.Sha256Impl
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bi BlockImpl2) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bi)
}

func (BlockImpl2) Print(block *Block) {
	spew.Dump(block)
}
