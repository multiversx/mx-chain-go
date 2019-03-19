package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorMock struct {
	ProcessBlockCalled            func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled             func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled      func()
	CreateGenesisBlockCalled      func(balances map[string]*big.Int) (rootHash []byte, err error)
	CreateBlockCalled             func(round int32, haveTime func() bool) (data.BodyHandler, error)
	RemoveBlockInfoFromPoolCalled func(body data.BodyHandler) error
	GetRootHashCalled             func() []byte
	SetOnRequestTransactionCalled func(f func(destShardID uint32, txHash []byte))
	CheckBlockValidityCalled      func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool
	CreateBlockHeaderCalled       func(body data.BodyHandler) (data.HeaderHandler, error)
}

// SetOnRequestTransaction mocks setting request transaction call back function
func (blProcMock *BlockProcessorMock) SetOnRequestTransaction(f func(destShardID uint32, txHash []byte)) {
	blProcMock.SetOnRequestTransactionCalled(f)
}

// ProcessBlock mocks pocessing a block
func (blProcMock *BlockProcessorMock) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return blProcMock.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (blProcMock *BlockProcessorMock) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.CommitBlockCalled(blockChain, header, body)
}

// RevertAccountState mocks revert of the accounts state
func (blProcMock *BlockProcessorMock) RevertAccountState() {
	blProcMock.RevertAccountStateCalled()
}

// CreateGenesisBlock mocks the creation of a genesis block body
func (blProcMock *BlockProcessorMock) CreateGenesisBlock(balances map[string]*big.Int) (rootHash []byte, err error) {
	return blProcMock.CreateGenesisBlockCalled(balances)
}

// CreateTxBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorMock) CreateBlockBody(round int32, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockCalled(round, haveTime)
}

// RemoveBlockTxsFromPool mocks the removal of block transactions from transaction pools
func (blProcMock *BlockProcessorMock) RemoveBlockInfoFromPool(body data.BodyHandler) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockInfoFromPoolCalled(body)
}

// GetRootHash mocks getting root hash
func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}

func (blProcMock BlockProcessorMock) CheckBlockValidity(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool {
	return blProcMock.CheckBlockValidityCalled(blockChain, header, body)
}

func (blProcMock BlockProcessorMock) CreateBlockHeader(body data.BodyHandler) (data.HeaderHandler, error) {
	return blProcMock.CreateBlockHeaderCalled(body)
}
