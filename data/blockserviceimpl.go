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
	//	return hasher.HasherSha256{""}.CalculateHash(message).(string)
	var h hasher.IHasher = &hasher.HasherSha256{""}
	return h.CalculateHash(message).(string)
}

func (bsi BlockServiceImpl) PrintImpl() {
	fmt.Printf("Implementation type: %T\n", bsi)
}
