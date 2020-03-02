package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
)

// BlockProcessorMock -
type BlockProcessorMock struct {
	ProcessBlockCalled                 func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                  func(header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled           func()
	CreateGenesisBlockCalled           func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockCalled                  func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled        func(header data.HeaderHandler, body data.BodyHandler) error
	SetOnRequestTransactionCalled      func(f func(destShardID uint32, txHash []byte))
	MarshalizedDataToBroadcastCalled   func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	AddLastNotarizedHdrCalled          func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled              func() data.HeaderHandler
	UpdateEpochStartTriggerRoundCalled func(round uint64)
	PruneStateOnRollbackCalled         func(currHeader data.HeaderHandler, prevHeader data.HeaderHandler)
	RevertStateToBlockCalled           func(header data.HeaderHandler) error
}

// ApplyProcessedMiniBlocks -
func (bpm *BlockProcessorMock) ApplyProcessedMiniBlocks(*processedMb.ProcessedMiniBlockTracker) {
}

// RestoreLastNotarizedHrdsToGenesis -
func (bpm *BlockProcessorMock) RestoreLastNotarizedHrdsToGenesis() {
}

// SetNumProcessedObj -
func (bpm *BlockProcessorMock) SetNumProcessedObj(_ uint64) {
}

// ProcessBlock -
func (bpm *BlockProcessorMock) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bpm.ProcessBlockCalled(header, body, haveTime)
}

// CommitBlock -
func (bpm *BlockProcessorMock) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.CommitBlockCalled(header, body)
}

// RevertAccountState -
func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

// CreateNewHeader -
func (bpm *BlockProcessorMock) CreateNewHeader() data.HeaderHandler {
	return bpm.CreateNewHeaderCalled()
}

// UpdateEpochStartTriggerRound -
func (bpm *BlockProcessorMock) UpdateEpochStartTriggerRound(round uint64) {
	if bpm.UpdateEpochStartTriggerRoundCalled != nil {
		bpm.UpdateEpochStartTriggerRoundCalled(round)
	}
}

// CreateGenesisBlock -
func (bpm *BlockProcessorMock) CreateGenesisBlock(balances map[string]*big.Int) (data.HeaderHandler, error) {
	return bpm.CreateGenesisBlockCalled(balances)
}

// CreateBlock -
func (bpm *BlockProcessorMock) CreateBlock(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	return bpm.CreateBlockCalled(initialHdrData, haveTime)
}

// RestoreBlockIntoPools -
func (bpm *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.RestoreBlockIntoPoolsCalled(header, body)
}

// MarshalizedDataToBroadcast -
func (bpm *BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bpm.MarshalizedDataToBroadcastCalled(header, body)
}

// AddLastNotarizedHdr -
func (bpm *BlockProcessorMock) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bpm.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

// RevertStateToBlock recreates thee state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) RevertStateToBlock(header data.HeaderHandler) error {
	if bpm.RevertStateToBlockCalled != nil {
		return bpm.RevertStateToBlockCalled(header)
	}

	return nil
}

// PruneStateOnRollback recreates thee state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) PruneStateOnRollback(currHeader data.HeaderHandler, prevHeader data.HeaderHandler) {
	if bpm.PruneStateOnRollbackCalled != nil {
		bpm.PruneStateOnRollbackCalled(currHeader, prevHeader)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpm *BlockProcessorMock) IsInterfaceNil() bool {
	return bpm == nil
}
