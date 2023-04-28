package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

// TransactionCoordinatorMock -
type TransactionCoordinatorMock struct {
	ComputeTransactionTypeCalled                         func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType)
	RequestMiniBlocksAndTransactionsCalled               func(header data.HeaderHandler)
	RequestBlockTransactionsCalled                       func(body *block.Body)
	IsDataPreparedForProcessingCalled                    func(haveTime func() time.Duration) error
	SaveTxsToStorageCalled                               func(body *block.Body)
	RestoreBlockDataFromStorageCalled                    func(body *block.Body) (int, error)
	RemoveBlockDataFromPoolCalled                        func(body *block.Body) error
	RemoveTxsFromPoolCalled                              func(body *block.Body) error
	ProcessBlockTransactionCalled                        func(header data.HeaderHandler, body *block.Body, haveTime func() time.Duration) error
	CreateBlockStartedCalled                             func()
	CreateMbsAndProcessCrossShardTransactionsDstMeCalled func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo, haveTime func() bool, haveAdditionalTime func() bool, scheduledMode bool) (block.MiniBlockSlice, uint32, bool, error)
	CreateMbsAndProcessTransactionsFromMeCalled          func(haveTime func() bool) block.MiniBlockSlice
	CreateMarshalizedDataCalled                          func(body *block.Body) map[string][][]byte
	GetCreatedInShardMiniBlocksCalled                    func() []*block.MiniBlock
	GetAllCurrentUsedTxsCalled                           func(blockType block.Type) map[string]data.TransactionHandler
	VerifyCreatedBlockTransactionsCalled                 func(hdr data.HeaderHandler, body *block.Body) error
	CreatePostProcessMiniBlocksCalled                    func() block.MiniBlockSlice
	VerifyCreatedMiniBlocksCalled                        func(hdr data.HeaderHandler, body *block.Body) error
	AddIntermediateTransactionsCalled                    func(mapSCRs map[block.Type][]data.TransactionHandler) error
	GetAllIntermediateTxsCalled                          func() map[block.Type]map[string]data.TransactionHandler
	AddTxsFromMiniBlocksCalled                           func(miniBlocks block.MiniBlockSlice)
	AddTransactionsCalled                                func(txHandlers []data.TransactionHandler, blockType block.Type)
}

// GetAllCurrentLogs -
func (tcm *TransactionCoordinatorMock) GetAllCurrentLogs() []*data.LogData {
	return nil
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
func (tcm *TransactionCoordinatorMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
	if tcm.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, process.MoveBalance
	}

	return tcm.ComputeTransactionTypeCalled(tx)
}

// RequestMiniBlocksAndTransactions -
func (tcm *TransactionCoordinatorMock) RequestMiniBlocksAndTransactions(header data.HeaderHandler) {
	if tcm.RequestMiniBlocksAndTransactionsCalled == nil {
		return
	}

	tcm.RequestMiniBlocksAndTransactionsCalled(header)
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

// SaveTxsToStorage -
func (tcm *TransactionCoordinatorMock) SaveTxsToStorage(body *block.Body) {
	if tcm.SaveTxsToStorageCalled == nil {
		return
	}

	tcm.SaveTxsToStorageCalled(body)
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

// RemoveTxsFromPool -
func (tcm *TransactionCoordinatorMock) RemoveTxsFromPool(body *block.Body) error {
	if tcm.RemoveTxsFromPoolCalled == nil {
		return nil
	}

	return tcm.RemoveTxsFromPoolCalled(body)
}

// ProcessBlockTransaction -
func (tcm *TransactionCoordinatorMock) ProcessBlockTransaction(header data.HeaderHandler, body *block.Body, haveTime func() time.Duration) error {
	if tcm.ProcessBlockTransactionCalled == nil {
		return nil
	}

	return tcm.ProcessBlockTransactionCalled(header, body, haveTime)
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
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
) (block.MiniBlockSlice, uint32, bool, error) {
	if tcm.CreateMbsAndProcessCrossShardTransactionsDstMeCalled == nil {
		return nil, 0, false, nil
	}

	return tcm.CreateMbsAndProcessCrossShardTransactionsDstMeCalled(header, processedMiniBlocksInfo, haveTime, haveAdditionalTime, scheduledMode)
}

// CreateMbsAndProcessTransactionsFromMe -
func (tcm *TransactionCoordinatorMock) CreateMbsAndProcessTransactionsFromMe(haveTime func() bool, _ []byte) block.MiniBlockSlice {
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

// GetCreatedInShardMiniBlocks -
func (tcm *TransactionCoordinatorMock) GetCreatedInShardMiniBlocks() []*block.MiniBlock {
	if tcm.GetCreatedInShardMiniBlocksCalled != nil {
		return tcm.GetCreatedInShardMiniBlocksCalled()
	}

	return nil
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

// VerifyCreatedMiniBlocks -
func (tcm *TransactionCoordinatorMock) VerifyCreatedMiniBlocks(hdr data.HeaderHandler, body *block.Body) error {
	if tcm.VerifyCreatedMiniBlocksCalled == nil {
		return nil
	}

	return tcm.VerifyCreatedMiniBlocksCalled(hdr, body)
}

// AddIntermediateTransactions -
func (tcm *TransactionCoordinatorMock) AddIntermediateTransactions(mapSCRs map[block.Type][]data.TransactionHandler) error {
	if tcm.AddIntermediateTransactionsCalled == nil {
		return nil
	}

	return tcm.AddIntermediateTransactionsCalled(mapSCRs)
}

// GetAllIntermediateTxs -
func (tcm *TransactionCoordinatorMock) GetAllIntermediateTxs() map[block.Type]map[string]data.TransactionHandler {
	if tcm.GetAllIntermediateTxsCalled == nil {
		return nil
	}

	return tcm.GetAllIntermediateTxsCalled()
}

// AddTxsFromMiniBlocks -
func (tcm *TransactionCoordinatorMock) AddTxsFromMiniBlocks(miniBlocks block.MiniBlockSlice) {
	if tcm.AddTxsFromMiniBlocksCalled == nil {
		return
	}

	tcm.AddTxsFromMiniBlocksCalled(miniBlocks)
}

// AddTransactions -
func (tcm *TransactionCoordinatorMock) AddTransactions(txHandlers []data.TransactionHandler, blockType block.Type) {
	if tcm.AddTransactionsCalled == nil {
		return
	}
	tcm.AddTransactionsCalled(txHandlers, blockType)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tcm *TransactionCoordinatorMock) IsInterfaceNil() bool {
	return tcm == nil
}
