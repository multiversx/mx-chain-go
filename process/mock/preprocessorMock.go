package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

// PreProcessorMock -
type PreProcessorMock struct {
	CreateBlockStartedCalled              func()
	IsDataPreparedCalled                  func(requestedTxs int, haveTime func() time.Duration) error
	RemoveBlockDataFromPoolsCalled        func(body *block.Body, miniBlockPool storage.Cacher) error
	RemoveTxsFromPoolsCalled              func(body *block.Body) error
	RestoreBlockDataIntoPoolsCalled       func(body *block.Body, miniBlockPool storage.Cacher) (int, error)
	SaveTxsToStorageCalled                func(body *block.Body) error
	ProcessBlockTransactionsCalled        func(header data.HeaderHandler, body *block.Body, haveTime func() bool) error
	RequestBlockTransactionsCalled        func(body *block.Body) int
	CreateMarshalledDataCalled            func(txHashes [][]byte) ([][]byte, error)
	RequestTransactionsForMiniBlockCalled func(miniBlock *block.MiniBlock) int
	ProcessMiniBlockCalled                func(miniBlock *block.MiniBlock, haveTime func() bool, haveAdditionalTime func() bool, scheduledMode bool, partialMbExecutionMode bool, indexOfLastTxProcessed int, preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler) ([][]byte, int, bool, error)
	CreateAndProcessMiniBlocksCalled      func(haveTime func() bool) (block.MiniBlockSlice, error)
	GetAllCurrentUsedTxsCalled            func() map[string]data.TransactionHandler
	AddTxsFromMiniBlocksCalled            func(miniBlocks block.MiniBlockSlice)
	AddTransactionsCalled                 func(txHandlers []data.TransactionHandler)
}

// CreateBlockStarted -
func (ppm *PreProcessorMock) CreateBlockStarted() {
	if ppm.CreateBlockStartedCalled == nil {
		return
	}
	ppm.CreateBlockStartedCalled()
}

// IsDataPrepared -
func (ppm *PreProcessorMock) IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error {
	if ppm.IsDataPreparedCalled == nil {
		return nil
	}
	return ppm.IsDataPreparedCalled(requestedTxs, haveTime)
}

// RemoveBlockDataFromPools -
func (ppm *PreProcessorMock) RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	if ppm.RemoveBlockDataFromPoolsCalled == nil {
		return nil
	}
	return ppm.RemoveBlockDataFromPoolsCalled(body, miniBlockPool)
}

// RemoveTxsFromPools -
func (ppm *PreProcessorMock) RemoveTxsFromPools(body *block.Body) error {
	if ppm.RemoveTxsFromPoolsCalled == nil {
		return nil
	}
	return ppm.RemoveTxsFromPoolsCalled(body)
}

// RestoreBlockDataIntoPools -
func (ppm *PreProcessorMock) RestoreBlockDataIntoPools(body *block.Body, miniBlockPool storage.Cacher) (int, error) {
	if ppm.RestoreBlockDataIntoPoolsCalled == nil {
		return 0, nil
	}
	return ppm.RestoreBlockDataIntoPoolsCalled(body, miniBlockPool)
}

// SaveTxsToStorage -
func (ppm *PreProcessorMock) SaveTxsToStorage(body *block.Body) error {
	if ppm.SaveTxsToStorageCalled == nil {
		return nil
	}
	return ppm.SaveTxsToStorageCalled(body)
}

// ProcessBlockTransactions -
func (ppm *PreProcessorMock) ProcessBlockTransactions(header data.HeaderHandler, body *block.Body, haveTime func() bool) error {
	if ppm.ProcessBlockTransactionsCalled == nil {
		return nil
	}
	return ppm.ProcessBlockTransactionsCalled(header, body, haveTime)
}

// RequestBlockTransactions -
func (ppm *PreProcessorMock) RequestBlockTransactions(body *block.Body) int {
	if ppm.RequestBlockTransactionsCalled == nil {
		return 0
	}
	return ppm.RequestBlockTransactionsCalled(body)
}

// CreateMarshalledData -
func (ppm *PreProcessorMock) CreateMarshalledData(txHashes [][]byte) ([][]byte, error) {
	if ppm.CreateMarshalledDataCalled == nil {
		return nil, nil
	}
	return ppm.CreateMarshalledDataCalled(txHashes)
}

// RequestTransactionsForMiniBlock -
func (ppm *PreProcessorMock) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if ppm.RequestTransactionsForMiniBlockCalled == nil {
		return 0
	}
	return ppm.RequestTransactionsForMiniBlockCalled(miniBlock)
}

// ProcessMiniBlock -
func (ppm *PreProcessorMock) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
	partialMbExecutionMode bool,
	indexOfLastTxProcessed int,
	preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler,
) ([][]byte, int, bool, error) {
	if ppm.ProcessMiniBlockCalled == nil {
		return nil, 0, false, nil
	}
	return ppm.ProcessMiniBlockCalled(miniBlock, haveTime, haveAdditionalTime, scheduledMode, partialMbExecutionMode, indexOfLastTxProcessed, preProcessorExecutionInfoHandler)
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (ppm *PreProcessorMock) CreateAndProcessMiniBlocks(haveTime func() bool, _ []byte) (block.MiniBlockSlice, error) {
	if ppm.CreateAndProcessMiniBlocksCalled == nil {
		return nil, nil
	}
	return ppm.CreateAndProcessMiniBlocksCalled(haveTime)
}

// GetAllCurrentUsedTxs -
func (ppm *PreProcessorMock) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	if ppm.GetAllCurrentUsedTxsCalled == nil {
		return nil
	}
	return ppm.GetAllCurrentUsedTxsCalled()
}

// AddTxsFromMiniBlocks -
func (ppm *PreProcessorMock) AddTxsFromMiniBlocks(miniBlocks block.MiniBlockSlice) {
	if ppm.AddTxsFromMiniBlocksCalled == nil {
		return
	}
	ppm.AddTxsFromMiniBlocksCalled(miniBlocks)
}

// AddTransactions -
func (ppm *PreProcessorMock) AddTransactions(txHandlers []data.TransactionHandler) {
	if ppm.AddTransactionsCalled == nil {
		return
	}
	ppm.AddTransactionsCalled(txHandlers)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppm *PreProcessorMock) IsInterfaceNil() bool {
	return ppm == nil
}
