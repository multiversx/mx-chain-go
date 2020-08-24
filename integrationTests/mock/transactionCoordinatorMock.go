package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TransactionCoordinatorMock -
type TransactionCoordinatorMock struct {
	ComputeTransactionTypeCalled                         func(tx data.TransactionHandler) process.TransactionType
	RequestMiniBlocksCalled                              func(header data.HeaderHandler)
	RequestBlockTransactionsCalled                       func(body *block.Body)
	IsDataPreparedForProcessingCalled                    func(haveTime func() time.Duration) error
	SaveBlockDataToStorageCalled                         func(body *block.Body) error
	RestoreBlockDataFromStorageCalled                    func(body *block.Body) (int, error)
	RemoveBlockDataFromPoolCalled                        func(body *block.Body) error
	ProcessBlockTransactionCalled                        func(body *block.Body, haveTime func() time.Duration) error
	CreateBlockStartedCalled                             func()
	CreateMbsAndProcessCrossShardTransactionsDstMeCalled func(header data.HeaderHandler,
		processedMiniBlocksHashes map[string]struct{},

		haveTime func() bool,
	) (block.MiniBlockSlice, uint32, bool, error)
	CreateMbsAndProcessTransactionsFromMeCalled func(haveTime func() bool) block.MiniBlockSlice
	CreateMarshalizedDataCalled                 func(body *block.Body) map[string][][]byte
	GetAllCurrentUsedTxsCalled                  func(blockType block.Type) map[string]data.TransactionHandler
	VerifyCreatedBlockTransactionsCalled        func(hdr data.HeaderHandler, body *block.Body) error
	CreatePostProcessMiniBlocksCalled           func() block.MiniBlockSlice
	CreateMarshalizedReceiptsCalled             func() ([]byte, []byte, error)
}

// CreatePostProcessMiniBlocks -
func (tcm *TransactionCoordinatorMock) CreatePostProcessMiniBlocks() block.MiniBlockSlice {
	if tcm.CreatePostProcessMiniBlocksCalled != nil {
		return tcm.CreatePostProcessMiniBlocksCalled()
	}
	return nil
}

// CreateReceiptsHash -
func (tcm *TransactionCoordinatorMock) CreateReceiptsHash() ([]byte, error) {
	return []byte("receiptHash"), nil
}

// ComputeTransactionType -
func (tcm *TransactionCoordinatorMock) ComputeTransactionType(tx data.TransactionHandler) process.TransactionType {
	if tcm.ComputeTransactionTypeCalled == nil {
		return 0
	}

	return tcm.ComputeTransactionTypeCalled(tx)
}

// RequestMiniBlocks -
func (tcm *TransactionCoordinatorMock) RequestMiniBlocks(header data.HeaderHandler) {
	if tcm.RequestMiniBlocksCalled == nil {
		return
	}

	tcm.RequestMiniBlocksCalled(header)
}

// RequestBlockTransactions -
func (tcm *TransactionCoordinatorMock) RequestBlockTransactions(body *block.Body) {
	if tcm.RequestBlockTransactionsCalled == nil {
		return
	}

	tcm.RequestBlockTransactionsCalled(body)
}

// IsDataPreparedForProcessing -
func (tcm *TransactionCoordinatorMock) IsDataPreparedForProcessing(haveTime func() time.Duration) error {
	if tcm.IsDataPreparedForProcessingCalled == nil {
		return nil
	}

	return tcm.IsDataPreparedForProcessingCalled(haveTime)
}

// SaveBlockDataToStorage -
func (tcm *TransactionCoordinatorMock) SaveBlockDataToStorage(body *block.Body) error {
	if tcm.SaveBlockDataToStorageCalled == nil {
		return nil
	}

	return tcm.SaveBlockDataToStorageCalled(body)
}

// RestoreBlockDataFromStorage -
func (tcm *TransactionCoordinatorMock) RestoreBlockDataFromStorage(body *block.Body) (int, error) {
	if tcm.RestoreBlockDataFromStorageCalled == nil {
		return 0, nil
	}

	return tcm.RestoreBlockDataFromStorageCalled(body)
}

// RemoveBlockDataFromPool -
func (tcm *TransactionCoordinatorMock) RemoveBlockDataFromPool(body *block.Body) error {
	if tcm.RemoveBlockDataFromPoolCalled == nil {
		return nil
	}

	return tcm.RemoveBlockDataFromPoolCalled(body)
}

// ProcessBlockTransaction -
func (tcm *TransactionCoordinatorMock) ProcessBlockTransaction(body *block.Body, haveTime func() time.Duration) error {
	if tcm.ProcessBlockTransactionCalled == nil {
		return nil
	}

	return tcm.ProcessBlockTransactionCalled(body, haveTime)
}

// CreateBlockStarted -
func (tcm *TransactionCoordinatorMock) CreateBlockStarted() {
	if tcm.CreateBlockStartedCalled == nil {
		return
	}

	tcm.CreateBlockStartedCalled()
}

// CreateMbsAndProcessCrossShardTransactionsDstMe -
func (tcm *TransactionCoordinatorMock) CreateMbsAndProcessCrossShardTransactionsDstMe(
	header data.HeaderHandler,
	processedMiniBlocksHashes map[string]struct{},

	haveTime func() bool,
) (block.MiniBlockSlice, uint32, bool, error) {
	if tcm.CreateMbsAndProcessCrossShardTransactionsDstMeCalled == nil {
		return nil, 0, false, nil
	}

	return tcm.CreateMbsAndProcessCrossShardTransactionsDstMeCalled(header, processedMiniBlocksHashes, haveTime)
}

// CreateMbsAndProcessTransactionsFromMe -
func (tcm *TransactionCoordinatorMock) CreateMbsAndProcessTransactionsFromMe(haveTime func() bool) block.MiniBlockSlice {
	if tcm.CreateMbsAndProcessTransactionsFromMeCalled == nil {
		return nil
	}

	return tcm.CreateMbsAndProcessTransactionsFromMeCalled(haveTime)
}

// CreateMarshalizedData -
func (tcm *TransactionCoordinatorMock) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	if tcm.CreateMarshalizedDataCalled == nil {
		return make(map[string][][]byte)
	}

	return tcm.CreateMarshalizedDataCalled(body)
}

// GetAllCurrentUsedTxs -
func (tcm *TransactionCoordinatorMock) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	if tcm.GetAllCurrentUsedTxsCalled == nil {
		return nil
	}

	return tcm.GetAllCurrentUsedTxsCalled(blockType)
}

// VerifyCreatedBlockTransactions -
func (tcm *TransactionCoordinatorMock) VerifyCreatedBlockTransactions(hdr data.HeaderHandler, body *block.Body) error {
	if tcm.VerifyCreatedBlockTransactionsCalled == nil {
		return nil
	}

	return tcm.VerifyCreatedBlockTransactionsCalled(hdr, body)
}

// CreateMarshalizedReceipts -
func (tcm *TransactionCoordinatorMock) CreateMarshalizedReceipts() ([]byte, []byte, error) {
	if tcm.CreateMarshalizedReceiptsCalled == nil {
		return nil, nil, nil
	}

	return tcm.CreateMarshalizedReceiptsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tcm *TransactionCoordinatorMock) IsInterfaceNil() bool {
	return tcm == nil
}
