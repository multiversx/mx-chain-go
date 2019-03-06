package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorStub struct {
	ProcessBlockCalled            func(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled             func(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled      func()
	CreateGenesisBlockCalled      func(balances map[string]*big.Int) (rootHash []byte, err error)
	CreateBlockCalled             func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (data.BodyHandler, error)
	RemoveBlockInfoFromPoolCalled func(body data.BodyHandler) error
	GetRootHashCalled             func() []byte
	SetOnRequestTransactionCalled func(f func(destShardID uint32, txHash []byte))
	CheckBlockValidityCalled      func(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler) bool
}

// SetOnRequestTransaction mocks setting request transaction call back function
func (blProcMock *BlockProcessorStub) SetOnRequestTransaction(f func(destShardID uint32, txHash []byte)) {
	blProcMock.SetOnRequestTransactionCalled(f)
}

// ProcessBlock mocks pocessing a block
func (blProcMock *BlockProcessorStub) ProcessBlock(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return blProcMock.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (blProcMock *BlockProcessorStub) CommitBlock(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.CommitBlockCalled(blockChain, header, body)
}

// RevertAccountState mocks revert of the accounts state
func (blProcMock *BlockProcessorStub) RevertAccountState() {
	blProcMock.RevertAccountStateCalled()
}

// CreateGenesisBlock mocks the creation of a genesis block body
func (blProcMock *BlockProcessorStub) CreateGenesisBlock(balances map[string]*big.Int) (rootHash []byte, err error) {
	return blProcMock.CreateGenesisBlockCalled(balances)
}

// CreateTxBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorStub) CreateBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockCalled(shardId, maxTxInBlock, round, haveTime)
}

// RemoveBlockTxsFromPool mocks the removal of block transactions from transaction pools
func (blProcMock *BlockProcessorStub) RemoveBlockInfoFromPool(body data.BodyHandler) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockInfoFromPoolCalled(body)
}

// GetRootHash mocks getting root hash
func (blProcMock BlockProcessorStub) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}

func (blProcMock BlockProcessorStub) CheckBlockValidity(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler) bool {
	return blProcMock.CheckBlockValidityCalled(blockChain, header, body)
}
