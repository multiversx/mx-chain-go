package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	CleanupHeadersForShardBehindNonceCalled   func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                 func(shardID uint32, header data.HeaderHandler) []data.HeaderHandler
	IsShardStuckCalled                        func(shardId uint32) bool
	LastHeaderForShardCalled                  func(shardId uint32) data.HeaderHandler
	RegisterSelfNotarizedHeadersHandlerCalled func(handler func(headers []data.HeaderHandler, headersHashes [][]byte))
}

func (bts *BlockTrackerStub) CleanupHeadersForShardBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64) {
	if bts.CleanupHeadersForShardBehindNonceCalled != nil {
		bts.CleanupHeadersForShardBehindNonceCalled(shardID, selfNotarizedNonce, crossNotarizedNonce)
	}
}

func (bts *BlockTrackerStub) ComputeLongestChain(shardID uint32, header data.HeaderHandler) []data.HeaderHandler {
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

func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
