package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// shardBlockTrack

func (sbt *shardBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	return sbt.getSelfHeaders(headerHandler)
}

func (sbt *shardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	return sbt.computeLongestSelfChain()
}

func (sbt *shardBlockTrack) ComputeNumPendingMiniBlocks(headers []data.HeaderHandler) {
	sbt.computeNumPendingMiniBlocks(headers)
}

// metaBlockTrack

func (mbt *metaBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	return mbt.getSelfHeaders(headerHandler)
}

func (mbt *metaBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	return mbt.computeLongestSelfChain()
}

func (sbt *shardBlockTrack) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	return sbt.blockBalancer.getNumPendingMiniBlocks(shardID)
}

// baseBlockTrack

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

func (bbt *baseBlockTrack) CleanupTrackedHeadersBehindNonce(shardID uint32, nonce uint64) {
	bbt.cleanupTrackedHeadersBehindNonce(shardID, nonce)
}

func (bbt *baseBlockTrack) DisplayTrackedHeadersForShard(shardID uint32, message string) {
	bbt.displayTrackedHeadersForShard(shardID, message)
}

// blockBalancer

func (bb *blockBalancer) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	bb.setNumPendingMiniBlocks(shardID, numPendingMiniBlocks)
}

func (bb *blockBalancer) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	return bb.getNumPendingMiniBlocks(shardID)
}

// blockNotifier

func (bn *blockNotifier) RegisterHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	bn.registerHandler(handler)
}

func (bn *blockNotifier) CallHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	bn.callHandlers(shardID, headers, headersHashes)
}

func (bn *blockNotifier) GetNotarizedHeadersHandlers() []func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	bn.mutNotarizedHeadersHandlers.RLock()
	notarizedHeadersHandlers := bn.notarizedHeadersHandlers
	bn.mutNotarizedHeadersHandlers.RUnlock()

	return notarizedHeadersHandlers
}

// blockNotarizer

func (bn *blockNotarizer) AddNotarizedHeader(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte) {
	bn.addNotarizedHeader(shardID, notarizedHeader, notarizedHeaderHash)
}

func (bn *blockNotarizer) GetNotarizedHeaders() map[uint32][]*headerInfo {
	bn.mutNotarizedHeaders.RLock()
	notarizedHeaders := bn.notarizedHeaders
	bn.mutNotarizedHeaders.RUnlock()

	return notarizedHeaders
}

func (bn *blockNotarizer) GetNotarizedHeader(shardID uint32, index int) data.HeaderHandler {
	bn.mutNotarizedHeaders.RLock()
	notarizedHeader := bn.notarizedHeaders[shardID][index].header
	bn.mutNotarizedHeaders.RUnlock()

	return notarizedHeader
}

func (bn *blockNotarizer) CleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64) {
	bn.cleanupNotarizedHeadersBehindNonce(shardID, nonce)
}

func (bn *blockNotarizer) DisplayNotarizedHeaders(shardID uint32, message string) {
	bn.displayNotarizedHeaders(shardID, message)
}

func (bn *blockNotarizer) GetLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	return bn.getLastNotarizedHeader(shardID)
}

func (bn *blockNotarizer) GetLastNotarizedHeaderNonce(shardID uint32) uint64 {
	return bn.getLastNotarizedHeaderNonce(shardID)
}

func (bn *blockNotarizer) LastNotarizedHeaderInfo(shardID uint32) *headerInfo {
	return bn.lastNotarizedHeaderInfo(shardID)
}

func (bn *blockNotarizer) GetNotarizedHeaderWithOffset(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	return bn.getNotarizedHeader(shardID, offset)
}

func (bn *blockNotarizer) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	return bn.initNotarizedHeaders(startHeaders)
}

func (bn *blockNotarizer) RemoveLastNotarizedHeader() {
	bn.removeLastNotarizedHeader()
}

func (bn *blockNotarizer) RestoreNotarizedHeadersToGenesis() {
	bn.restoreNotarizedHeadersToGenesis()
}
