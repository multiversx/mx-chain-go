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

type BlockTrackerStub struct {
	AddTrackedHeaderCalled                    func(header data.HeaderHandler, hash []byte)
	AddCrossNotarizedHeaderCalled             func(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeaderCalled              func(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	CleanupHeadersForShardBehindNonceCalled   func(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	ComputeLongestChainCalled                 func(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	DisplayTrackedHeadersCalled               func()
	GetCrossNotarizedHeaderCalled             func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeaderCalled         func(shardID uint32) (data.HeaderHandler, []byte, error)
	GetTrackedHeadersForShardCalled           func(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersForShardWithNonceCalled  func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuckCalled                        func(shardId uint32) bool
	LastHeaderForShardCalled                  func(shardId uint32) data.HeaderHandler
	RegisterSelfNotarizedHeadersHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastCrossNotarizedHeaderCalled      func()
	RestoreHeadersToGenesisCalled             func()

	mutCrossNotarizedHeaders sync.RWMutex
	crossNotarizedHeaders    map[uint32][]*headerInfo

	mutHeaders sync.RWMutex
	headers    map[uint32][]*headerInfo
}

func NewBlockTrackerStub(startHeaders map[uint32]data.HeaderHandler) *BlockTrackerStub {
	bts := BlockTrackerStub{}
	bts.headers = make(map[uint32][]*headerInfo)
	bts.InitCrossNotarizedHeaders(startHeaders)
	return &bts
}

func (bts *BlockTrackerStub) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	if bts.AddTrackedHeaderCalled != nil {
		bts.AddTrackedHeaderCalled(header, hash)
	}

	if check.IfNil(header) {
		return
	}

	shardID := header.GetShardID()

	bts.mutHeaders.Lock()
	defer bts.mutHeaders.Unlock()

	headersForShard, ok := bts.headers[shardID]
	if !ok {
		headersForShard = make([]*headerInfo, 0)
	}

	for _, headerInfo := range headersForShard {
		if bytes.Equal(headerInfo.hash, hash) {
			return
		}
	}

	headersForShard = append(headersForShard, &headerInfo{hash: hash, header: header})
	bts.headers[shardID] = headersForShard
}

func (bts *BlockTrackerStub) InitCrossNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bts.mutCrossNotarizedHeaders.Lock()
	defer bts.mutCrossNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrNotarizedHeadersSliceIsNil
	}

	bts.crossNotarizedHeaders = make(map[uint32][]*headerInfo)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		bts.crossNotarizedHeaders[shardID] = append(bts.crossNotarizedHeaders[shardID], &headerInfo{header: startHeader, hash: nil})
	}

	return nil
}

func (bts *BlockTrackerStub) AddCrossNotarizedHeader(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte) {
	if bts.AddCrossNotarizedHeaderCalled != nil {
		bts.AddCrossNotarizedHeaderCalled(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
		return
	}

	if check.IfNil(crossNotarizedHeader) {
		return
	}

	bts.mutCrossNotarizedHeaders.Lock()
	bts.crossNotarizedHeaders[shardID] = append(bts.crossNotarizedHeaders[shardID], &headerInfo{header: crossNotarizedHeader, hash: crossNotarizedHeaderHash})
	if len(bts.crossNotarizedHeaders[shardID]) > 1 {
		sort.Slice(bts.crossNotarizedHeaders[shardID], func(i, j int) bool {
			return bts.crossNotarizedHeaders[shardID][i].header.GetNonce() < bts.crossNotarizedHeaders[shardID][j].header.GetNonce()
		})
	}
	bts.mutCrossNotarizedHeaders.Unlock()
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

	headersInfo, ok := bts.headers[shardID]
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

	bts.mutCrossNotarizedHeaders.RLock()
	defer bts.mutCrossNotarizedHeaders.RUnlock()

	if bts.crossNotarizedHeaders == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceIsNil
	}

	headerInfo := bts.lastCrossNotarizedHdrForShard(shardID)
	if headerInfo == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	return headerInfo.header, headerInfo.hash, nil
}

func (bts *BlockTrackerStub) lastCrossNotarizedHdrForShard(shardID uint32) *headerInfo {
	crossNotarizedHeadersCount := len(bts.crossNotarizedHeaders[shardID])
	if crossNotarizedHeadersCount > 0 {
		return bts.crossNotarizedHeaders[shardID][crossNotarizedHeadersCount-1]
	}

	return nil
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
