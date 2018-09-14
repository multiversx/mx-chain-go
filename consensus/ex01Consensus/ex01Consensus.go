package main

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"time"
)

func main() {

	block := data.NewBlock(0, time.Now().String(), "", "", "", "")
	hash := service.GetBlockService().CalculateHash(block)
	block.SetHash(hash)
	block.Print()

}
