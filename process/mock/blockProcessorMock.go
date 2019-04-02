package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type BlockProcessorMock struct {
	ProcessBlockCalled                 func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                  func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled           func()
	CreateGenesisBlockCalled           func(balances map[string]*big.Int) (rootHash []byte, err error)
	CreateBlockCalled                  func(round int32, haveTime func() bool) (data.BodyHandler, error)
	RemoveBlockInfoFromPoolCalled      func(body data.BodyHandler) error
	GetRootHashCalled                  func() []byte
	noShards                           uint32
	SetOnRequestTransactionCalled      func(f func(destShardID uint32, txHash []byte))
	CheckBlockValidityCalled           func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool
	CreateBlockHeaderCalled            func(body data.BodyHandler) (data.HeaderHandler, error)
	MarshalizedDataForCrossShardCalled func(body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error)
}

func (bpm *BlockProcessorMock) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bpm.ProcessBlockCalled(blockChain, header, body, haveTime)
}

func (bpm *BlockProcessorMock) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.CommitBlockCalled(blockChain, header, body)
}

func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

func (blProcMock BlockProcessorMock) CreateGenesisBlock(balances map[string]*big.Int) (rootHash []byte, err error) {
	return blProcMock.CreateGenesisBlockCalled(balances)
}

func (blProcMock BlockProcessorMock) CreateBlockBody(round int32, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockCalled(round, haveTime)
}

func (blProcMock BlockProcessorMock) RemoveBlockInfoFromPool(body data.BodyHandler) error {
	// pretend we removed the data
	return blProcMock.RemoveBlockInfoFromPoolCalled(body)
}

func (blProcMock BlockProcessorMock) GetRootHash() []byte {
	return blProcMock.GetRootHashCalled()
}

func (blProcMock BlockProcessorMock) CheckBlockValidity(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool {
	return blProcMock.CheckBlockValidityCalled(blockChain, header, nil)
}

func (blProcMock BlockProcessorMock) CreateBlockHeader(body data.BodyHandler) (data.HeaderHandler, error) {
	return blProcMock.CreateBlockHeaderCalled(body)
}

func (blProcMock BlockProcessorMock) MarshalizedDataForCrossShard(body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error) {
	return blProcMock.MarshalizedDataForCrossShardCalled(body)
}
