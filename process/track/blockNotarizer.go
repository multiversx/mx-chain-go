package track

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (bbt *baseBlockTrack) setCrossNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrCrossNotarizedHdrsSliceIsNil
	}

	bbt.crossNotarizedHeaders = make(map[uint32][]data.HeaderHandler)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], startHeader)
	}

	return nil
}

func (bbt *baseBlockTrack) addCrossNotarizedHeaders(shardID uint32, crossNotarizedHeaders []data.HeaderHandler) {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

	bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], crossNotarizedHeaders...)
}

func (bbt *baseBlockTrack) getLastCrossNotarizedHeaderNonce(shardID uint32) uint64 {
	lastCrossNotarizedHeader, err := bbt.getLastCrossNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastCrossNotarizedHeader.GetNonce()
}

func (bbt *baseBlockTrack) getLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, error) {
	bbt.mutCrossNotarizedHeaders.RLock()
	defer bbt.mutCrossNotarizedHeaders.RUnlock()

	if bbt.crossNotarizedHeaders == nil {
		return nil, process.ErrCrossNotarizedHdrsSliceIsNil
	}

	headerHandler := bbt.lastCrossNotarizedHdrForShard(shardID)
	if check.IfNil(headerHandler) {
		return nil, process.ErrCrossNotarizedHdrsSliceForShardIsNil
	}

	return headerHandler, nil
}

func (bbt *baseBlockTrack) lastCrossNotarizedHdrForShard(shardID uint32) data.HeaderHandler {
	crossNotarizedHeadersCount := len(bbt.crossNotarizedHeaders[shardID])
	if crossNotarizedHeadersCount > 0 {
		return bbt.crossNotarizedHeaders[shardID][crossNotarizedHeadersCount-1]
	}

	return nil
}

func (bbt *baseBlockTrack) setSelfNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrSelfNotarizedHdrsSliceIsNil
	}

	bbt.selfNotarizedHeaders = make(map[uint32][]data.HeaderHandler)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], startHeader)
	}

	return nil
}

func (bbt *baseBlockTrack) addSelfNotarizedHeaders(shardID uint32, selfNotarizedHeaders []data.HeaderHandler) {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

	bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], selfNotarizedHeaders...)
}

func (bbt *baseBlockTrack) getLastSelfNotarizedHeaderNonce(shardID uint32) uint64 {
	lastSelfNotarizedHeader, err := bbt.getLastSelfNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastSelfNotarizedHeader.GetNonce()
}

func (bbt *baseBlockTrack) getLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, error) {
	bbt.mutSelfNotarizedHeaders.RLock()
	defer bbt.mutSelfNotarizedHeaders.RUnlock()

	if bbt.selfNotarizedHeaders == nil {
		return nil, process.ErrSelfNotarizedHdrsSliceIsNil
	}

	headerHandler := bbt.lastSelfNotarizedHdrForShard(shardID)
	if check.IfNil(headerHandler) {
		return nil, process.ErrSelfNotarizedHdrsSliceForShardIsNil
	}

	return headerHandler, nil
}

func (bbt *baseBlockTrack) lastSelfNotarizedHdrForShard(shardID uint32) data.HeaderHandler {
	selfNotarizedHeadersCount := len(bbt.selfNotarizedHeaders[shardID])
	if selfNotarizedHeadersCount > 0 {
		return bbt.selfNotarizedHeaders[shardID][selfNotarizedHeadersCount-1]
	}

	return nil
}
