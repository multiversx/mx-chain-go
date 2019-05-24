package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorMock struct {
	NrCommitBlockCalled              uint32
	Marshalizer                      marshal.Marshalizer
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	CreateBlockCalled                func(round int32, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	CreateBlockHeaderCalled          func(body data.BodyHandler, round int32, haveTime func() bool) (data.HeaderHandler, error)
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	GetUnnotarisedHeadersCalled      func(blockChain data.ChainHandler) []data.HeaderHandler
	SetBroadcastRoundCalled          func(nonce uint64, round int32)
	GetBroadcastRoundCalled          func(nonce uint64) int32
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

// CreateTxBlockBody mocks the creation of a transaction block body
func (blProcMock *BlockProcessorMock) CreateBlockBody(round int32, haveTime func() bool) (data.BodyHandler, error) {
	return blProcMock.CreateBlockCalled(round, haveTime)
}

func (blProcMock *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return blProcMock.RestoreBlockIntoPoolsCalled(header, body)
}

func (blProcMock BlockProcessorMock) CreateBlockHeader(body data.BodyHandler, round int32, haveTime func() bool) (data.HeaderHandler, error) {
	return blProcMock.CreateBlockHeaderCalled(body, round, haveTime)
}

func (blProcMock BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error) {
	return blProcMock.MarshalizedDataToBroadcastCalled(header, body)
}

// DecodeBlockBody method decodes block body from a given byte array
func (blProcMock BlockProcessorMock) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return nil
	}

	var body block.Body

	err := blProcMock.Marshalizer.Unmarshal(&body, dta)
	if err != nil {
		return nil
	}

	return body
}

// DecodeBlockHeader method decodes block header from a given byte array
func (blProcMock BlockProcessorMock) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	var header block.Header

	err := blProcMock.Marshalizer.Unmarshal(&header, dta)
	if err != nil {
		return nil
	}

	return &header
}

func (blProcMock BlockProcessorMock) GetUnnotarisedHeaders(blockChain data.ChainHandler) []data.HeaderHandler {
	return blProcMock.GetUnnotarisedHeadersCalled(blockChain)
}

func (blProcMock BlockProcessorMock) SetBroadcastRound(nonce uint64, round int32) {
	blProcMock.SetBroadcastRoundCalled(nonce, round)
}

func (blProcMock BlockProcessorMock) GetBroadcastRound(nonce uint64) int32 {
	return blProcMock.GetBroadcastRoundCalled(nonce)
}
