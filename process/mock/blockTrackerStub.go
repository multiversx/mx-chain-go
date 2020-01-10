package mock

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type headerInfo struct {
	hash   []byte
	header data.HeaderHandler
}

type BlockTrackerMock struct {
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
	RegisterCrossNotarizedHeadersHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandlerCalled  func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastNotarizedHeadersCalled           func()
	RestoreToGenesisCalled                     func()

	mutCrossNotarizedHeaders sync.RWMutex
	crossNotarizedHeaders    map[uint32][]*headerInfo

	mutHeaders sync.RWMutex
	headers    map[uint32][]*headerInfo
}

func NewBlockTrackerMock(startHeaders map[uint32]data.HeaderHandler) *BlockTrackerMock {
	bts := BlockTrackerMock{}
	bts.headers = make(map[uint32][]*headerInfo)
	_ = bts.InitCrossNotarizedHeaders(startHeaders)
	return &bts
}

func (btm *BlockTrackerMock) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	if btm.AddTrackedHeaderCalled != nil {
		btm.AddTrackedHeaderCalled(header, hash)
	}

	if check.IfNil(header) {
		return
	}

	shardID := header.GetShardID()

	btm.mutHeaders.Lock()
	defer btm.mutHeaders.Unlock()

	headersForShard, ok := btm.headers[shardID]
	if !ok {
		headersForShard = make([]*headerInfo, 0)
	}

	for _, headerInfo := range headersForShard {
		if bytes.Equal(headerInfo.hash, hash) {
			return
		}
	}

	headersForShard = append(headersForShard, &headerInfo{hash: hash, header: header})
	btm.headers[shardID] = headersForShard
}

func (btm *BlockTrackerMock) InitCrossNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	btm.mutCrossNotarizedHeaders.Lock()
	defer btm.mutCrossNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrNotarizedHeadersSliceIsNil
	}

	btm.crossNotarizedHeaders = make(map[uint32][]*headerInfo)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		btm.crossNotarizedHeaders[shardID] = append(btm.crossNotarizedHeaders[shardID], &headerInfo{header: startHeader, hash: nil})
	}

	return nil
}

func (btm *BlockTrackerMock) AddCrossNotarizedHeader(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte) {
	if btm.AddCrossNotarizedHeaderCalled != nil {
		btm.AddCrossNotarizedHeaderCalled(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
		return
	}

	if check.IfNil(crossNotarizedHeader) {
		return
	}

	btm.mutCrossNotarizedHeaders.Lock()
	btm.crossNotarizedHeaders[shardID] = append(btm.crossNotarizedHeaders[shardID], &headerInfo{header: crossNotarizedHeader, hash: crossNotarizedHeaderHash})
	if len(btm.crossNotarizedHeaders[shardID]) > 1 {
		sort.Slice(btm.crossNotarizedHeaders[shardID], func(i, j int) bool {
			return btm.crossNotarizedHeaders[shardID][i].header.GetNonce() < btm.crossNotarizedHeaders[shardID][j].header.GetNonce()
		})
	}
	btm.mutCrossNotarizedHeaders.Unlock()
}

func (btm *BlockTrackerMock) AddSelfNotarizedHeader(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte) {
	if btm.AddSelfNotarizedHeaderCalled != nil {
		btm.AddSelfNotarizedHeaderCalled(shardID, selfNotarizedHeader, selfNotarizedHeaderHash)
	}
}

func (btm *BlockTrackerMock) CleanupHeadersBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64) {
	if btm.CleanupHeadersBehindNonceCalled != nil {
		btm.CleanupHeadersBehindNonceCalled(shardID, selfNotarizedNonce, crossNotarizedNonce)
	}
}

func (btm *BlockTrackerMock) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	if btm.ComputeLongestChainCalled != nil {
		return btm.ComputeLongestChainCalled(shardID, header)
	}

	headersInfo, ok := btm.headers[shardID]
	if !ok {
		return nil, nil
	}

	headers := make([]data.HeaderHandler, 0)
	hashes := make([][]byte, 0)

	for _, headerInfo := range headersInfo {
		headers = append(headers, headerInfo.header)
		hashes = append(hashes, headerInfo.hash)
	}

	return headers, hashes
}

func (btm *BlockTrackerMock) DisplayTrackedHeaders() {
	if btm.DisplayTrackedHeadersCalled != nil {
		btm.DisplayTrackedHeadersCalled()
	}
}

func (btm *BlockTrackerMock) GetCrossNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if btm.GetCrossNotarizedHeaderCalled != nil {
		return btm.GetCrossNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

func (btm *BlockTrackerMock) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if btm.GetLastCrossNotarizedHeaderCalled != nil {
		return btm.GetLastCrossNotarizedHeaderCalled(shardID)
	}

	btm.mutCrossNotarizedHeaders.RLock()
	defer btm.mutCrossNotarizedHeaders.RUnlock()

	if btm.crossNotarizedHeaders == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceIsNil
	}

	headerInfo := btm.lastCrossNotarizedHdrForShard(shardID)
	if headerInfo == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	return headerInfo.header, headerInfo.hash, nil
}

func (btm *BlockTrackerMock) lastCrossNotarizedHdrForShard(shardID uint32) *headerInfo {
	crossNotarizedHeadersCount := len(btm.crossNotarizedHeaders[shardID])
	if crossNotarizedHeadersCount > 0 {
		return btm.crossNotarizedHeaders[shardID][crossNotarizedHeadersCount-1]
	}

	return nil
}

func (btm *BlockTrackerMock) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	if btm.GetTrackedHeadersCalled != nil {
		return btm.GetTrackedHeadersCalled(shardID)
	}

	return nil, nil
}

func (btm *BlockTrackerMock) GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if btm.GetTrackedHeadersWithNonceCalled != nil {
		return btm.GetTrackedHeadersWithNonceCalled(shardID, nonce)
	}

	return nil, nil
}

func (btm *BlockTrackerMock) IsShardStuck(shardId uint32) bool {
	return btm.IsShardStuckCalled(shardId)
}

func (btm *BlockTrackerMock) RegisterCrossNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if btm.RegisterCrossNotarizedHeadersHandlerCalled != nil {
		btm.RegisterCrossNotarizedHeadersHandlerCalled(handler)
	}
}

func (btm *BlockTrackerMock) RegisterSelfNotarizedHeadersHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if btm.RegisterSelfNotarizedHeadersHandlerCalled != nil {
		btm.RegisterSelfNotarizedHeadersHandlerCalled(handler)
	}
}

func (btm *BlockTrackerMock) RemoveLastNotarizedHeaders() {
	if btm.RemoveLastNotarizedHeadersCalled != nil {
		btm.RemoveLastNotarizedHeadersCalled()
	}
}

func (btm *BlockTrackerMock) RestoreToGenesis() {
	if btm.RestoreToGenesisCalled != nil {
		btm.RestoreToGenesisCalled()
	}
}

func (btm *BlockTrackerMock) IsInterfaceNil() bool {
	return btm == nil
}
