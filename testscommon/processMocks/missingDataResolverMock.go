package processMocks

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// MissingDataResolverMock -
type MissingDataResolverMock struct {
	RequestMissingMetaHeadersBlockingCalled           func(shardHeader data.ShardHeaderHandler, timeout time.Duration) error
	RequestMissingMetaHeadersCalled                   func(shardHeader data.ShardHeaderHandler) error
	WaitForMissingDataCalled                          func(timeout time.Duration) error
	RequestBlockTransactionsCalled                    func(body *block.Body)
	RequestMiniBlocksAndTransactionsCalled            func(header data.HeaderHandler)
	GetFinalCrossMiniBlockInfoAndRequestMissingCalled func(header data.HeaderHandler) []*data.MiniBlockInfo
	RequestMissingShardHeadersCalled                  func(header data.MetaHeaderHandler) error
	ResetCalled                                       func()
}

// RequestMissingMetaHeadersBlocking -
func (mock *MissingDataResolverMock) RequestMissingMetaHeadersBlocking(shardHeader data.ShardHeaderHandler, timeout time.Duration) error {
	if mock.RequestMissingMetaHeadersBlockingCalled != nil {
		return mock.RequestMissingMetaHeadersBlockingCalled(shardHeader, timeout)
	}
	return nil
}

// RequestMissingMetaHeaders -
func (mock *MissingDataResolverMock) RequestMissingMetaHeaders(shardHeader data.ShardHeaderHandler) error {
	if mock.RequestMissingMetaHeadersCalled != nil {
		return mock.RequestMissingMetaHeadersCalled(shardHeader)
	}
	return nil
}

// WaitForMissingData -
func (mock *MissingDataResolverMock) WaitForMissingData(timeout time.Duration) error {
	if mock.WaitForMissingDataCalled != nil {
		return mock.WaitForMissingDataCalled(timeout)
	}
	return nil
}

// RequestBlockTransactions -
func (mock *MissingDataResolverMock) RequestBlockTransactions(body *block.Body) {
	if mock.RequestBlockTransactionsCalled != nil {
		mock.RequestBlockTransactionsCalled(body)
	}
}

// RequestMiniBlocksAndTransactions -
func (mock *MissingDataResolverMock) RequestMiniBlocksAndTransactions(header data.HeaderHandler) {
	if mock.RequestMiniBlocksAndTransactionsCalled != nil {
		mock.RequestMiniBlocksAndTransactionsCalled(header)
	}
}

// GetFinalCrossMiniBlockInfoAndRequestMissing -
func (mock *MissingDataResolverMock) GetFinalCrossMiniBlockInfoAndRequestMissing(header data.HeaderHandler) []*data.MiniBlockInfo {
	if mock.GetFinalCrossMiniBlockInfoAndRequestMissingCalled != nil {
		return mock.GetFinalCrossMiniBlockInfoAndRequestMissingCalled(header)
	}
	return make([]*data.MiniBlockInfo, 0)
}

// RequestMissingShardHeaders -
func (mock *MissingDataResolverMock) RequestMissingShardHeaders(header data.MetaHeaderHandler) error {
	if mock.RequestMissingShardHeadersCalled != nil {
		return mock.RequestMissingShardHeadersCalled(header)
	}
	return nil
}

// Reset -
func (mock *MissingDataResolverMock) Reset() {
	if mock.ResetCalled != nil {
		mock.ResetCalled()
	}
}

// IsInterfaceNil -
func (mock *MissingDataResolverMock) IsInterfaceNil() bool {
	return mock == nil
}
