package testscommon

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

// BlockProcessorStub mocks the implementation for a blockProcessor
type BlockProcessorStub struct {
	SetNumProcessedObjCalled         func(numObj uint64)
	ProcessBlockCalled               func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error)
	ProcessScheduledBlockCalled      func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(header data.HeaderHandler, body data.BodyHandler) error
	RevertCurrentBlockCalled         func()
	PruneStateOnRollbackCalled       func(currHeader data.HeaderHandler, currHeaderHash []byte, prevHeader data.HeaderHandler, prevHeaderHash []byte)
	CreateBlockCalled                func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	RestoreBlockBodyIntoPoolsCalled  func(body data.BodyHandler) error
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	CreateNewHeaderCalled            func(round uint64, nonce uint64) (data.HeaderHandler, error)
	RevertStateToBlockCalled         func(header data.HeaderHandler, rootHash []byte) error
	NonceOfFirstCommittedBlockCalled func() core.OptionalUint64
	SetProcessDebuggerCalled         func(debugger process.Debugger)
	CloseCalled                      func() error
}

// SetNumProcessedObj -
func (bps *BlockProcessorStub) SetNumProcessedObj(numObj uint64) {
	if bps.SetNumProcessedObjCalled != nil {
		bps.SetNumProcessedObjCalled(numObj)
	}
}

// ProcessBlock mocks processing a block
func (bps *BlockProcessorStub) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
	if bps.ProcessBlockCalled != nil {
		return bps.ProcessBlockCalled(header, body, haveTime)
	}

	return header, body, nil
}

// ProcessScheduledBlock mocks processing a scheduled block
func (bps *BlockProcessorStub) ProcessScheduledBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	if bps.ProcessScheduledBlockCalled != nil {
		return bps.ProcessScheduledBlockCalled(header, body, haveTime)
	}

	return nil
}

// CommitBlock mocks the commit of a block
func (bps *BlockProcessorStub) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	if bps.CommitBlockCalled != nil {
		return bps.CommitBlockCalled(header, body)
	}

	return nil
}

// RevertCurrentBlock mocks revert of the current block
func (bps *BlockProcessorStub) RevertCurrentBlock() {
	if bps.RevertCurrentBlockCalled != nil {
		bps.RevertCurrentBlockCalled()
	}
}

// PruneStateOnRollback recreates the state tries to the root hashes indicated by the provided header
func (bps *BlockProcessorStub) PruneStateOnRollback(currHeader data.HeaderHandler, currHeaderHash []byte, prevHeader data.HeaderHandler, prevHeaderHash []byte) {
	if bps.PruneStateOnRollbackCalled != nil {
		bps.PruneStateOnRollbackCalled(currHeader, currHeaderHash, prevHeader, prevHeaderHash)
	}
}

// CreateBlock mocks the creation of a new block with header and body
func (bps *BlockProcessorStub) CreateBlock(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	if bps.CreateBlockCalled != nil {
		return bps.CreateBlockCalled(initialHdrData, haveTime)
	}

	return nil, nil, ErrNotImplemented
}

// RestoreBlockIntoPools -
func (bps *BlockProcessorStub) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	if bps.RestoreBlockIntoPoolsCalled != nil {
		return bps.RestoreBlockIntoPoolsCalled(header, body)
	}

	return nil
}

// RestoreBlockBodyIntoPools -
func (bps *BlockProcessorStub) RestoreBlockBodyIntoPools(body data.BodyHandler) error {
	if bps.RestoreBlockBodyIntoPoolsCalled != nil {
		return bps.RestoreBlockBodyIntoPoolsCalled(body)
	}

	return nil
}

// MarshalizedDataToBroadcast -
func (bps *BlockProcessorStub) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	if bps.MarshalizedDataToBroadcastCalled != nil {
		return bps.MarshalizedDataToBroadcastCalled(header, body)
	}

	return nil, nil, ErrNotImplemented
}

// DecodeBlockBody -
func (bps *BlockProcessorStub) DecodeBlockBody(dta []byte) data.BodyHandler {
	if bps.DecodeBlockBodyCalled != nil {
		return bps.DecodeBlockBodyCalled(dta)
	}

	return nil
}

// DecodeBlockHeader -
func (bps *BlockProcessorStub) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if bps.DecodeBlockHeaderCalled != nil {
		return bps.DecodeBlockHeaderCalled(dta)
	}

	return nil
}

// CreateNewHeader creates a new header
func (bps *BlockProcessorStub) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	if bps.CreateNewHeaderCalled != nil {
		return bps.CreateNewHeaderCalled(round, nonce)
	}

	return nil, ErrNotImplemented
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided header
func (bps *BlockProcessorStub) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	if bps.RevertStateToBlockCalled != nil {
		return bps.RevertStateToBlockCalled(header, rootHash)
	}

	return nil
}

// NonceOfFirstCommittedBlock -
func (bps *BlockProcessorStub) NonceOfFirstCommittedBlock() core.OptionalUint64 {
	if bps.NonceOfFirstCommittedBlockCalled != nil {
		return bps.NonceOfFirstCommittedBlockCalled()
	}

	return core.OptionalUint64{
		HasValue: false,
	}
}

// SetProcessDebugger -
func (bps *BlockProcessorStub) SetProcessDebugger(debugger process.Debugger) error {
	if bps.SetProcessDebuggerCalled != nil {
		bps.SetProcessDebuggerCalled(debugger)
	}

	return nil
}

// Close -
func (bps *BlockProcessorStub) Close() error {
	if bps.CloseCalled != nil {
		return bps.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bps *BlockProcessorStub) IsInterfaceNil() bool {
	return bps == nil
}
