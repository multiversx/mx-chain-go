package track

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (bbt *baseBlockTrack) setCrossNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
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

//func (bbt *baseBlockTrack) setCrossNotarizedHeadersForShard(
//	shardID uint32,
//	lastCrossNotarizedHeader data.HeaderHandler,
//	lastCrossNotarizedHeaderHash []byte,
//	crossNotarizedHeaders []data.HeaderHandler,
//	crossNotarizedHeadersHashes [][]byte,
//) {
//	bbt.mutCrossNotarizedHeaders.Lock()
//	delete(bbt.crossNotarizedHeaders, shardID)
//	bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], &headerInfo{header: lastCrossNotarizedHeader, hash: lastCrossNotarizedHeaderHash})
//	for index, crossNotarizedHeader := range crossNotarizedHeaders {
//		bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], &headerInfo{header: crossNotarizedHeader, hash: crossNotarizedHeadersHashes[index]})
//	}
//	bbt.mutCrossNotarizedHeaders.Unlock()
//}

func (bbt *baseBlockTrack) addCrossNotarizedHeaders(
	shardID uint32,
	crossNotarizedHeaders []data.HeaderHandler,
	crossNotarizedHeadersHashes [][]byte,
) ([]data.HeaderHandler, [][]byte) {

	addedCrossNotarizedHeaders := make([]data.HeaderHandler, 0)
	addedCrossNotarizedHeadersHashes := make([][]byte, 0)

	bbt.mutCrossNotarizedHeaders.Lock()
	for index, crossNotarizedHeader := range crossNotarizedHeaders {
		if bbt.isHashInCrossNotarizedHeadersForShard(shardID, crossNotarizedHeadersHashes[index]) {
			continue
		}

		header := crossNotarizedHeader
		hash := crossNotarizedHeadersHashes[index]
		bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], &headerInfo{header: header, hash: hash})
		addedCrossNotarizedHeaders = append(addedCrossNotarizedHeaders, header)
		addedCrossNotarizedHeadersHashes = append(addedCrossNotarizedHeadersHashes, hash)
	}
	bbt.mutCrossNotarizedHeaders.Unlock()

	return addedCrossNotarizedHeaders, addedCrossNotarizedHeadersHashes
}

func (bbt *baseBlockTrack) isHashInCrossNotarizedHeadersForShard(shardID uint32, hash []byte) bool {
	crossNotarizedHeadersInfo := bbt.crossNotarizedHeaders[shardID]
	for _, crossNotarizedHeaderInfo := range crossNotarizedHeadersInfo {
		if bytes.Equal(crossNotarizedHeaderInfo.hash, hash) {
			return true
		}
	}

	return false
}

func (bbt *baseBlockTrack) getLastCrossNotarizedHeaderNonce(shardID uint32) uint64 {
	lastCrossNotarizedHeaderInfo, err := bbt.getLastCrossNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastCrossNotarizedHeaderInfo.header.GetNonce()
}

func (bbt *baseBlockTrack) getLastCrossNotarizedHeader(shardID uint32) (*headerInfo, error) {
	bbt.mutCrossNotarizedHeaders.RLock()
	defer bbt.mutCrossNotarizedHeaders.RUnlock()

	if bbt.crossNotarizedHeaders == nil {
		return nil, process.ErrCrossNotarizedHdrsSliceIsNil
	}

	headerInfo := bbt.lastCrossNotarizedHdrForShard(shardID)
	if headerInfo == nil {
		return nil, process.ErrCrossNotarizedHdrsSliceForShardIsNil
	}

	return headerInfo, nil
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

func (bbt *baseBlockTrack) setSelfNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
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

//func (bbt *baseBlockTrack) setSelfNotarizedHeadersForShard(
//	shardID uint32,
//	lastSelfNotarizedHeader data.HeaderHandler,
//	lastSelfNotarizedHeaderHash []byte,
//	selfNotarizedHeaders []data.HeaderHandler,
//	selfNotarizedHeadersHashes [][]byte,
//) {
//	bbt.mutSelfNotarizedHeaders.Lock()
//	delete(bbt.selfNotarizedHeaders, shardID)
//	bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], &headerInfo{header: lastSelfNotarizedHeader, hash: lastSelfNotarizedHeaderHash})
//	for index, selfNotarizedHeader := range selfNotarizedHeaders {
//		bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], &headerInfo{header: selfNotarizedHeader, hash: selfNotarizedHeadersHashes[index]})
//	}
//	bbt.mutSelfNotarizedHeaders.Unlock()
//}

func (bbt *baseBlockTrack) addSelfNotarizedHeaders(
	shardID uint32,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) ([]data.HeaderHandler, [][]byte) {

	addedSelfNotarizedHeaders := make([]data.HeaderHandler, 0)
	addedSelfNotarizedHeadersHashes := make([][]byte, 0)

	bbt.mutSelfNotarizedHeaders.Lock()
	for index, selfNotarizedHeader := range selfNotarizedHeaders {
		if bbt.isHashInSelfNotarizedHeadersForShard(shardID, selfNotarizedHeadersHashes[index]) {
			continue
		}

		header := selfNotarizedHeader
		hash := selfNotarizedHeadersHashes[index]
		bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], &headerInfo{header: header, hash: hash})
		addedSelfNotarizedHeaders = append(addedSelfNotarizedHeaders, header)
		addedSelfNotarizedHeadersHashes = append(addedSelfNotarizedHeadersHashes, hash)
	}
	bbt.mutSelfNotarizedHeaders.Unlock()

	return addedSelfNotarizedHeaders, addedSelfNotarizedHeadersHashes
}

func (bbt *baseBlockTrack) isHashInSelfNotarizedHeadersForShard(shardID uint32, hash []byte) bool {
	selfNotarizedHeadersInfo := bbt.selfNotarizedHeaders[shardID]
	for _, selfNotarizedHeaderInfo := range selfNotarizedHeadersInfo {
		if bytes.Equal(selfNotarizedHeaderInfo.hash, hash) {
			return true
		}
	}

	return false
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
