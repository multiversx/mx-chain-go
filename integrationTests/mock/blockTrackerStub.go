package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	AddTrackedHeaderCalled                     func(header data.HeaderHandler, hash []byte)
	AddCrossNotarizedHeaderCalled              func(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeaderCalled               func(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	CleanupHeadersBehindNonceCalled            func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                  func(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	DisplayTrackedHeadersCalled                func()
	GetCrossNotarizedHeaderCalled              func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeaderCalled          func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetTrackedHeadersCalled                    func(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersWithNonceCalled           func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuckCalled                         func(shardId uint32) bool
	GetLastHeaderCalled                        func(shardId uint32) data.HeaderHandler
	RegisterCrossNotarizedHeadersHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandlerCalled  func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastNotarizedHeadersCalled           func()
	RestoreHeadersToGenesisCalled              func()
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

func (bts *BlockTrackerStub) CleanupHeadersBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64) {
	if bts.CleanupHeadersBehindNonceCalled != nil {
		bts.CleanupHeadersBehindNonceCalled(shardID, selfNotarizedNonce, crossNotarizedNonce)
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

func (bts *BlockTrackerStub) GetCrossNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if bts.GetCrossNotarizedHeaderCalled != nil {
		return bts.GetCrossNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

func (bts *BlockTrackerStub) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bts.GetLastCrossNotarizedHeaderCalled != nil {
		return bts.GetLastCrossNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

func (bts *BlockTrackerStub) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersCalled != nil {
		return bts.GetTrackedHeadersCalled(shardID)
	}

	return nil, nil
}

func (bts *BlockTrackerStub) GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersWithNonceCalled != nil {
		return bts.GetTrackedHeadersWithNonceCalled(shardID, nonce)
	}

	return nil, nil
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	return bts.IsShardStuckCalled(shardId)
}

func (bts *BlockTrackerStub) GetLastHeader(shardId uint32) data.HeaderHandler {
	return bts.GetLastHeaderCalled(shardId)
}

func (bts *BlockTrackerStub) RegisterCrossNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterCrossNotarizedHeadersHandlerCalled != nil {
		bts.RegisterCrossNotarizedHeadersHandlerCalled(handler)
	}
}

func (bts *BlockTrackerStub) RegisterSelfNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterSelfNotarizedHeadersHandlerCalled != nil {
		bts.RegisterSelfNotarizedHeadersHandlerCalled(handler)
	}
}

func (bts *BlockTrackerStub) RemoveLastNotarizedHeaders() {
	if bts.RemoveLastNotarizedHeadersCalled != nil {
		bts.RemoveLastNotarizedHeadersCalled()
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
