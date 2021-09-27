package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// PreProcessorMock -
type PreProcessorMock struct {
	CreateBlockStartedCalled              func()
	IsDataPreparedCalled                  func(requestedTxs int, haveTime func() time.Duration) error
	RemoveMiniBlocksFromPoolsCalled       func(body *block.Body, miniBlockPool storage.Cacher) error
	RemoveTxsFromPoolsCalled              func(body *block.Body) error
	RestoreMiniBlocksIntoPoolsCalled      func(body *block.Body, miniBlockPool storage.Cacher) error
	RestoreTxsIntoPoolsCalled             func(body *block.Body) (int, error)
	SaveTxsToStorageCalled                func(body *block.Body) error
	ProcessBlockTransactionsCalled        func(header data.HeaderHandler, body *block.Body, haveTime func() bool, scheduledMode bool, gasConsumedInfo *process.GasConsumedInfo) error
	RequestBlockTransactionsCalled        func(body *block.Body) int
	CreateMarshalizedDataCalled           func(txHashes [][]byte) ([][]byte, error)
	RequestTransactionsForMiniBlockCalled func(miniBlock *block.MiniBlock) int
	ProcessMiniBlockCalled                func(miniBlock *block.MiniBlock, haveTime func() bool, haveAdditionalTime func() bool, getNumOfCrossInterMbsAndTxs func() (int, int), scheduledMode bool, isNewMiniBlock bool, gasConsumedInfo *process.GasConsumedInfo) ([][]byte, int, error)
	CreateAndProcessMiniBlocksCalled      func(haveTime func() bool) (block.MiniBlockSlice, error)
	GetAllCurrentUsedTxsCalled            func() map[string]data.TransactionHandler
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

// RemoveMiniBlocksFromPools -
func (ppm *PreProcessorMock) RemoveMiniBlocksFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	if ppm.RemoveMiniBlocksFromPoolsCalled == nil {
		return nil
	}
	return ppm.RemoveMiniBlocksFromPoolsCalled(body, miniBlockPool)
}

// RemoveTxsFromPools -
func (ppm *PreProcessorMock) RemoveTxsFromPools(body *block.Body) error {
	if ppm.RemoveTxsFromPoolsCalled == nil {
		return nil
	}
	return ppm.RemoveTxsFromPoolsCalled(body)
}

// RestoreMiniBlocksIntoPools -
func (ppm *PreProcessorMock) RestoreMiniBlocksIntoPools(body *block.Body, miniBlockPool storage.Cacher) error {
	if ppm.RestoreMiniBlocksIntoPoolsCalled == nil {
		return nil
	}
	return ppm.RestoreMiniBlocksIntoPoolsCalled(body, miniBlockPool)
}

// RestoreTxsIntoPools -
func (ppm *PreProcessorMock) RestoreTxsIntoPools(body *block.Body) (int, error) {
	if ppm.RestoreTxsIntoPoolsCalled == nil {
		return 0, nil
	}
	return ppm.RestoreTxsIntoPoolsCalled(body)
}

// SaveTxsToStorage -
func (ppm *PreProcessorMock) SaveTxsToStorage(body *block.Body) error {
	if ppm.SaveTxsToStorageCalled == nil {
		return nil
	}
	return ppm.SaveTxsToStorageCalled(body)
}

// ProcessBlockTransactions -
func (ppm *PreProcessorMock) ProcessBlockTransactions(header data.HeaderHandler, body *block.Body, haveTime func() bool, scheduledMode bool, gasConsumedInfo *process.GasConsumedInfo) error {
	if ppm.ProcessBlockTransactionsCalled == nil {
		return nil
	}
	return ppm.ProcessBlockTransactionsCalled(header, body, haveTime, scheduledMode, gasConsumedInfo)
}

// RequestBlockTransactions -
func (ppm *PreProcessorMock) RequestBlockTransactions(body *block.Body) int {
	if ppm.RequestBlockTransactionsCalled == nil {
		return 0
	}
	return ppm.RequestBlockTransactionsCalled(body)
}

// CreateMarshalizedData -
func (ppm *PreProcessorMock) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	if ppm.CreateMarshalizedDataCalled == nil {
		return nil, nil
	}
	return ppm.CreateMarshalizedDataCalled(txHashes)
}

// RequestTransactionsForMiniBlock -
func (ppm *PreProcessorMock) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if ppm.RequestTransactionsForMiniBlockCalled == nil {
		return 0
	}
	return ppm.RequestTransactionsForMiniBlockCalled(miniBlock)
}

// ProcessMiniBlock -
func (ppm *PreProcessorMock) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, haveAdditionalTime func() bool, getNumOfCrossInterMbsAndTxs func() (int, int), scheduledMode bool, isNewMiniBlock bool, gasConsumedInfo *process.GasConsumedInfo) ([][]byte, int, error) {
	if ppm.ProcessMiniBlockCalled == nil {
		return nil, 0, nil
	}
	return ppm.ProcessMiniBlockCalled(miniBlock, haveTime, haveAdditionalTime, getNumOfCrossInterMbsAndTxs, scheduledMode, isNewMiniBlock, gasConsumedInfo)
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (ppm *PreProcessorMock) CreateAndProcessMiniBlocks(haveTime func() bool) (block.MiniBlockSlice, error) {
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

// IsInterfaceNil returns true if there is no value under the interface
func (ppm *PreProcessorMock) IsInterfaceNil() bool {
	return ppm == nil
}
