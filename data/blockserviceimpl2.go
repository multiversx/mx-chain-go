package data

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/hasher"
	"strconv"
)

type BlockServiceImpl2 struct {
}

func (BlockServiceImpl2) CalculateHash(block Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetMetaData() + block.GetPrevHash()
	var h hasher.HasherSha256
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bsi BlockServiceImpl2) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bsi)
}
