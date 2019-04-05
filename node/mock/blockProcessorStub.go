package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// BlockProcessorStub mocks the implementation for a blockProcessor
type BlockProcessorStub struct {
	ProcessBlockCalled                 func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                  func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled           func()
	CreateGenesisBlockCalled           func(balances map[string]*big.Int) (rootHash []byte, err error)
	CreateBlockCalled                  func(round int32, haveTime func() bool) (data.BodyHandler, error)
	RemoveBlockInfoFromPoolCalled      func(body data.BodyHandler) error
	RestoreBlockInfoIntoPoolCalled     func(blockChain data.ChainHandler, body data.BodyHandler) error
	GetRootHashCalled                  func() []byte
	SetOnRequestTransactionCalled      func(f func(destShardID uint32, txHash []byte))
	CheckBlockValidityCalled           func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool
	MarshalizedDataForCrossShardCalled func(body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error)
}

// ProcessBlock mocks pocessing a block
func (blProcMock *BlockProcessorStub) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return blProcMock.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (blProcMock *BlockProcessorStub) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
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
func (blProcMock *BlockProcessorStub) CreateBlockBody(round int32, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockCalled(round, haveTime)
}

// RemoveBlockTxsFromPool mocks the removal of block transactions from transaction pools
func (blProcMock *BlockProcessorStub) RemoveBlockInfoFromPool(body data.BodyHandler) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockInfoFromPoolCalled(body)
}

func (blProcMock *BlockProcessorStub) RestoreBlockInfoIntoPool(blockChain data.ChainHandler, body data.BodyHandler) error {
	return blProcMock.RestoreBlockInfoIntoPoolCalled(blockChain, body)
}

// GetRootHash mocks getting root hash
func (blProcMock BlockProcessorStub) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}

func (blProcMock BlockProcessorStub) CheckBlockValidity(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool {
	return blProcMock.CheckBlockValidityCalled(blockChain, header, body)
}

func (bps BlockProcessorStub) CreateBlockHeader(body data.BodyHandler) (data.HeaderHandler, error) {
	panic("implement me")
}

func (blProcMock BlockProcessorStub) MarshalizedDataForCrossShard(body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error) {
	return blProcMock.MarshalizedDataForCrossShardCalled(body)
}
