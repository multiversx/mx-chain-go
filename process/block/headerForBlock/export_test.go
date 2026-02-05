package headerForBlock

import (
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
