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

// CrossShardMetaDataMock -
type CrossShardMetaDataMock struct {
	GetNonceCalled      func() uint64
	GetShardIdCalled    func() uint32
	GetHeaderHashCalled func() []byte
}

// GetNonce -
func (crossShardMetaDataMock *CrossShardMetaDataMock) GetNonce() uint64 {
	if crossShardMetaDataMock.GetNonceCalled != nil {
		return crossShardMetaDataMock.GetNonceCalled()
	}

	return 0
}

// GetShardID -
func (crossShardMetaDataMock *CrossShardMetaDataMock) GetShardID() uint32 {
	if crossShardMetaDataMock.GetShardIdCalled != nil {
		return crossShardMetaDataMock.GetShardIdCalled()
	}

	return 0
}

// GetHeaderHash -
func (crossShardMetaDataMock *CrossShardMetaDataMock) GetHeaderHash() []byte {
	if crossShardMetaDataMock.GetHeaderHashCalled != nil {
		return crossShardMetaDataMock.GetHeaderHashCalled()
	}
	return nil
}
