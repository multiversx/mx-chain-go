package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
)

// BlockProcessorStub mocks the implementation for a blockProcessor
type BlockProcessorStub struct {
	ProcessBlockCalled               func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func(header data.HeaderHandler)
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockCalled                func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled        func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled            func(round uint64, nonce uint64) data.HeaderHandler
	PruneStateOnRollbackCalled       func(currHeader data.HeaderHandler, prevHeader data.HeaderHandler)
	RevertStateToBlockCalled         func(header data.HeaderHandler) error
	RevertIndexedBlockCalled         func(header data.HeaderHandler)
}

// RestoreLastNotarizedHrdsToGenesis -
func (bps *BlockProcessorStub) RestoreLastNotarizedHrdsToGenesis() {
}

// SetNumProcessedObj -
func (bps *BlockProcessorStub) SetNumProcessedObj(_ uint64) {
}

// ProcessBlock mocks pocessing a block
func (bps *BlockProcessorStub) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bps.ProcessBlockCalled(header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (bps *BlockProcessorStub) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	return bps.CommitBlockCalled(header, body)
}

// RevertAccountState mocks revert of the accounts state
func (bps *BlockProcessorStub) RevertAccountState(header data.HeaderHandler) {
	bps.RevertAccountStateCalled(header)
}

// CreateGenesisBlock mocks the creation of a genesis block body
func (bps *BlockProcessorStub) CreateGenesisBlock(balances map[string]*big.Int) (data.HeaderHandler, error) {
	return bps.CreateGenesisBlockCalled(balances)
}

// PruneStateOnRollback recreates thee state tries to the root hashes indicated by the provided header
func (bps *BlockProcessorStub) PruneStateOnRollback(currHeader data.HeaderHandler, prevHeader data.HeaderHandler) {
	if bps.PruneStateOnRollbackCalled != nil {
		bps.PruneStateOnRollbackCalled(currHeader, prevHeader)
	}
}

// CreateBlock mocks the creation of a new block with header and body
func (bps *BlockProcessorStub) CreateBlock(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	return bps.CreateBlockCalled(initialHdrData, haveTime)
}

// RestoreBlockIntoPools -
func (bps *BlockProcessorStub) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return bps.RestoreBlockIntoPoolsCalled(header, body)
}

// MarshalizedDataToBroadcast -
func (bps *BlockProcessorStub) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bps.MarshalizedDataToBroadcastCalled(header, body)
}

// DecodeBlockBody -
func (bps *BlockProcessorStub) DecodeBlockBody(dta []byte) data.BodyHandler {
	return bps.DecodeBlockBodyCalled(dta)
}

// DecodeBlockHeader -
func (bps *BlockProcessorStub) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	return bps.DecodeBlockHeaderCalled(dta)
}

// AddLastNotarizedHdr -
func (bps *BlockProcessorStub) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bps.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

// CreateNewHeader creates a new header
func (bps *BlockProcessorStub) CreateNewHeader(round uint64, nonce uint64) data.HeaderHandler {
	return bps.CreateNewHeaderCalled(round, nonce)
}

// ApplyProcessedMiniBlocks -
func (bps *BlockProcessorStub) ApplyProcessedMiniBlocks(_ *processedMb.ProcessedMiniBlockTracker) {
}

// Close -
func (bps *BlockProcessorStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bps *BlockProcessorStub) IsInterfaceNil() bool {
	return bps == nil
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided header
func (bps *BlockProcessorStub) RevertStateToBlock(header data.HeaderHandler) error {
	if bps.RevertStateToBlockCalled != nil {
		return bps.RevertStateToBlockCalled(header)
	}

	return nil
}

// RevertIndexedBlock -
func (bps *BlockProcessorStub) RevertIndexedBlock(header data.HeaderHandler) {
	if bps.RevertIndexedBlockCalled != nil {
		bps.RevertIndexedBlockCalled(header)
	}
}
