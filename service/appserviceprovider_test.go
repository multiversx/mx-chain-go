package service

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"testing"
)

func TestAppServiceProvider(t *testing.T) {

	PutService("IBlockService", data.BlockServiceImpl{})
	blockService := GetBlockService()

	switch blockService.(type) {
	case data.BlockServiceImpl:
		blockService.PrintImpl()
	default:
		t.Fatal("Service implementation selecation error")
	}

	PutService("IBlockService", data.BlockServiceImpl2{})
	blockService = GetBlockService()

	switch blockService.(type) {
	case data.BlockServiceImpl2:
		blockService.PrintImpl()
	default:
		t.Fatal("Service implementation selecation error")
	}
}
