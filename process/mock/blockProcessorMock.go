package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type BlockProcessorMock struct {
	ProcessBlockCalled            func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, maxProcessingTime time.Duration) error
	ProcessAndCommitCalled        func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error
	CommitBlockCalled             func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error
	RevertAccountStateCalled      func()
	CreateGenesisBlockCalled      func(balances map[string]big.Int, shardId uint32) *block.StateBlockBody
	CreateTxBlockCalled           func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error)
	CreateEmptyBlockBodyCalled    func(shardId uint32, round int32) *block.TxBlockBody
	RemoveBlockTxsFromPoolCalled  func(body *block.TxBlockBody) error
	GetRootHashCalled             func() []byte
	noShards                      uint32
	SetOnRequestTransactionCalled func(f func(destShardID uint32, txHash []byte))
}

func (bpm *BlockProcessorMock) SetOnRequestTransaction(f func(destShardID uint32, txHash []byte)) {
	bpm.SetOnRequestTransactionCalled(f)
}

func (bpm *BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, maxProcessingTime time.Duration) error {
	return bpm.ProcessBlockCalled(blockChain, header, body, maxProcessingTime)
}

func (bpm *BlockProcessorMock) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	return bpm.ProcessAndCommitCalled(blockChain, header, body)
}

func (bpm *BlockProcessorMock) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
	return bpm.CommitBlockCalled(blockChain, header, block)
}

func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

func (blProcMock BlockProcessorMock) CreateGenesisBlockBody(balances map[string]big.Int, shardId uint32) *block.StateBlockBody {
	panic("implement me")
}

func (blProcMock BlockProcessorMock) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
	return blProcMock.CreateTxBlockCalled(shardId, maxTxInBlock, round, haveTime)
}

func (blProcMock BlockProcessorMock) CreateEmptyBlockBody(shardId uint32, round int32) *block.TxBlockBody {
	return blProcMock.CreateEmptyBlockBodyCalled(shardId, round)
}

func (blProcMock BlockProcessorMock) RemoveBlockTxsFromPool(body *block.TxBlockBody) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockTxsFromPoolCalled(body)
}

func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}
