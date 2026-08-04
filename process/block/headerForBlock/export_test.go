package headerForBlock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
)

// NewHeaderInfo -
func NewHeaderInfo(
	hdr data.HeaderHandler,
	usedInBlock bool,
	hasProof bool,
	hasProofRequested bool,
) *headerInfo {
	return newHeaderInfo(hdr, usedInBlock, hasProof, hasProofRequested)
}

// NewEmptyHeaderInfo -
func NewEmptyHeaderInfo() *headerInfo {
	return newEmptyHeaderInfo()
}

// NewLastNotarizedHeaderInfo -
func NewLastNotarizedHeaderInfo(
	header data.HeaderHandler,
	hash []byte,
	notarizedBasedOnProof bool,
	hasProof bool,
) *lastNotarizedHeaderInfo {
	return newLastNotarizedHeaderInfo(header, hash, notarizedBasedOnProof, hasProof)
}

// FilterHeadersWithoutProofs -
func (hfb *headersForBlock) FilterHeadersWithoutProofs() (map[string]HeaderInfo, error) {
	return hfb.filterHeadersWithoutProofs()
}

// RequestMissingAndUpdateBasedOnCrossShardData -
func (hfb *headersForBlock) RequestMissingAndUpdateBasedOnCrossShardData(cd crossShardMetaData) {
	hfb.requestMissingAndUpdateBasedOnCrossShardData(cd)
}

// ComputeExistingAndRequestMissingShardHeaders -
func (hfb *headersForBlock) ComputeExistingAndRequestMissingShardHeaders(metaBlock data.MetaHeaderHandler) {
	hfb.computeExistingAndRequestMissingShardHeaders(metaBlock)
}

// UpdateLastNotarizedBlockForShard -
func (hfb *headersForBlock) UpdateLastNotarizedBlockForShard(hdr data.ShardHeaderHandler, headerHash []byte) {
	hfb.updateLastNotarizedBlockForShard(hdr, headerHash)
}

// SetLastNotarizedHeaderForShard -
func (hfb *headersForBlock) SetLastNotarizedHeaderForShard(shardID uint32, info LastNotarizedHeaderInfoHandler) {
	hfb.lastNotarizedShardHeaders[shardID] = info
}

// SetHighestHdrNonceForCurrentBlock -
func (hfb *headersForBlock) SetHighestHdrNonceForCurrentBlock(shardID uint32, nonce uint64) {
	hfb.highestHdrNonce[shardID] = nonce
}

// SetShardBlockFinality -
func (hfb *headersForBlock) SetShardBlockFinality(finality uint32) {
	hfb.blockFinality = finality
}

// RequestMissingFinalityAttestingShardHeaders -
func (hfb *headersForBlock) RequestMissingFinalityAttestingShardHeaders() uint32 {
	return hfb.requestMissingFinalityAttestingShardHeaders()
}

// ScheduleMiniBlocksRequestIfNeeded -
func (hfb *headersForBlock) ScheduleMiniBlocksRequestIfNeeded(header data.HeaderHandler, headerHash []byte) {
	hfb.scheduleMiniBlocksRequestIfNeeded(header, headerHash)
}

// SetPendingMbRequestFallbackDelay -
func (hfb *headersForBlock) SetPendingMbRequestFallbackDelay(delay time.Duration) {
	hfb.pendingMbRequestFallbackDelay = delay
}

// SetMaxPendingMbRequests -
func (hfb *headersForBlock) SetMaxPendingMbRequests(maxPending int) {
	hfb.maxPendingMbRequests = maxPending
}

// NumPendingMbRequests -
func (hfb *headersForBlock) NumPendingMbRequests() int {
	hfb.mutPendingMbRequests.Lock()
	defer hfb.mutPendingMbRequests.Unlock()

	return len(hfb.pendingMbRequests)
}
