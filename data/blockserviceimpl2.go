package data

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/hasher"
	"github.com/davecgh/go-spew/spew"
	"strconv"
)

type BlockServiceImpl2 struct {
}

func (BlockServiceImpl2) CalculateHash(block *Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetMetaData() + block.GetPrevHash()
	var h hasher.Sha256Impl
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bsi BlockServiceImpl2) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bsi)
}

func (BlockServiceImpl2) Print(block *Block) {
	spew.Dump(block)
}
