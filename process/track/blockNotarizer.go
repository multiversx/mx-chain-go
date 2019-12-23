package track

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (bbt *baseBlockTrack) initCrossNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrCrossNotarizedHdrsSliceIsNil
	}

	bbt.crossNotarizedHeaders = make(map[uint32][]*headerInfo)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		startHeaderHash, err := core.CalculateHash(bbt.marshalizer, bbt.hasher, startHeader)
		if err != nil {
			return err
		}

		bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], &headerInfo{header: startHeader, hash: startHeaderHash})
	}

	return nil
}

func (bbt *baseBlockTrack) getLastCrossNotarizedHeaderNonce(shardID uint32) uint64 {
	lastCrossNotarizedHeader, _, err := bbt.GetLastCrossNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastCrossNotarizedHeader.GetNonce()
}

func (bbt *baseBlockTrack) lastCrossNotarizedHdrForShard(shardID uint32) *headerInfo {
	crossNotarizedHeadersCount := len(bbt.crossNotarizedHeaders[shardID])
	if crossNotarizedHeadersCount > 0 {
		return bbt.crossNotarizedHeaders[shardID][crossNotarizedHeadersCount-1]
	}

	return nil
}

func (bbt *baseBlockTrack) restoreCrossNotarizedHeadersToGenesis() {
	bbt.mutCrossNotarizedHeaders.Lock()
	for shardID := range bbt.crossNotarizedHeaders {
		crossNotarizedHeadersCount := len(bbt.crossNotarizedHeaders[shardID])
		if crossNotarizedHeadersCount > 1 {
			bbt.crossNotarizedHeaders[shardID] = bbt.crossNotarizedHeaders[shardID][:1]
		}
	}
	bbt.mutCrossNotarizedHeaders.Unlock()
}

func (bbt *baseBlockTrack) initSelfNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrSelfNotarizedHdrsSliceIsNil
	}

	bbt.selfNotarizedHeaders = make(map[uint32][]*headerInfo)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		startHeaderHash, err := core.CalculateHash(bbt.marshalizer, bbt.hasher, startHeader)
		if err != nil {
			return err
		}

		bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], &headerInfo{header: startHeader, hash: startHeaderHash})
	}

	return nil
}

func (bbt *baseBlockTrack) getLastSelfNotarizedHeaderNonce(shardID uint32) uint64 {
	lastSelfNotarizedHeaderInfo, err := bbt.getLastSelfNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastSelfNotarizedHeaderInfo.header.GetNonce()
}

func (bbt *baseBlockTrack) getLastSelfNotarizedHeader(shardID uint32) (*headerInfo, error) {
	bbt.mutSelfNotarizedHeaders.RLock()
	defer bbt.mutSelfNotarizedHeaders.RUnlock()

	if bbt.selfNotarizedHeaders == nil {
		return nil, process.ErrSelfNotarizedHdrsSliceIsNil
	}

	headerInfo := bbt.lastSelfNotarizedHdrForShard(shardID)
	if headerInfo == nil {
		return nil, process.ErrSelfNotarizedHdrsSliceForShardIsNil
	}

	return headerInfo, nil
}

func (bbt *baseBlockTrack) lastSelfNotarizedHdrForShard(shardID uint32) *headerInfo {
	selfNotarizedHeadersCount := len(bbt.selfNotarizedHeaders[shardID])
	if selfNotarizedHeadersCount > 0 {
		return bbt.selfNotarizedHeaders[shardID][selfNotarizedHeadersCount-1]
	}

	return nil
}

func (bbt *baseBlockTrack) restoreSelfNotarizedHeadersToGenesis() {
	bbt.mutSelfNotarizedHeaders.Lock()
	for shardID := range bbt.selfNotarizedHeaders {
		selfNotarizedHeadersCount := len(bbt.selfNotarizedHeaders[shardID])
		if selfNotarizedHeadersCount > 1 {
			bbt.selfNotarizedHeaders[shardID] = bbt.selfNotarizedHeaders[shardID][:1]
		}
	}
	bbt.mutSelfNotarizedHeaders.Unlock()
}

func (bbt *baseBlockTrack) restoreTrackedHeadersToGenesis() {
	bbt.mutHeaders.Lock()
	bbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	bbt.mutHeaders.Unlock()
}
