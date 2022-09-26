package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

// BlockProcessorMock mocks the implementation for a blockProcessor
type BlockProcessorMock struct {
	NrCommitBlockCalled                     uint32
	Marshalizer                             marshal.Marshalizer
	ProcessBlockCalled                      func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	ProcessScheduledBlockCalled             func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                       func(header data.HeaderHandler, body data.BodyHandler) error
	RevertCurrentBlockCalled                func()
	CreateBlockCalled                       func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled             func(header data.HeaderHandler, body data.BodyHandler) error
	RestoreBlockBodyIntoPoolsCalled         func(body data.BodyHandler) error
	MarshalizedDataToBroadcastCalled        func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyCalled                   func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled                 func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled               func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled                   func(round uint64, nonce uint64) (data.HeaderHandler, error)
	PruneStateOnRollbackCalled              func(currHeader data.HeaderHandler, currHeaderHash []byte, prevHeader data.HeaderHandler, prevHeaderHash []byte)
	RestoreLastNotarizedHrdsToGenesisCalled func()
	RevertStateToBlockCalled                func(header data.HeaderHandler, rootHash []byte) error
	RevertIndexedBlockCalled                func(header data.HeaderHandler)
}

// RestoreLastNotarizedHrdsToGenesis -
func (bpm *BlockProcessorMock) RestoreLastNotarizedHrdsToGenesis() {
	if bpm.RestoreLastNotarizedHrdsToGenesisCalled != nil {
		bpm.RestoreLastNotarizedHrdsToGenesisCalled()
	}
}

// ProcessBlock mocks processing a block
func (bpm *BlockProcessorMock) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	if bpm.ProcessBlockCalled != nil {
		return bpm.ProcessBlockCalled(header, body, haveTime)
	}

	return nil
}

// ProcessScheduledBlock mocks processing a scheduled block
func (bpm *BlockProcessorMock) ProcessScheduledBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	if bpm.ProcessScheduledBlockCalled != nil {
		return bpm.ProcessScheduledBlockCalled(header, body, haveTime)
	}

	return nil
}

// CommitBlock mocks the commit of a block
func (bpm *BlockProcessorMock) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	if bpm.CommitBlockCalled != nil {
		return bpm.CommitBlockCalled(header, body)
	}

	return nil
}

// RevertCurrentBlock mocks revert of the current block
func (bpm *BlockProcessorMock) RevertCurrentBlock() {
	if bpm.RevertCurrentBlockCalled != nil {
		bpm.RevertCurrentBlockCalled()
	}
}

// CreateNewHeader -
func (bpm *BlockProcessorMock) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	if bpm.CreateNewHeaderCalled != nil {
		return bpm.CreateNewHeaderCalled(round, nonce)
	}

	return nil, nil
}

// CreateBlock -
func (bpm *BlockProcessorMock) CreateBlock(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	if bpm.CreateBlockCalled != nil {
		return bpm.CreateBlockCalled(initialHdrData, haveTime)
	}

	return nil, nil, nil
}

// RestoreBlockIntoPools -
func (bpm *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	if bpm.RestoreBlockIntoPoolsCalled != nil {
		return bpm.RestoreBlockIntoPoolsCalled(header, body)
	}

	return nil
}

// RestoreBlockBodyIntoPools -
func (bpm *BlockProcessorMock) RestoreBlockBodyIntoPools(body data.BodyHandler) error {
	if bpm.RestoreBlockBodyIntoPoolsCalled != nil {
		return bpm.RestoreBlockBodyIntoPoolsCalled(body)
	}
	return nil
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	if bpm.RevertStateToBlockCalled != nil {
		return bpm.RevertStateToBlockCalled(header, rootHash)
	}
	return nil
}

// PruneStateOnRollback recreates the state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) PruneStateOnRollback(currHeader data.HeaderHandler, currHeaderHash []byte, prevHeader data.HeaderHandler, prevHeaderHash []byte) {
	if bpm.PruneStateOnRollbackCalled != nil {
		bpm.PruneStateOnRollbackCalled(currHeader, currHeaderHash, prevHeader, prevHeaderHash)
	}
}

// SetNumProcessedObj -
func (bpm *BlockProcessorMock) SetNumProcessedObj(_ uint64) {
}

// MarshalizedDataToBroadcast -
func (bpm *BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bpm.MarshalizedDataToBroadcastCalled(header, body)
}

// DecodeBlockBody method decodes block body from a given byte array
func (bpm *BlockProcessorMock) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return &block.Body{}
	}

	var body block.Body

	err := bpm.Marshalizer.Unmarshal(&body, dta)
	if err != nil {
		return nil
	}

	return &body
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

// Close -
func (bpm *BlockProcessorMock) Close() error {
	return nil
}

// RevertIndexedBlock -
func (bpm *BlockProcessorMock) RevertIndexedBlock(header data.HeaderHandler) {
	if bpm.RevertIndexedBlockCalled != nil {
		bpm.RevertIndexedBlockCalled(header)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpm *BlockProcessorMock) IsInterfaceNil() bool {
	return bpm == nil
}
