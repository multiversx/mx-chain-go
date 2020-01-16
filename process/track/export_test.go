package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

func (sbt *shardBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	return sbt.getSelfHeaders(headerHandler)
}

func (mbt *metaBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	return mbt.getSelfHeaders(headerHandler)
}

func (sbt *shardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	return sbt.computeLongestSelfChain()
}

func (mbt *metaBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	return mbt.computeLongestSelfChain()
}

func (sbt *shardBlockTrack) ComputePendingMiniBlockHeaders(headers []data.HeaderHandler) {
	sbt.computePendingMiniBlockHeaders(headers)
}

func (sbt *shardBlockTrack) PendingMiniBlockHeaders(shardID uint32) uint32 {
	return sbt.blockBalancer.pendingMiniBlockHeaders(shardID)
}

func (bbt *baseBlockTrack) ReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	bbt.receivedHeader(headerHandler, headerHash)
}

func CheckTrackerNilParameters(arguments ArgBaseTracker) error {
	return checkTrackerNilParameters(arguments)
}

func (bbt *baseBlockTrack) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	return bbt.initNotarizedHeaders(startHeaders)
}

func (bbt *baseBlockTrack) GetLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	return bbt.selfNotarizer.getLastNotarizedHeader(shardID)
}

func (bbt *baseBlockTrack) ReceivedShardHeader(headerHandler data.HeaderHandler, shardHeaderHash []byte) {
	bbt.receivedShardHeader(headerHandler, shardHeaderHash)
}

func (bbt *baseBlockTrack) ReceivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	bbt.receivedMetaBlock(headerHandler, metaBlockHash)
}

func (bbt *baseBlockTrack) AddHeader(header data.HeaderHandler, hash []byte) {
	bbt.addHeader(header, hash)
}
