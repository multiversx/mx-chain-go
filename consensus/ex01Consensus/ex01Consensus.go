package main

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"time"
)

var CONSENSUS_GROUP_SIZE = 21
var ROUND_TIME = 4
var SELF_ID = 1

func main() {

	block := data.NewBlock(0, time.Now().String(), "", "", "", "")
	hash := service.GetBlockService().CalculateHash(&block)
	block.SetHash(hash)
	service.GetBlockService().Print(&block)

	blockChain := data.NewBlockChain(nil)
	service.GetBlockChainService().AddBlock(&blockChain, data.NewBlock(0, time.Now().String(), "", "", "", ""))
	service.GetBlockChainService().AddBlock(&blockChain, data.NewBlock(1, time.Now().String(), "", "", "", ""))
	service.GetBlockChainService().AddBlock(&blockChain, data.NewBlock(2, time.Now().String(), "", "", "", ""))
	service.GetBlockChainService().Print(&blockChain)
}
