package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type BlockProcessorMock struct {
	ProcessBlockCalled           func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error
	CreateGenesisBlockCalled     func(balances map[string]big.Int, shardId uint32, genTime uint64) *block.StateBlockBody
	CreateTxBlockCalled          func(nbShards int, shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error)
	RemoveBlockTxsFromPoolCalled func(body *block.TxBlockBody)
}

func (blProcMock BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	panic("implement me")
}

func (blProcMock BlockProcessorMock) CreateGenesisBlock(balances map[string]big.Int, shardId uint32, genTime uint64) *block.StateBlockBody {
	panic("implement me")
}

func (blProcMock BlockProcessorMock) CreateTxBlock(nbShards int, shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error) {
	return blProcMock.CreateTxBlockCalled(nbShards, shardId, maxTxInBlock, haveTime)
}

func (blProcMock BlockProcessorMock) RemoveBlockTxsFromPool(body *block.TxBlockBody) {
	// pretend we removed the data
	blProcMock.RemoveBlockTxsFromPoolCalled(body)
}
