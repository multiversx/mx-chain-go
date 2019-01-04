package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

// BlockProcessorMock mocks the implementation for a BlockProcessor
type BlockProcessorMock struct {
	ProcessBlockCalled           func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error
	ProcessAndCommitCalled       func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error
	CommitBlockCalled            func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error
	RevertAccountStateCalled     func()
	CreateGenesisBlockCalled     func(balances map[string]big.Int, shardId uint32) *block.StateBlockBody
	CreateTxBlockCalled          func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error)
	RemoveBlockTxsFromPoolCalled func(body *block.TxBlockBody) error
	GetRootHashCalled            func() []byte
	noShards                     uint32
}

// ProcessBlock mocks pocessing a block
func (blProcMock *BlockProcessorMock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	return blProcMock.ProcessBlockCalled(blockChain, header, body)
}

// ProcessAndCommit mocks processesing and committing a block
func (blProcMock *BlockProcessorMock) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	return blProcMock.ProcessAndCommitCalled(blockChain, header, body)
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
func (blProcMock *BlockProcessorMock) CreateGenesisBlockBody(balances map[string]big.Int, shardId uint32) *block.StateBlockBody {
	panic("implement me")
}

// CreateTxBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorMock) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
	return blProcMock.CreateTxBlockCalled(shardId, maxTxInBlock, round, haveTime)
}

// RemoveBlockTxsFromPool mocks the removal of block transactions from transaction pools
func (blProcMock *BlockProcessorMock) RemoveBlockTxsFromPool(body *block.TxBlockBody) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockTxsFromPoolCalled(body)
}

// NoShards mocks the number of shards
func (blProcMock *BlockProcessorMock) NoShards() uint32 {
	return blProcMock.noShards
}

// SetNoShards mocks setting the number of shards
func (blProcMock *BlockProcessorMock) SetNoShards(noShards uint32) {
	blProcMock.noShards = noShards
}

// GetRootHash mocks the current root hash of the account state trie
func (blProcMock *BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}
