package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

// BlockProcessorStub mocks the implementation for a blockProcessor
type BlockProcessorStub struct {
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockBodyCalled            func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	ApplyBodyToHeaderCalled          func(header data.HeaderHandler, body data.BodyHandler) error
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled        func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled            func() data.HeaderHandler
	RevertStateToBlockCalled         func(header data.HeaderHandler) error
	HeaderCopyCalled                 func(hdr data.HeaderHandler) data.HeaderHandler
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
func (blProcMock *BlockProcessorStub) CreateGenesisBlock(balances map[string]*big.Int) (data.HeaderHandler, error) {
	return blProcMock.CreateGenesisBlockCalled(balances)
}

// RevertStateToBlock recreates thee state tries to the root hashes indicated by the provided header
func (blProcMock *BlockProcessorStub) RevertStateToBlock(header data.HeaderHandler) error {
	if blProcMock.RevertStateToBlockCalled != nil {
		return blProcMock.RevertStateToBlock(header)
	}
	return nil
}

// CreateTxBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorStub) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockBodyCalled(initialHdrData, haveTime)
}

func (blProcMock *BlockProcessorStub) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.RestoreBlockIntoPoolsCalled(header, body)
}

func (blProcMock BlockProcessorStub) ApplyBodyToHeader(header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.ApplyBodyToHeaderCalled(header, body)
}

func (blProcMock BlockProcessorStub) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return blProcMock.MarshalizedDataToBroadcastCalled(header, body)
}

func (blProcMock BlockProcessorStub) DecodeBlockBody(dta []byte) data.BodyHandler {
	return blProcMock.DecodeBlockBodyCalled(dta)
}

func (blProcMock BlockProcessorStub) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	return blProcMock.DecodeBlockHeaderCalled(dta)
}

func (blProcMock BlockProcessorStub) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	blProcMock.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

func (blProcMock BlockProcessorStub) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
	panic("implement me")
}

// CreateNewHeader creates a new header
func (blProcMock BlockProcessorStub) CreateNewHeader() data.HeaderHandler {
	return blProcMock.CreateNewHeaderCalled()
}

func (blProcMock *BlockProcessorStub) HeaderCopy(hdr data.HeaderHandler) data.HeaderHandler {
	return blProcMock.HeaderCopyCalled(hdr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (blProcMock *BlockProcessorStub) IsInterfaceNil() bool {
	if blProcMock == nil {
		return true
	}
	return false
}
