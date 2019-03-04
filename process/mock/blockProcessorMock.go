package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type BlockProcessorMock struct {
	ProcessBlockCalled            func(blockChain *blockchain.BlockChain, header *block.Header, body block.Body, haveTime func() time.Duration) error
	ProcessAndCommitCalled        func(blockChain *blockchain.BlockChain, header *block.Header, body block.Body, haveTime func() time.Duration) error
	CommitBlockCalled             func(blockChain *blockchain.BlockChain, header *block.Header, block block.Body) error
	RevertAccountStateCalled      func()
	CreateGenesisBlockCalled      func(balances map[string]*big.Int) (rootHash []byte, err error)
	CreateTxBlockCalled           func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (block.Body, error)
	RemoveBlockTxsFromPoolCalled  func(body block.Body) error
	GetRootHashCalled             func() []byte
	noShards                      uint32
	SetOnRequestTransactionCalled func(f func(destShardID uint32, txHash []byte))
	CheckBlockValidityCalled      func(blockChain *blockchain.BlockChain, header *block.Header) bool
}

func (bpm *BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body block.Body, haveTime func() time.Duration) error {
	return bpm.ProcessBlockCalled(blockChain, header, body, haveTime)
}

func (bpm *BlockProcessorMock) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body block.Body, haveTime func() time.Duration) error {
	return bpm.ProcessAndCommitCalled(blockChain, header, body, haveTime)
}

func (bpm *BlockProcessorMock) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block block.Body) error {
	return bpm.CommitBlockCalled(blockChain, header, block)
}

func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

func (blProcMock BlockProcessorMock) CreateGenesisBlock(balances map[string]*big.Int) (rootHash []byte, err error) {
	return blProcMock.CreateGenesisBlockCalled(balances)
}

func (blProcMock BlockProcessorMock) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (block.Body, error) {
	return blProcMock.CreateTxBlockCalled(shardId, maxTxInBlock, round, haveTime)
}

func (blProcMock BlockProcessorMock) RemoveBlockTxsFromPool(body block.Body) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockTxsFromPoolCalled(body)
}

func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}

func (blProcMock BlockProcessorMock) CheckBlockValidity(blockChain *blockchain.BlockChain, header *block.Header) bool {
	return blProcMock.CheckBlockValidityCalled(blockChain, header)
}
