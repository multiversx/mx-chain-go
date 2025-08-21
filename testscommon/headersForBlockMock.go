package testscommon

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
)

// HeadersForBlockMock -
type HeadersForBlockMock struct {
	AddHeaderCalled              func(hash string, header data.HeaderHandler, usedInBlock bool, hasProof bool, hasProofRequested bool)
	RequestShardHeadersCalled    func(metaBlock data.MetaHeaderHandler)
	RequestMetaHeadersCalled     func(shardHeader data.ShardHeaderHandler)
	WaitForHeadersIfNeededCalled func(haveTime func() time.Duration) error
	GetHeaderInfoCalled          func(hash string) (headerForBlock.HeaderInfo, bool)
	GetHeadersInfoMapCalled      func() map[string]headerForBlock.HeaderInfo
	GetHeadersMapCalled          func() map[string]data.HeaderHandler
	GetMissingDataCalled         func() (uint32, uint32, uint32)
	ResetCalled                  func()
}

// AddHeader -
func (mock *HeadersForBlockMock) AddHeader(hash string, header data.HeaderHandler, usedInBlock bool, hasProof bool, hasProofRequested bool) {
	if mock.AddHeaderCalled != nil {
		mock.AddHeaderCalled(hash, header, usedInBlock, hasProof, hasProofRequested)
	}
}

// RequestShardHeaders -
func (mock *HeadersForBlockMock) RequestShardHeaders(metaBlock data.MetaHeaderHandler) {
	if mock.RequestShardHeadersCalled != nil {
		mock.RequestShardHeadersCalled(metaBlock)
	}
}

// RequestMetaHeaders -
func (mock *HeadersForBlockMock) RequestMetaHeaders(shardHeader data.ShardHeaderHandler) {
	if mock.RequestMetaHeadersCalled != nil {
		mock.RequestMetaHeadersCalled(shardHeader)
	}
}

// WaitForHeadersIfNeeded -
func (mock *HeadersForBlockMock) WaitForHeadersIfNeeded(haveTime func() time.Duration) error {
	if mock.WaitForHeadersIfNeededCalled != nil {
		return mock.WaitForHeadersIfNeededCalled(haveTime)
	}

	return nil
}

// GetHeaderInfo -
func (mock *HeadersForBlockMock) GetHeaderInfo(hash string) (headerForBlock.HeaderInfo, bool) {
	if mock.GetHeaderInfoCalled != nil {
		return mock.GetHeaderInfoCalled(hash)
	}

	return nil, false
}

// GetHeadersInfoMap -
func (mock *HeadersForBlockMock) GetHeadersInfoMap() map[string]headerForBlock.HeaderInfo {
	if mock.GetHeadersInfoMapCalled != nil {
		return mock.GetHeadersInfoMapCalled()
	}

	return nil
}

// GetHeadersMap -
func (mock *HeadersForBlockMock) GetHeadersMap() map[string]data.HeaderHandler {
	if mock.GetHeadersMapCalled != nil {
		return mock.GetHeadersMapCalled()
	}

	return nil
}

// GetMissingData -
func (mock *HeadersForBlockMock) GetMissingData() (uint32, uint32, uint32) {
	if mock.GetMissingDataCalled != nil {
		return mock.GetMissingDataCalled()
	}

	return 0, 0, 0
}

// Reset -
func (mock *HeadersForBlockMock) Reset() {
	if mock.ResetCalled != nil {
		mock.ResetCalled()
	}
}

// IsInterfaceNil -
func (mock *HeadersForBlockMock) IsInterfaceNil() bool {
	return mock == nil
}
