package data

import (
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/davecgh/go-spew/spew"
)

type BlockServiceImpl struct {
}

func (BlockServiceImpl) CalculateHash(block *Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetTimeStamp() + block.GetMetaData() + block.GetPrevHash()
	var h hashing.Sha256Impl
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bsi BlockServiceImpl) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bsi)
}

func (BlockServiceImpl) Print(block *Block) {
	spew.Dump(block)
}
