package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type BlockProcessorMock struct {
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (rootHash []byte, err error)
	CreateBlockCalled                func(round int32, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	noShards                         uint32
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	CreateBlockHeaderCalled          func(body data.BodyHandler, round int32, haveTime func() bool) (data.HeaderHandler, error)
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error)
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

func (blProcMock BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.RestoreBlockIntoPoolsCalled(header, body)
}

func (blProcMock BlockProcessorMock) CreateBlockHeader(body data.BodyHandler, round int32, haveTime func() bool) (data.HeaderHandler, error) {
	return blProcMock.CreateBlockHeaderCalled(body, round, haveTime)
}

func (blProcMock BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error) {
	return blProcMock.MarshalizedDataToBroadcastCalled(header, body)
}
