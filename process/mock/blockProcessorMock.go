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
	nbShards                     uint32
}

func (bpm *BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	return bpm.ProcessBlockCalled(blockChain, header, body)
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

func (blProcMock BlockProcessorMock) GetNbShards() uint32 {
	return blProcMock.nbShards
}

func (blProcMock BlockProcessorMock) SetNbShards(nbShards uint32) {
	blProcMock.nbShards = nbShards
}

func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}
