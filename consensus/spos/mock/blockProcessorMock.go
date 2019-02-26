package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorMock struct {
	ProcessBlockCalled            func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error
	ProcessAndCommitCalled        func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error
	CommitBlockCalled             func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error
	RevertAccountStateCalled      func()
	CreateGenesisBlockCalled      func(balances map[string]*big.Int, shardId uint32) (*block.StateBlockBody, error)
	CreateTxBlockCalled           func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error)
	CreateEmptyBlockBodyCalled    func(shardId uint32, round int32) *block.TxBlockBody
	RemoveBlockTxsFromPoolCalled  func(body *block.TxBlockBody) error
	GetRootHashCalled             func() []byte
	SetOnRequestTransactionCalled func(f func(destShardID uint32, txHash []byte))
	CheckBlockValidityCalled      func(blockChain *blockchain.BlockChain, header *block.Header) bool
}

// SetOnRequestTransaction mocks setting request transaction call back function
func (blProcMock *BlockProcessorMock) SetOnRequestTransaction(f func(destShardID uint32, txHash []byte)) {
	blProcMock.SetOnRequestTransactionCalled(f)
}

// ProcessBlock mocks pocessing a block
func (blProcMock *BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
	return blProcMock.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// ProcessAndCommit mocks processesing and committing a block
func (blProcMock *BlockProcessorMock) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
	return blProcMock.ProcessAndCommitCalled(blockChain, header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (blProcMock *BlockProcessorMock) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
	return blProcMock.CommitBlockCalled(blockChain, header, block)
}

// RevertAccountState mocks revert of the accounts state
func (blProcMock *BlockProcessorMock) RevertAccountState() {
	blProcMock.RevertAccountStateCalled()
}

// CreateGenesisBlockBody mocks the creation of a genesis block body
func (blProcMock *BlockProcessorMock) CreateGenesisBlockBody(balances map[string]*big.Int, shardId uint32) (*block.StateBlockBody, error) {
	return blProcMock.CreateGenesisBlockCalled(balances, shardId)
}

// CreateTxBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorMock) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
	return blProcMock.CreateTxBlockCalled(shardId, maxTxInBlock, round, haveTime)
}

// CreateEmptyBlockBody mocks the creation of an empty block body
func (blProcMock *BlockProcessorMock) CreateEmptyBlockBody(shardId uint32, round int32) *block.TxBlockBody {
	return blProcMock.CreateEmptyBlockBodyCalled(shardId, round)
}

// RemoveBlockTxsFromPool mocks the removal of block transactions from transaction pools
func (blProcMock *BlockProcessorMock) RemoveBlockTxsFromPool(body *block.TxBlockBody) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockTxsFromPoolCalled(body)
}

// GetRootHash mocks getting root hash
func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}

func (blProcMock BlockProcessorMock) CheckBlockValidity(blockChain *blockchain.BlockChain, header *block.Header) bool {
	return blProcMock.CheckBlockValidityCalled(blockChain, header)
}
