package track

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// shardBlockTrack

// SetNumPendingMiniBlocks -
func (sbt *shardBlockTrack) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	sbt.blockBalancer.SetNumPendingMiniBlocks(shardID, numPendingMiniBlocks)
}

// GetNumPendingMiniBlocks -
func (sbt *shardBlockTrack) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	return sbt.blockBalancer.GetNumPendingMiniBlocks(shardID)
}

// SetLastShardProcessedMetaNonce -
func (sbt *shardBlockTrack) SetLastShardProcessedMetaNonce(shardID uint32, nonce uint64) {
	sbt.blockBalancer.SetLastShardProcessedMetaNonce(shardID, nonce)
}

// GetLastShardProcessedMetaNonce -
func (sbt *shardBlockTrack) GetLastShardProcessedMetaNonce(shardID uint32) uint64 {
	return sbt.blockBalancer.GetLastShardProcessedMetaNonce(shardID)
}

// GetTrackedShardHeaderWithNonceAndHash -
func (sbt *shardBlockTrack) GetTrackedShardHeaderWithNonceAndHash(shardID uint32, nonce uint64, hash []byte) (data.HeaderHandler, error) {
	return sbt.getTrackedShardHeaderWithNonceAndHash(shardID, nonce, hash)
}

// metaBlockTrack

// GetTrackedMetaBlockWithHash -
func (mbt *metaBlockTrack) GetTrackedMetaBlockWithHash(hash []byte) (*block.MetaBlock, error) {
	return mbt.getTrackedMetaBlockWithHash(hash)
}

// baseBlockTrack

// ReceivedHeader -
func (bbt *baseBlockTrack) ReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	bbt.receivedHeader(headerHandler, headerHash)
}

// CheckTrackerNilParameters -
func CheckTrackerNilParameters(arguments ArgBaseTracker) error {
	return checkTrackerNilParameters(arguments)
}

// InitNotarizedHeaders -
func (bbt *baseBlockTrack) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	return bbt.initNotarizedHeaders(startHeaders)
}

// ReceivedShardHeader -
func (bbt *baseBlockTrack) ReceivedShardHeader(headerHandler data.HeaderHandler, shardHeaderHash []byte) {
	bbt.receivedShardHeader(headerHandler, shardHeaderHash)
}

// ReceivedMetaBlock -
func (bbt *baseBlockTrack) ReceivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	bbt.receivedMetaBlock(headerHandler, metaBlockHash)
}

// GetMaxNumHeadersToKeepPerShard -
func (bbt *baseBlockTrack) GetMaxNumHeadersToKeepPerShard() int {
	return bbt.maxNumHeadersToKeepPerShard
}

// ShouldAddHeaderForCrossShard -
func (bbt *baseBlockTrack) ShouldAddHeaderForCrossShard(headerHandler data.HeaderHandler) bool {
	return bbt.shouldAddHeaderForShard(headerHandler, bbt.crossNotarizer, headerHandler.GetShardID())
}

// ShouldAddHeaderForSelfShard -
func (bbt *baseBlockTrack) ShouldAddHeaderForSelfShard(headerHandler data.HeaderHandler) bool {
	return bbt.shouldAddHeaderForShard(headerHandler, bbt.selfNotarizer, core.MetachainShardId)
}

// AddHeader -
func (bbt *baseBlockTrack) AddHeader(header data.HeaderHandler, hash []byte) bool {
	return bbt.addHeader(header, hash)
}

// AppendTrackedHeader -
func (bbt *baseBlockTrack) AppendTrackedHeader(headerHandler data.HeaderHandler) {
	bbt.mutHeaders.Lock()
	headersForShard, ok := bbt.headers[headerHandler.GetShardID()]
	if !ok {
		headersForShard = make(map[uint64][]*HeaderInfo)
		bbt.headers[headerHandler.GetShardID()] = headersForShard
	}

	headersForShard[headerHandler.GetNonce()] = append(headersForShard[headerHandler.GetNonce()], &HeaderInfo{Header: headerHandler})
	bbt.mutHeaders.Unlock()
}

// CleanupTrackedHeadersBehindNonce -
func (bbt *baseBlockTrack) CleanupTrackedHeadersBehindNonce(shardID uint32, nonce uint64) {
	bbt.cleanupTrackedHeadersBehindNonce(shardID, nonce)
}

// DisplayTrackedHeadersForShard -
func (bbt *baseBlockTrack) DisplayTrackedHeadersForShard(shardID uint32, message string) {
	bbt.displayTrackedHeadersForShard(shardID, message)
}

// SetRoundHandler -
func (bbt *baseBlockTrack) SetRoundHandler(roundHandler process.RoundHandler) {
	bbt.roundHandler = roundHandler
}

// SetCrossNotarizer -
func (bbt *baseBlockTrack) SetCrossNotarizer(notarizer blockNotarizerHandler) {
	bbt.crossNotarizer = notarizer
}

// SetSelfNotarizer -
func (bbt *baseBlockTrack) SetSelfNotarizer(notarizer blockNotarizerHandler) {
	bbt.selfNotarizer = notarizer
}

// SetShardCoordinator -
func (bbt *baseBlockTrack) SetShardCoordinator(coordinator sharding.Coordinator) {
	bbt.shardCoordinator = coordinator
}

// NewBaseBlockTrack -
func NewBaseBlockTrack() *baseBlockTrack {
	return &baseBlockTrack{}
}

// DoWhitelistWithMetaBlockIfNeeded -
func (bbt *baseBlockTrack) DoWhitelistWithMetaBlockIfNeeded(metaBlock *block.MetaBlock) {
	bbt.doWhitelistWithMetaBlockIfNeeded(metaBlock)
}

// DoWhitelistWithShardHeaderIfNeeded -
func (bbt *baseBlockTrack) DoWhitelistWithShardHeaderIfNeeded(shardHeader *block.Header) {
	bbt.doWhitelistWithShardHeaderIfNeeded(shardHeader)
}

// IsHeaderOutOfRange -
func (bbt *baseBlockTrack) IsHeaderOutOfRange(headerHandler data.HeaderHandler) bool {
	return bbt.isHeaderOutOfRange(headerHandler)
}

// blockNotifier

// GetNotarizedHeadersHandlers -
func (bn *blockNotifier) GetNotarizedHeadersHandlers() []func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	bn.mutNotarizedHeadersHandlers.RLock()
	notarizedHeadersHandlers := bn.notarizedHeadersHandlers
	bn.mutNotarizedHeadersHandlers.RUnlock()

	return notarizedHeadersHandlers
}

// blockNotarizer

// AppendNotarizedHeader -
func (bn *blockNotarizer) AppendNotarizedHeader(headerHandler data.HeaderHandler) {
	bn.mutNotarizedHeaders.Lock()
	bn.notarizedHeaders[headerHandler.GetShardID()] = append(bn.notarizedHeaders[headerHandler.GetShardID()], &HeaderInfo{Header: headerHandler})
	bn.mutNotarizedHeaders.Unlock()
}

// GetNotarizedHeaders -
func (bn *blockNotarizer) GetNotarizedHeaders() map[uint32][]*HeaderInfo {
	bn.mutNotarizedHeaders.RLock()
	notarizedHeaders := bn.notarizedHeaders
	bn.mutNotarizedHeaders.RUnlock()

	return notarizedHeaders
}

// GetNotarizedHeaderWithIndex -
func (bn *blockNotarizer) GetNotarizedHeaderWithIndex(shardID uint32, index int) data.HeaderHandler {
	bn.mutNotarizedHeaders.RLock()
	notarizedHeader := bn.notarizedHeaders[shardID][index].Header
	bn.mutNotarizedHeaders.RUnlock()

	return notarizedHeader
}

// LastNotarizedHeaderInfo -
func (bn *blockNotarizer) LastNotarizedHeaderInfo(shardID uint32) *HeaderInfo {
	return bn.lastNotarizedHeaderInfo(shardID)
}

// blockProcessor

// DoJobOnReceivedHeader -
func (bp *blockProcessor) DoJobOnReceivedHeader(shardID uint32) {
	bp.doJobOnReceivedHeader(shardID)
}

// DoJobOnReceivedCrossNotarizedHeader -
func (bp *blockProcessor) DoJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	bp.doJobOnReceivedCrossNotarizedHeader(shardID)
}

// ComputeLongestChainFromLastCrossNotarized -
func (bp *blockProcessor) ComputeLongestChainFromLastCrossNotarized(shardID uint32) (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	return bp.computeLongestChainFromLastCrossNotarized(shardID)
}

// ComputeSelfNotarizedHeaders -
func (bp *blockProcessor) ComputeSelfNotarizedHeaders(headers []data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	return bp.computeSelfNotarizedHeaders(headers)
}

// GetNextHeader -
func (bp *blockProcessor) GetNextHeader(
	longestChainHeadersIndexes *[]int,
	headersIndexes []int,
	prevHeader data.HeaderHandler,
	sortedHeaders []data.HeaderHandler,
	sortedHashes [][]byte,
	index int,
) {
	bp.getNextHeader(longestChainHeadersIndexes, headersIndexes, prevHeader, sortedHeaders, sortedHashes, index)
}

// CheckHeaderFinality -
func (bp *blockProcessor) CheckHeaderFinality(
	header data.HeaderHandler,
	sortedHeaders []data.HeaderHandler,
	sortedHashes [][]byte,
	index int,
) error {
	return bp.checkHeaderFinality(header, sortedHeaders, sortedHashes, index)
}

// RequestHeadersIfNeeded -
func (bp *blockProcessor) RequestHeadersIfNeeded(lastNotarizedHeader data.HeaderHandler, sortedHeaders []data.HeaderHandler, longestChainHeaders []data.HeaderHandler) {
	bp.requestHeadersIfNeeded(lastNotarizedHeader, sortedHeaders, longestChainHeaders)
}

// GetLatestValidHeader -
func (bp *blockProcessor) GetLatestValidHeader(lastNotarizedHeader data.HeaderHandler, longestChainHeaders []data.HeaderHandler) data.HeaderHandler {
	return bp.getLatestValidHeader(lastNotarizedHeader, longestChainHeaders)
}

// GetHighestRoundInReceivedHeaders -
func (bp *blockProcessor) GetHighestRoundInReceivedHeaders(latestValidHeader data.HeaderHandler, sortedReceivedHeaders []data.HeaderHandler) uint64 {
	return bp.getHighestRoundInReceivedHeaders(latestValidHeader, sortedReceivedHeaders)
}

// RequestHeadersIfNothingNewIsReceived -
func (bp *blockProcessor) RequestHeadersIfNothingNewIsReceived(lastNotarizedHeaderNonce uint64, latestValidHeader data.HeaderHandler, highestRoundInReceivedHeaders uint64) {
	bp.requestHeadersIfNothingNewIsReceived(lastNotarizedHeaderNonce, latestValidHeader, highestRoundInReceivedHeaders)
}

// RequestHeaders -
func (bp *blockProcessor) RequestHeaders(shardID uint32, fromNonce uint64) {
	bp.requestHeaders(shardID, fromNonce)
}

// ShouldProcessReceivedHeader -
func (bp *blockProcessor) ShouldProcessReceivedHeader(headerHandler data.HeaderHandler) bool {
	return bp.shouldProcessReceivedHeader(headerHandler)
}

// miniBlockTrack

// ReceivedMiniBlock -
func (mbt *miniBlockTrack) ReceivedMiniBlock(key []byte, value interface{}) {
	mbt.receivedMiniBlock(key, value)
}

// GetTransactionPool -
func (mbt *miniBlockTrack) GetTransactionPool(mbType block.Type) dataRetriever.ShardedDataCacherNotifier {
	return mbt.getTransactionPool(mbType)
}

// SetBlockTransactionsPool -
func (mbt *miniBlockTrack) SetBlockTransactionsPool(blockTransactionsPool dataRetriever.ShardedDataCacherNotifier) {
	mbt.blockTransactionsPool = blockTransactionsPool
}
