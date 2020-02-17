package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorMock struct {
	NrCommitBlockCalled                     uint32
	Marshalizer                             marshal.Marshalizer
	ProcessBlockCalled                      func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                       func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled                func()
	CreateBlockCalled                       func(blockChain data.ChainHandler, initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled             func(header data.HeaderHandler, body data.BodyHandler) error
	ApplyBodyToHeaderCalled                 func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error)
	MarshalizedDataToBroadcastCalled        func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyAndHeaderCalled          func(dta []byte) (data.BodyHandler, data.HeaderHandler)
	DecodeBlockBodyCalled                   func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled                 func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled               func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled                   func() data.HeaderHandler
	RevertStateToBlockCalled                func(header data.HeaderHandler) error
	RestoreLastNotarizedHrdsToGenesisCalled func()
}

// RestoreLastNotarizedHrdsToGenesis -
func (bpm *BlockProcessorMock) RestoreLastNotarizedHrdsToGenesis() {
	if bpm.RestoreLastNotarizedHrdsToGenesisCalled != nil {
		bpm.RestoreLastNotarizedHrdsToGenesisCalled()
	}
}

// ProcessBlock mocks pocessing a block
func (bpm *BlockProcessorMock) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bpm.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// ApplyProcessedMiniBlocks -
func (bpm *BlockProcessorMock) ApplyProcessedMiniBlocks(_ *processedMb.ProcessedMiniBlockTracker) {
}

// CommitBlock mocks the commit of a block
func (bpm *BlockProcessorMock) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.CommitBlockCalled(blockChain, header, body)
}

// RevertAccountState mocks revert of the accounts state
func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

// CreateNewHeader -
func (bpm *BlockProcessorMock) CreateNewHeader() data.HeaderHandler {
	return bpm.CreateNewHeaderCalled()
}

// CreateBlockBody -
func (bpm *BlockProcessorMock) CreateBlockBody(blockChain data.ChainHandler, initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return bpm.CreateBlockCalled(blockChain, initialHdrData, haveTime)
}

// RestoreBlockIntoPools -
func (bpm *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.RestoreBlockIntoPoolsCalled(header, body)
}

// ApplyBodyToHeader -
func (bpm *BlockProcessorMock) ApplyBodyToHeader(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error) {
	return bpm.ApplyBodyToHeaderCalled(blockChain, header, body)
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) RevertStateToBlock(header data.HeaderHandler) error {
	if bpm.RevertStateToBlockCalled != nil {
		return bpm.RevertStateToBlockCalled(header)
	}

	return nil
}

// SetNumProcessedObj -
func (bpm *BlockProcessorMock) SetNumProcessedObj(_ uint64) {
}

// MarshalizedDataToBroadcast -
func (bpm *BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bpm.MarshalizedDataToBroadcastCalled(header, body)
}

// DecodeBlockBodyAndHeader method decodes block body and header from a given byte array
func (bpm *BlockProcessorMock) DecodeBlockBodyAndHeader(dta []byte) (data.BodyHandler, data.HeaderHandler) {
	if dta == nil {
		return nil, nil
	}

	var marshalizedBodyAndHeader data.MarshalizedBodyAndHeader
	err := bpm.Marshalizer.Unmarshal(&marshalizedBodyAndHeader, dta)
	if err != nil {
		return nil, nil
	}

	var body block.Body
	err = bpm.Marshalizer.Unmarshal(&body, marshalizedBodyAndHeader.Body)
	if err != nil {
		return nil, nil
	}

	var header block.Header
	err = bpm.Marshalizer.Unmarshal(&header, marshalizedBodyAndHeader.Header)
	if err != nil {
		return nil, nil
	}

	return body, &header
}

// DecodeBlockBody method decodes block body from a given byte array
func (bpm *BlockProcessorMock) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return nil
	}

	var body block.Body

	err := bpm.Marshalizer.Unmarshal(&body, dta)
	if err != nil {
		return nil
	}

	return body
}

// DecodeBlockHeader method decodes block header from a given byte array
func (bpm *BlockProcessorMock) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	var header block.Header

	err := bpm.Marshalizer.Unmarshal(&header, dta)
	if err != nil {
		return nil
	}

	return &header
}

// AddLastNotarizedHdr -
func (bpm *BlockProcessorMock) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bpm.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpm *BlockProcessorMock) IsInterfaceNil() bool {
	return bpm == nil
}
