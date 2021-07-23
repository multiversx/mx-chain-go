package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// PreProcessorMock -
type PreProcessorMock struct {
	CreateBlockStartedCalled              func()
	IsDataPreparedCalled                  func(requestedTxs int, haveTime func() time.Duration) error
	RemoveBlockDataFromPoolsCalled        func(body *block.Body, miniBlockPool storage.Cacher) error
	RemoveTxsFromPoolsCalled              func(body *block.Body) error
	RestoreBlockDataIntoPoolsCalled       func(body *block.Body, miniBlockPool storage.Cacher) (int, error)
	SaveTxsToStorageCalled                func(body *block.Body) error
	ProcessBlockTransactionsCalled        func(body *block.Body, haveTime func() bool) error
	RequestBlockTransactionsCalled        func(body *block.Body) int
	CreateMarshalizedDataCalled           func(txHashes [][]byte) ([][]byte, error)
	RequestTransactionsForMiniBlockCalled func(miniBlock *block.MiniBlock) int
	ProcessMiniBlockCalled                func(miniBlock *block.MiniBlock, haveTime func() bool, getNumOfCrossInterMbsAndTxs func() (int, int)) ([][]byte, int, error)
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
func (ppm *PreProcessorMock) ProcessBlockTransactions(body *block.Body, haveTime func() bool) error {
	if ppm.ProcessBlockTransactionsCalled == nil {
		return nil
	}
	return ppm.ProcessBlockTransactionsCalled(body, haveTime)
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
func (ppm *PreProcessorMock) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, getNumOfCrossInterMbsAndTxs func() (int, int)) ([][]byte, int, error) {
	if ppm.ProcessMiniBlockCalled == nil {
		return nil, 0, nil
	}
	return ppm.ProcessMiniBlockCalled(miniBlock, haveTime, getNumOfCrossInterMbsAndTxs)
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
