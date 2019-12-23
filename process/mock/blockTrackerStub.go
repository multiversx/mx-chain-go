package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	AddTrackedHeaderCalled                    func(header data.HeaderHandler, hash []byte)
	AddCrossNotarizedHeaderCalled             func(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeaderCalled              func(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	CleanupHeadersForShardBehindNonceCalled   func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                 func(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	DisplayTrackedHeadersCalled               func()
	GetLastCrossNotarizedHeaderCalled         func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetTrackedHeadersForShardCalled           func(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersForShardWithNonceCalled  func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuckCalled                        func(shardId uint32) bool
	LastHeaderForShardCalled                  func(shardId uint32) data.HeaderHandler
	RegisterSelfNotarizedHeadersHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastCrossNotarizedHeaderCalled      func()
	RestoreHeadersToGenesisCalled             func()
}

func (bts *BlockTrackerStub) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	if bts.AddTrackedHeaderCalled != nil {
		bts.AddTrackedHeaderCalled(header, hash)
	}
}

func (bts *BlockTrackerStub) AddCrossNotarizedHeader(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte) {
	if bts.AddCrossNotarizedHeaderCalled != nil {
		bts.AddCrossNotarizedHeaderCalled(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
	}
}

func (bts *BlockTrackerStub) AddSelfNotarizedHeader(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte) {
	if bts.AddSelfNotarizedHeaderCalled != nil {
		bts.AddSelfNotarizedHeaderCalled(shardID, selfNotarizedHeader, selfNotarizedHeaderHash)
	}
}

func (bts *BlockTrackerStub) CleanupHeadersForShardBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64) {
	if bts.CleanupHeadersForShardBehindNonceCalled != nil {
		bts.CleanupHeadersForShardBehindNonceCalled(shardID, selfNotarizedNonce, crossNotarizedNonce)
	}
}

func (bts *BlockTrackerStub) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	if bts.ComputeLongestChainCalled != nil {
		return bts.ComputeLongestChainCalled(shardID, header)
	}
	return nil, nil
}

func (bts *BlockTrackerStub) DisplayTrackedHeaders() {
	if bts.DisplayTrackedHeadersCalled != nil {
		bts.DisplayTrackedHeadersCalled()
	}
}

func (bts *BlockTrackerStub) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bts.GetLastCrossNotarizedHeaderCalled != nil {
		return bts.GetLastCrossNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

func (bts *BlockTrackerStub) GetTrackedHeadersForShard(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersForShardCalled != nil {
		return bts.GetTrackedHeadersForShardCalled(shardID)
	}

	return nil, nil
}

func (bts *BlockTrackerStub) GetTrackedHeadersForShardWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersForShardWithNonceCalled != nil {
		return bts.GetTrackedHeadersForShardWithNonceCalled(shardID, nonce)
	}

	return nil, nil
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	return bts.IsShardStuckCalled(shardId)
}

func (bts *BlockTrackerStub) LastHeaderForShard(shardId uint32) data.HeaderHandler {
	return bts.LastHeaderForShardCalled(shardId)
}

func (bts *BlockTrackerStub) RegisterSelfNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterSelfNotarizedHeadersHandlerCalled != nil {
		bts.RegisterSelfNotarizedHeadersHandlerCalled(handler)
	}
}

func (bts *BlockTrackerStub) RemoveLastCrossNotarizedHeader() {
	if bts.RemoveLastCrossNotarizedHeaderCalled != nil {
		bts.RemoveLastCrossNotarizedHeaderCalled()
	}
}

func (bts *BlockTrackerStub) RestoreHeadersToGenesis() {
	if bts.RestoreHeadersToGenesisCalled != nil {
		bts.RestoreHeadersToGenesisCalled()
	}
}

func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
