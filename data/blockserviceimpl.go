package data

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/hasher"
	"strconv"
)

type BlockServiceImpl struct {
}

func (BlockServiceImpl) CalculateHash(block Block) string {
	message := strconv.Itoa(block.GetNonce()) + block.GetTimeStamp() + block.GetMetaData() + block.GetPrevHash()
	var h hasher.HasherSha256
	hash := h.CalculateHash(message)
	return hash.(string)
}

func (bsi BlockServiceImpl) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bsi)
}
