package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorMock struct {
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	RevertStateToBlockCalled         func(header data.HeaderHandler) error
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockCalled                func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	ApplyBodyToHeaderCalled          func(hdr data.HeaderHandler, body data.BodyHandler) error
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled        func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled            func() data.HeaderHandler
}

func (blProcMock *BlockProcessorMock) CreateNewHeader() data.HeaderHandler {
	return blProcMock.CreateNewHeaderCalled()
}

// ProcessBlock mocks pocessing a block
func (blProcMock *BlockProcessorMock) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return blProcMock.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (blProcMock *BlockProcessorMock) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.CommitBlockCalled(blockChain, header, body)
}

// RevertStateToBlock recreates thee state tries to the root hashes indicated by the provided header
func (blProcMock *BlockProcessorMock) RevertStateToBlock(header data.HeaderHandler) error {
	if blProcMock.RevertStateToBlockCalled != nil {
		return blProcMock.RevertStateToBlock(header)
	}
	return nil
}

// RevertAccountState mocks revert of the accounts state
func (blProcMock *BlockProcessorMock) RevertAccountState() {
	blProcMock.RevertAccountStateCalled()
}

// CreateBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorMock) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockCalled(initialHdrData, haveTime)
}

func (blProcMock *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.RestoreBlockIntoPoolsCalled(header, body)
}

func (blProcMock BlockProcessorMock) ApplyBodyToHeader(hdr data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.ApplyBodyToHeaderCalled(hdr, body)
}

func (blProcMock BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return blProcMock.MarshalizedDataToBroadcastCalled(header, body)
}

func (blProcMock BlockProcessorMock) DecodeBlockBody(dta []byte) data.BodyHandler {
	return blProcMock.DecodeBlockBodyCalled(dta)
}

func (blProcMock BlockProcessorMock) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	return blProcMock.DecodeBlockHeaderCalled(dta)
}

func (blProcMock BlockProcessorMock) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	blProcMock.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

func (blProcMock BlockProcessorMock) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (blProcMock *BlockProcessorMock) IsInterfaceNil() bool {
	if blProcMock == nil {
		return true
	}
	return false
}
