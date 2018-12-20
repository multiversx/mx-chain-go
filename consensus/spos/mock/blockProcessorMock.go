package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type BlockProcessorMock struct {
	ProcessBlockCalled           func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error
	CreateGenesisBlockCalled     func(balances map[string]big.Int, shardId uint32) *block.StateBlockBody
	CreateTxBlockCalled          func(shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error)
	RemoveBlockTxsFromPoolCalled func(body *block.TxBlockBody) error
	GetRootHashCalled            func() []byte
	noShards                     uint32
}

func (blProcMock BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	panic("implement me")
}

func (blProcMock BlockProcessorMock) CreateGenesisBlockBody(balances map[string]big.Int, shardId uint32) *block.StateBlockBody {
	panic("implement me")
}

func (blProcMock BlockProcessorMock) CreateTxBlockBody(shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error) {
	return blProcMock.CreateTxBlockCalled(shardId, maxTxInBlock, haveTime)
}

func (blProcMock BlockProcessorMock) RemoveBlockTxsFromPool(body *block.TxBlockBody) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockTxsFromPoolCalled(body)
}

func (blProcMock BlockProcessorMock) NoShards() uint32 {
	return blProcMock.noShards
}

func (blProcMock BlockProcessorMock) SetNoShards(noShards uint32) {
	blProcMock.noShards = noShards
}

func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}
