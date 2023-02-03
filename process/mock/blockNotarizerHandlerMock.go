package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// BlockNotarizerHandlerMock -
type BlockNotarizerHandlerMock struct {
	AddNotarizedHeaderCalled                 func(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte)
	CleanupNotarizedHeadersBehindNonceCalled func(shardID uint32, nonce uint64)
	DisplayNotarizedHeadersCalled            func(shardID uint32, message string)
	GetFirstNotarizedHeaderCalled            func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastNotarizedHeaderCalled             func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastNotarizedHeaderNonceCalled        func(shardID uint32) uint64
	GetNotarizedHeaderCalled                 func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	InitNotarizedHeadersCalled               func(startHeaders map[uint32]data.HeaderHandler) error
	RemoveLastNotarizedHeaderCalled          func()
	RestoreNotarizedHeadersToGenesisCalled   func()
}

// AddNotarizedHeader -
func (bngm *BlockNotarizerHandlerMock) AddNotarizedHeader(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte) {
	if bngm.AddNotarizedHeaderCalled != nil {
		bngm.AddNotarizedHeaderCalled(shardID, notarizedHeader, notarizedHeaderHash)
	}
}

// CleanupNotarizedHeadersBehindNonce -
func (bngm *BlockNotarizerHandlerMock) CleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64) {
	if bngm.CleanupNotarizedHeadersBehindNonceCalled != nil {
		bngm.CleanupNotarizedHeadersBehindNonceCalled(shardID, nonce)
	}
}

// DisplayNotarizedHeaders -
func (bngm *BlockNotarizerHandlerMock) DisplayNotarizedHeaders(shardID uint32, message string) {
	if bngm.DisplayNotarizedHeadersCalled != nil {
		bngm.DisplayNotarizedHeadersCalled(shardID, message)
	}
}

// GetFirstNotarizedHeader -
func (bngm *BlockNotarizerHandlerMock) GetFirstNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bngm.GetFirstNotarizedHeaderCalled != nil {
		return bngm.GetFirstNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

// GetLastNotarizedHeader -
func (bngm *BlockNotarizerHandlerMock) GetLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bngm.GetLastNotarizedHeaderCalled != nil {
		return bngm.GetLastNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

// GetLastNotarizedHeaderNonce -
func (bngm *BlockNotarizerHandlerMock) GetLastNotarizedHeaderNonce(shardID uint32) uint64 {
	if bngm.GetLastNotarizedHeaderNonceCalled != nil {
		return bngm.GetLastNotarizedHeaderNonceCalled(shardID)
	}

	return 0
}

// GetNotarizedHeader -
func (bngm *BlockNotarizerHandlerMock) GetNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if bngm.GetNotarizedHeaderCalled != nil {
		return bngm.GetNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

// InitNotarizedHeaders -
func (bngm *BlockNotarizerHandlerMock) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	if bngm.InitNotarizedHeadersCalled != nil {
		return bngm.InitNotarizedHeadersCalled(startHeaders)
	}

	return nil
}

// RemoveLastNotarizedHeader -
func (bngm *BlockNotarizerHandlerMock) RemoveLastNotarizedHeader() {
	if bngm.RemoveLastNotarizedHeaderCalled != nil {
		bngm.RemoveLastNotarizedHeaderCalled()
	}
}

// RestoreNotarizedHeadersToGenesis -
func (bngm *BlockNotarizerHandlerMock) RestoreNotarizedHeadersToGenesis() {
	if bngm.RestoreNotarizedHeadersToGenesisCalled != nil {
		bngm.RestoreNotarizedHeadersToGenesisCalled()
	}
}

// IsInterfaceNil -
func (bngm *BlockNotarizerHandlerMock) IsInterfaceNil() bool {
	return bngm == nil
}
