package data

import (
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/davecgh/go-spew/spew"
)

type BlockServiceImpl2 struct {
}

func (BlockServiceImpl2) CalculateHash(block *Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetMetaData() + block.GetPrevHash()
	var h hashing.Sha256Impl
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bsi BlockServiceImpl2) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bsi)
}

func (BlockServiceImpl2) Print(block *Block) {
	spew.Dump(block)
}
