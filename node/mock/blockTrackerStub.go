package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// BlockTrackerStub -
type BlockTrackerStub struct {
	AddTrackedHeaderCalled                            func(header data.HeaderHandler, hash []byte)
	AddCrossNotarizedHeaderCalled                     func(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeaderCalled                      func(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	CheckBlockAgainstRounderCalled                    func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstFinalCalled                      func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstWhitelistCalled                  func(interceptedData process.InterceptedData) bool
	CleanupHeadersBehindNonceCalled                   func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                         func(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	ComputeLongestMetaChainFromLastNotarizedCalled    func() ([]data.HeaderHandler, [][]byte, error)
	ComputeLongestShardsChainsFromLastNotarizedCalled func() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error)
	DisplayTrackedHeadersCalled                       func()
	GetCrossNotarizedHeaderCalled                     func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeaderCalled                 func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeadersForAllShardsCalled    func() (map[uint32]data.HeaderHandler, error)
	GetLastSelfNotarizedHeaderCalled                  func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetSelfNotarizedHeaderCalled                      func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetTrackedHeadersCalled                           func(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersForAllShardsCalled               func() map[uint32][]data.HeaderHandler
	GetTrackedHeadersWithNonceCalled                  func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuckCalled                                func(shardId uint32) bool
	RegisterCrossNotarizedHeadersHandlerCalled        func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandlerCalled         func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterFinalMetachainHeadersHandlerCalled        func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastNotarizedHeadersCalled                  func()
	RestoreToGenesisCalled                            func()
	ShouldAddHeaderCalled                             func(headerHandler data.HeaderHandler) bool
}

// AddTrackedHeader -
func (bts *BlockTrackerStub) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	if bts.AddTrackedHeaderCalled != nil {
		bts.AddTrackedHeaderCalled(header, hash)
	}
}

// AddCrossNotarizedHeader -
func (bts *BlockTrackerStub) AddCrossNotarizedHeader(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte) {
	if bts.AddCrossNotarizedHeaderCalled != nil {
		bts.AddCrossNotarizedHeaderCalled(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
	}
}

// AddSelfNotarizedHeader -
func (bts *BlockTrackerStub) AddSelfNotarizedHeader(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte) {
	if bts.AddSelfNotarizedHeaderCalled != nil {
		bts.AddSelfNotarizedHeaderCalled(shardID, selfNotarizedHeader, selfNotarizedHeaderHash)
	}
}

// CheckBlockAgainstRounder -
func (bts *BlockTrackerStub) CheckBlockAgainstRounder(headerHandler data.HeaderHandler) error {
	if bts.CheckBlockAgainstRounderCalled != nil {
		return bts.CheckBlockAgainstRounderCalled(headerHandler)
	}

	return nil
}

// CheckBlockAgainstFinal -
func (bts *BlockTrackerStub) CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error {
	if bts.CheckBlockAgainstFinalCalled != nil {
		return bts.CheckBlockAgainstFinalCalled(headerHandler)
	}

	return nil
}

// CheckBlockAgainstWhitelist -
func (bts *BlockTrackerStub) CheckBlockAgainstWhitelist(interceptedData process.InterceptedData) bool {
	if bts.CheckBlockAgainstWhitelistCalled != nil {
		return bts.CheckBlockAgainstWhitelistCalled(interceptedData)
	}

	return false
}

// CleanupHeadersBehindNonce -
func (bts *BlockTrackerStub) CleanupHeadersBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64) {
	if bts.CleanupHeadersBehindNonceCalled != nil {
		bts.CleanupHeadersBehindNonceCalled(shardID, selfNotarizedNonce, crossNotarizedNonce)
	}
}

// CleanupInvalidCrossHeaders -
func (bts *BlockTrackerStub) CleanupInvalidCrossHeaders(_ uint32, _ uint64) {
}

// ComputeLongestChain -
func (bts *BlockTrackerStub) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	if bts.ComputeLongestChainCalled != nil {
		return bts.ComputeLongestChainCalled(shardID, header)
	}
	return nil, nil
}

// ComputeLongestMetaChainFromLastNotarized -
func (bts *BlockTrackerStub) ComputeLongestMetaChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	if bts.ComputeLongestMetaChainFromLastNotarizedCalled != nil {
		return bts.ComputeLongestMetaChainFromLastNotarizedCalled()
	}

	return nil, nil, nil
}

// ComputeLongestShardsChainsFromLastNotarized -
func (bts *BlockTrackerStub) ComputeLongestShardsChainsFromLastNotarized() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
	if bts.ComputeLongestShardsChainsFromLastNotarizedCalled != nil {
		return bts.ComputeLongestShardsChainsFromLastNotarizedCalled()
	}

	return nil, nil, nil, nil
}

// DisplayTrackedHeaders -
func (bts *BlockTrackerStub) DisplayTrackedHeaders() {
	if bts.DisplayTrackedHeadersCalled != nil {
		bts.DisplayTrackedHeadersCalled()
	}
}

// GetCrossNotarizedHeader -
func (bts *BlockTrackerStub) GetCrossNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if bts.GetCrossNotarizedHeaderCalled != nil {
		return bts.GetCrossNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

// GetLastCrossNotarizedHeader -
func (bts *BlockTrackerStub) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bts.GetLastCrossNotarizedHeaderCalled != nil {
		return bts.GetLastCrossNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

// GetLastCrossNotarizedHeadersForAllShards -
func (bts *BlockTrackerStub) GetLastCrossNotarizedHeadersForAllShards() (map[uint32]data.HeaderHandler, error) {
	if bts.GetLastCrossNotarizedHeadersForAllShardsCalled != nil {
		return bts.GetLastCrossNotarizedHeadersForAllShardsCalled()
	}

	return nil, nil
}

// GetLastSelfNotarizedHeader -
func (bts *BlockTrackerStub) GetLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bts.GetLastSelfNotarizedHeaderCalled != nil {
		return bts.GetLastSelfNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

// GetSelfNotarizedHeader -
func (bts *BlockTrackerStub) GetSelfNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if bts.GetCrossNotarizedHeaderCalled != nil {
		return bts.GetCrossNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

// GetTrackedHeaders -
func (bts *BlockTrackerStub) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersCalled != nil {
		return bts.GetTrackedHeadersCalled(shardID)
	}

	return nil, nil
}

// GetTrackedHeadersForAllShards -
func (bts *BlockTrackerStub) GetTrackedHeadersForAllShards() map[uint32][]data.HeaderHandler {
	if bts.GetTrackedHeadersForAllShardsCalled != nil {
		return bts.GetTrackedHeadersForAllShardsCalled()
	}

	return nil
}

// GetTrackedHeadersWithNonce -
func (bts *BlockTrackerStub) GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if bts.GetTrackedHeadersWithNonceCalled != nil {
		return bts.GetTrackedHeadersWithNonceCalled(shardID, nonce)
	}

	return nil, nil
}

// IsShardStuck -
func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	if bts.IsShardStuckCalled != nil {
		return bts.IsShardStuckCalled(shardId)
	}

	return false
}

// RegisterCrossNotarizedHeadersHandler -
func (bts *BlockTrackerStub) RegisterCrossNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterCrossNotarizedHeadersHandlerCalled != nil {
		bts.RegisterCrossNotarizedHeadersHandlerCalled(handler)
	}
}

// RegisterSelfNotarizedHeadersHandler -
func (bts *BlockTrackerStub) RegisterSelfNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterSelfNotarizedHeadersHandlerCalled != nil {
		bts.RegisterSelfNotarizedHeadersHandlerCalled(handler)
	}
}

// RegisterFinalMetachainHeadersHandler -
func (bts *BlockTrackerStub) RegisterFinalMetachainHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bts.RegisterFinalMetachainHeadersHandlerCalled != nil {
		bts.RegisterFinalMetachainHeadersHandlerCalled(handler)
	}
}

// RemoveLastNotarizedHeaders -
func (bts *BlockTrackerStub) RemoveLastNotarizedHeaders() {
	if bts.RemoveLastNotarizedHeadersCalled != nil {
		bts.RemoveLastNotarizedHeadersCalled()
	}
}

// RestoreToGenesis -
func (bts *BlockTrackerStub) RestoreToGenesis() {
	if bts.RestoreToGenesisCalled != nil {
		bts.RestoreToGenesisCalled()
	}
}

// ShouldAddHeader -
func (bts *BlockTrackerStub) ShouldAddHeader(headerHandler data.HeaderHandler) bool {
	if bts.ShouldAddHeaderCalled != nil {
		return bts.ShouldAddHeaderCalled(headerHandler)
	}

	return true
}

// IsInterfaceNil -
func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
