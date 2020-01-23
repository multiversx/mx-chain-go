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

func (bbt *baseBlockTrack) SortHeadersFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	return bbt.sortHeadersFromNonce(shardID, nonce)
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

// blockProcessor

// BlockTrackerHandlerMock

type BlockTrackerHandlerMock struct {
	getSelfHeadersCalled              func(headerHandler data.HeaderHandler) []*headerInfo
	computeNumPendingMiniBlocksCalled func(headers []data.HeaderHandler)
	computeLongestSelfChainCalled     func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte)
	sortHeadersFromNonceCalled        func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
}

func (bthm *BlockTrackerHandlerMock) getSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	if bthm.getSelfHeadersCalled != nil {
		return bthm.getSelfHeadersCalled(headerHandler)
	}

	return nil
}

func (bthm *BlockTrackerHandlerMock) computeNumPendingMiniBlocks(headers []data.HeaderHandler) {
	if bthm.computeNumPendingMiniBlocksCalled != nil {
		bthm.computeNumPendingMiniBlocksCalled(headers)
	}
}

func (bthm *BlockTrackerHandlerMock) computeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	if bthm.computeLongestSelfChainCalled != nil {
		return bthm.computeLongestSelfChainCalled()
	}

	return nil, nil, nil, nil
}

func (bthm *BlockTrackerHandlerMock) sortHeadersFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if bthm.sortHeadersFromNonceCalled != nil {
		return bthm.sortHeadersFromNonceCalled(shardID, nonce)
	}

	return nil, nil
}

// BlockNotarizerHandlerMock

type BlockNotarizerHandlerMock struct {
	addNotarizedHeaderCalled                 func(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte)
	cleanupNotarizedHeadersBehindNonceCalled func(shardID uint32, nonce uint64)
	displayNotarizedHeadersCalled            func(shardID uint32, message string)
	getLastNotarizedHeaderCalled             func(shardID uint32) (data.HeaderHandler, []byte, error)
	getLastNotarizedHeaderNonceCalled        func(shardID uint32) uint64
	getNotarizedHeaderCalled                 func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	initNotarizedHeadersCalled               func(startHeaders map[uint32]data.HeaderHandler) error
	removeLastNotarizedHeaderCalled          func()
	restoreNotarizedHeadersToGenesisCalled   func()
}

func (bngm *BlockNotarizerHandlerMock) addNotarizedHeader(shardID uint32, notarizedHeader data.HeaderHandler, notarizedHeaderHash []byte) {
	if bngm.addNotarizedHeaderCalled != nil {
		bngm.addNotarizedHeaderCalled(shardID, notarizedHeader, notarizedHeaderHash)
	}
}

func (bngm *BlockNotarizerHandlerMock) cleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64) {
	if bngm.cleanupNotarizedHeadersBehindNonceCalled != nil {
		bngm.cleanupNotarizedHeadersBehindNonceCalled(shardID, nonce)
	}
}

func (bngm *BlockNotarizerHandlerMock) displayNotarizedHeaders(shardID uint32, message string) {
	if bngm.displayNotarizedHeadersCalled != nil {
		bngm.displayNotarizedHeadersCalled(shardID, message)
	}
}

func (bngm *BlockNotarizerHandlerMock) getLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if bngm.getLastNotarizedHeaderCalled != nil {
		return bngm.getLastNotarizedHeaderCalled(shardID)
	}

	return nil, nil, nil
}

func (bngm *BlockNotarizerHandlerMock) getLastNotarizedHeaderNonce(shardID uint32) uint64 {
	if bngm.getLastNotarizedHeaderNonceCalled != nil {
		return bngm.getLastNotarizedHeaderNonceCalled(shardID)
	}

	return 0
}

func (bngm *BlockNotarizerHandlerMock) getNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	if bngm.getNotarizedHeaderCalled != nil {
		return bngm.getNotarizedHeaderCalled(shardID, offset)
	}

	return nil, nil, nil
}

func (bngm *BlockNotarizerHandlerMock) initNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	if bngm.initNotarizedHeadersCalled != nil {
		return bngm.initNotarizedHeadersCalled(startHeaders)
	}

	return nil
}

func (bngm *BlockNotarizerHandlerMock) removeLastNotarizedHeader() {
	if bngm.removeLastNotarizedHeaderCalled != nil {
		bngm.removeLastNotarizedHeaderCalled()
	}
}

func (bngm *BlockNotarizerHandlerMock) restoreNotarizedHeadersToGenesis() {
	if bngm.restoreNotarizedHeadersToGenesisCalled != nil {
		bngm.restoreNotarizedHeadersToGenesisCalled()
	}
}

// BlockNotifierHandlerMock

type BlockNotifierHandlerMock struct {
	callHandlersCalled    func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	registerHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
}

func (bnhm *BlockNotifierHandlerMock) callHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if bnhm.callHandlersCalled != nil {
		bnhm.callHandlersCalled(shardID, headers, headersHashes)
	}
}

func (bnhm *BlockNotifierHandlerMock) registerHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bnhm.registerHandlerCalled != nil {
		bnhm.registerHandlerCalled(handler)
	}
}

func (bp *blockProcessor) ProcessReceivedHeader(header data.HeaderHandler) {
	bp.processReceivedHeader(header)
}
