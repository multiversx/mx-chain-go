package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	AddTrackedHeaderCalled                    func(header data.HeaderHandler, hash []byte)
	AddCrossNotarizedHeaderCalled             func(shardID uint32, crossNotarizedHeaders data.HeaderHandler)
	AddSelfNotarizedHeaderCalled              func(shardID uint32, selfNotarizedHeaders data.HeaderHandler)
	CleanupHeadersForShardBehindNonceCalled   func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                 func(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	IsShardStuckCalled                        func(shardId uint32) bool
	LastHeaderForShardCalled                  func(shardId uint32) data.HeaderHandler
	RegisterSelfNotarizedHeadersHandlerCalled func(handler func(headers []data.HeaderHandler, headersHashes [][]byte))
	RestoreHeadersToGenesisCalled             func()
}

func (bts *BlockTrackerStub) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	if bts.AddTrackedHeaderCalled != nil {
		bts.AddTrackedHeaderCalled(header, hash)
	}
}

func (bts *BlockTrackerStub) AddCrossNotarizedHeader(shardID uint32, crossNotarizedHeader data.HeaderHandler) {
	if bts.AddCrossNotarizedHeaderCalled != nil {
		bts.AddCrossNotarizedHeaderCalled(shardID, crossNotarizedHeader)
	}
}

func (bts *BlockTrackerStub) AddSelfNotarizedHeader(shardID uint32, selfNotarizedHeader data.HeaderHandler) {
	if bts.AddSelfNotarizedHeaderCalled != nil {
		bts.AddSelfNotarizedHeaderCalled(shardID, selfNotarizedHeader)
	}
}

func (bts *BlockTrackerStub) CleanupHeadersForShardBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64) {
	if bts.CleanupHeadersForShardBehindNonceCalled != nil {
		bts.CleanupHeadersForShardBehindNonceCalled(shardID, selfNotarizedNonce, crossNotarizedNonce)
	}
}

func (bts *BlockTrackerStub) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	return bts.ComputeLongestChainCalled(shardID, header)
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	return bts.IsShardStuckCalled(shardId)
}

func (bts *BlockTrackerStub) LastHeaderForShard(shardId uint32) data.HeaderHandler {
	return bts.LastHeaderForShardCalled(shardId)
}

func (bts *BlockTrackerStub) RegisterSelfNotarizedHeadersHandler(handler func(headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterSelfNotarizedHeadersHandlerCalled != nil {
		bts.RegisterSelfNotarizedHeadersHandlerCalled(handler)
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
