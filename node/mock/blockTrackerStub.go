package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	AddTrackedHeaderCalled                            func(header data.HeaderHandler, hash []byte)
	AddCrossNotarizedHeaderCalled                     func(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeaderCalled                      func(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	CheckBlockBasicValidityCalled                     func(headerHandler data.HeaderHandler) error
	CleanupHeadersBehindNonceCalled                   func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                         func(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	ComputeLongestMetaChainFromLastNotarizedCalled    func() ([]data.HeaderHandler, [][]byte, error)
	ComputeLongestShardsChainsFromLastNotarizedCalled func() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error)
	DisplayTrackedHeadersCalled                       func()
	GetCrossNotarizedHeaderCalled                     func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetFinalHeaderCalled                              func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeaderCalled                 func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeadersForAllShardsCalled    func() (map[uint32]data.HeaderHandler, error)
	GetTrackedHeadersCalled                           func(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersForAllShardsCalled               func() map[uint32][]data.HeaderHandler
	GetTrackedHeadersWithNonceCalled                  func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuckCalled                                func(shardId uint32) bool
	RegisterCrossNotarizedHeadersHandlerCalled        func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandlerCalled         func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastNotarizedHeadersCalled                  func()
	RestoreToGenesisCalled                            func()
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

func (bts *BlockTrackerStub) CheckBlockBasicValidity(headerHandler data.HeaderHandler) error {
	if bts.CheckBlockBasicValidityCalled != nil {
		return bts.CheckBlockBasicValidityCalled(headerHandler)
	}

	return nil
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

func (bts *BlockTrackerStub) ComputeLongestMetaChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	if bts.ComputeLongestMetaChainFromLastNotarizedCalled != nil {
		return bts.ComputeLongestMetaChainFromLastNotarizedCalled()
	}

	return nil, nil, nil
}

func (bts *BlockTrackerStub) ComputeLongestShardsChainsFromLastNotarized() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
	if bts.ComputeLongestShardsChainsFromLastNotarizedCalled != nil {
		return bts.ComputeLongestShardsChainsFromLastNotarizedCalled()
	}

	return nil, nil, nil, nil
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

func (bts *BlockTrackerStub) GetFinalHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bts.GetFinalHeaderCalled != nil {
		return bts.GetFinalHeaderCalled(shardID)
	}

	return nil, nil, nil
}

func (bts *BlockTrackerStub) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bts.GetLastCrossNotarizedHeaderCalled != nil {
		return bts.GetLastCrossNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

func (bts *BlockTrackerStub) GetLastCrossNotarizedHeadersForAllShards() (map[uint32]data.HeaderHandler, error) {
	if bts.GetLastCrossNotarizedHeadersForAllShardsCalled != nil {
		return bts.GetLastCrossNotarizedHeadersForAllShardsCalled()
	}

	return nil, nil
}

func (bts *BlockTrackerStub) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersCalled != nil {
		return bts.GetTrackedHeadersCalled(shardID)
	}

	return nil, nil
}

func (bts *BlockTrackerStub) GetTrackedHeadersForAllShards() map[uint32][]data.HeaderHandler {
	if bts.GetTrackedHeadersForAllShardsCalled != nil {
		return bts.GetTrackedHeadersForAllShardsCalled()
	}

	return nil
}

func (bts *BlockTrackerStub) GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersWithNonceCalled != nil {
		return bts.GetTrackedHeadersWithNonceCalled(shardID, nonce)
	}

	return nil, nil
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	if bts.IsShardStuckCalled != nil {
		return bts.IsShardStuckCalled(shardId)
	}

	return false
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

func (bts *BlockTrackerStub) RestoreToGenesis() {
	if bts.RestoreToGenesisCalled != nil {
		bts.RestoreToGenesisCalled()
	}
}

func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
