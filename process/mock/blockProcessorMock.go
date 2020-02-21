package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
)

// BlockProcessorMock -
type BlockProcessorMock struct {
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockCalled                func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	ApplyBodyToHeaderCalled          func(header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error)
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyAndHeaderCalled   func(dta []byte) (data.BodyHandler, data.HeaderHandler)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled        func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled            func() data.HeaderHandler
	RevertStateCalled                func(currHeader data.HeaderHandler, prevHeader data.HeaderHandler) error
	RecreateStateTriesCalled         func(header data.HeaderHandler) error
}

// ApplyProcessedMiniBlocks -
func (bpm *BlockProcessorMock) ApplyProcessedMiniBlocks(*processedMb.ProcessedMiniBlockTracker) {
}

// RestoreLastNotarizedHrdsToGenesis -
func (bpm *BlockProcessorMock) RestoreLastNotarizedHrdsToGenesis() {
}

// SetNumProcessedObj -
func (bpm *BlockProcessorMock) SetNumProcessedObj(numObj uint64) {
}

// ProcessBlock -
func (bpm *BlockProcessorMock) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bpm.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// CommitBlock -
func (bpm *BlockProcessorMock) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.CommitBlockCalled(blockChain, header, body)
}

// RevertAccountState -
func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

// CreateNewHeader -
func (bpm *BlockProcessorMock) CreateNewHeader(_ uint64) data.HeaderHandler {
	return bpm.CreateNewHeaderCalled()
}

// CreateGenesisBlock -
func (bpm *BlockProcessorMock) CreateGenesisBlock(balances map[string]*big.Int) (data.HeaderHandler, error) {
	return bpm.CreateGenesisBlockCalled(balances)
}

// CreateBlockBody -
func (bpm *BlockProcessorMock) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return bpm.CreateBlockCalled(initialHdrData, haveTime)
}

// RestoreBlockIntoPools -
func (bpm *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.RestoreBlockIntoPoolsCalled(header, body)
}

// ApplyBodyToHeader -
func (bpm *BlockProcessorMock) ApplyBodyToHeader(header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error) {
	return bpm.ApplyBodyToHeaderCalled(header, body)
}

// MarshalizedDataToBroadcast -
func (bpm *BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bpm.MarshalizedDataToBroadcastCalled(header, body)
}

// DecodeBlockBodyAndHeader -
func (bpm *BlockProcessorMock) DecodeBlockBodyAndHeader(dta []byte) (data.BodyHandler, data.HeaderHandler) {
	return bpm.DecodeBlockBodyAndHeaderCalled(dta)
}

// DecodeBlockBody -
func (bpm *BlockProcessorMock) DecodeBlockBody(dta []byte) data.BodyHandler {
	return bpm.DecodeBlockBodyCalled(dta)
}

// DecodeBlockHeader -
func (bpm *BlockProcessorMock) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	return bpm.DecodeBlockHeaderCalled(dta)
}

// AddLastNotarizedHdr -
func (bpm *BlockProcessorMock) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bpm.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

// SetConsensusData -
func (bpm *BlockProcessorMock) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
	panic("implement me")
}

// RevertState recreates thee state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) RevertState(currHeader data.HeaderHandler, prevHeader data.HeaderHandler) error {
	if bpm.RevertStateCalled != nil {
		return bpm.RevertStateCalled(currHeader, prevHeader)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpm *BlockProcessorMock) IsInterfaceNil() bool {
	return bpm == nil
}

// RecreateStateTries recreates the state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) RecreateStateTries(header data.HeaderHandler) error {
	if bpm.RecreateStateTriesCalled != nil {
		return bpm.RecreateStateTriesCalled(header)
	}

	return nil
}
