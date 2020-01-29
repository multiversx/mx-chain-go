package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockNotarizerHandlerMock struct {
	AddNotarizedHeaderCalled                 func(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte)
	CleanupNotarizedHeadersBehindNonceCalled func(shardID uint32, nonce uint64)
	DisplayNotarizedHeadersCalled            func(shardID uint32, message string)
	GetLastNotarizedHeaderCalled             func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastNotarizedHeaderNonceCalled        func(shardID uint32) uint64
	GetNotarizedHeaderCalled                 func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	InitNotarizedHeadersCalled               func(startHeaders map[uint32]data.HeaderHandler) error
	RemoveLastNotarizedHeaderCalled          func()
	RestoreNotarizedHeadersToGenesisCalled   func()
}

func (bngm *BlockNotarizerHandlerMock) AddNotarizedHeader(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte) {
	if bngm.AddNotarizedHeaderCalled != nil {
		bngm.AddNotarizedHeaderCalled(shardID, notarizedHeader, notarizedHeaderHash)
	}
}

func (bngm *BlockNotarizerHandlerMock) CleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64) {
	if bngm.CleanupNotarizedHeadersBehindNonceCalled != nil {
		bngm.CleanupNotarizedHeadersBehindNonceCalled(shardID, nonce)
	}
}

func (bngm *BlockNotarizerHandlerMock) DisplayNotarizedHeaders(shardID uint32, message string) {
	if bngm.DisplayNotarizedHeadersCalled != nil {
		bngm.DisplayNotarizedHeadersCalled(shardID, message)
	}
}

func (bngm *BlockNotarizerHandlerMock) GetLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bngm.GetLastNotarizedHeaderCalled != nil {
		return bngm.GetLastNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

func (bngm *BlockNotarizerHandlerMock) GetLastNotarizedHeaderNonce(shardID uint32) uint64 {
	if bngm.GetLastNotarizedHeaderNonceCalled != nil {
		return bngm.GetLastNotarizedHeaderNonceCalled(shardID)
	}

	return 0
}

func (bngm *BlockNotarizerHandlerMock) GetNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if bngm.GetNotarizedHeaderCalled != nil {
		return bngm.GetNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

func (bngm *BlockNotarizerHandlerMock) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	if bngm.InitNotarizedHeadersCalled != nil {
		return bngm.InitNotarizedHeadersCalled(startHeaders)
	}

	return nil
}

func (bngm *BlockNotarizerHandlerMock) RemoveLastNotarizedHeader() {
	if bngm.RemoveLastNotarizedHeaderCalled != nil {
		bngm.RemoveLastNotarizedHeaderCalled()
	}
}

func (bngm *BlockNotarizerHandlerMock) RestoreNotarizedHeadersToGenesis() {
	if bngm.RestoreNotarizedHeadersToGenesisCalled != nil {
		bngm.RestoreNotarizedHeadersToGenesisCalled()
	}
}

func (bngm *BlockNotarizerHandlerMock) IsInterfaceNil() bool {
	return bngm == nil
}
