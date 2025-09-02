package preprocMocks

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// BlockDataRequesterStub -
type BlockDataRequesterStub struct {
	RequestBlockTransactionsCalled                    func(body *block.Body)
	RequestMiniBlocksAndTransactionsCalled            func(header data.HeaderHandler)
	GetFinalCrossMiniBlockInfoAndRequestMissingCalled func(header data.HeaderHandler) []*data.MiniBlockInfo
	IsDataPreparedForProcessingCalled                 func(haveTime func() time.Duration) error
}

// RequestBlockTransactions -
func (bdr *BlockDataRequesterStub) RequestBlockTransactions(body *block.Body) {
	if bdr.RequestBlockTransactionsCalled != nil {
		bdr.RequestBlockTransactionsCalled(body)
	}
}

// RequestMiniBlocksAndTransactions -
func (bdr *BlockDataRequesterStub) RequestMiniBlocksAndTransactions(header data.HeaderHandler) {
	if bdr.RequestMiniBlocksAndTransactionsCalled != nil {
		bdr.RequestMiniBlocksAndTransactionsCalled(header)
	}
}

// GetFinalCrossMiniBlockInfoAndRequestMissing -
func (bdr *BlockDataRequesterStub) GetFinalCrossMiniBlockInfoAndRequestMissing(header data.HeaderHandler) []*data.MiniBlockInfo {
	if bdr.GetFinalCrossMiniBlockInfoAndRequestMissingCalled != nil {
		return bdr.GetFinalCrossMiniBlockInfoAndRequestMissingCalled(header)
	}
	return nil
}

// IsDataPreparedForProcessing -
func (bdr *BlockDataRequesterStub) IsDataPreparedForProcessing(haveTime func() time.Duration) error {
	if bdr.IsDataPreparedForProcessingCalled != nil {
		return bdr.IsDataPreparedForProcessingCalled(haveTime)
	}
	return nil
}

// IsInterfaceNil -
func (bdr *BlockDataRequesterStub) IsInterfaceNil() bool {
	return bdr == nil
}
