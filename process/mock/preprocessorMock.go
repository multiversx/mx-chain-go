package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// PreProcessorMock -
type PreProcessorMock struct {
	CreateBlockStartedCalled              func()
	IsDataPreparedCalled                  func(requestedTxs int, haveTime func() time.Duration) error
	RemoveTxBlockFromPoolsCalled          func(body *block.Body, miniBlockPool storage.Cacher) error
	RestoreTxBlockIntoPoolsCalled         func(body *block.Body, miniBlockPool storage.Cacher) (int, error)
	SaveTxBlockToStorageCalled            func(body *block.Body) error
	ProcessBlockTransactionsCalled        func(body *block.Body, haveTime func() bool) error
	RequestBlockTransactionsCalled        func(body *block.Body) int
	CreateMarshalizedDataCalled           func(txHashes [][]byte) ([][]byte, error)
	RequestTransactionsForMiniBlockCalled func(miniBlock *block.MiniBlock) int
	ProcessMiniBlockCalled                func(miniBlock *block.MiniBlock, haveTime func() bool, getNumOfCrossInterMbsAndTxs func() (int, int)) ([][]byte, error)
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

// RemoveTxBlockFromPools -
func (ppm *PreProcessorMock) RemoveTxBlockFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	if ppm.RemoveTxBlockFromPoolsCalled == nil {
		return nil
	}
	return ppm.RemoveTxBlockFromPoolsCalled(body, miniBlockPool)
}

// RestoreTxBlockIntoPools -
func (ppm *PreProcessorMock) RestoreTxBlockIntoPools(body *block.Body, miniBlockPool storage.Cacher) (int, error) {
	if ppm.RestoreTxBlockIntoPoolsCalled == nil {
		return 0, nil
	}
	return ppm.RestoreTxBlockIntoPoolsCalled(body, miniBlockPool)
}

// SaveTxBlockToStorage -
func (ppm *PreProcessorMock) SaveTxBlockToStorage(body *block.Body) error {
	if ppm.SaveTxBlockToStorageCalled == nil {
		return nil
	}
	return ppm.SaveTxBlockToStorageCalled(body)
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
func (ppm *PreProcessorMock) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, getNumOfCrossInterMbsAndTxs func() (int, int)) ([][]byte, error) {
	if ppm.ProcessMiniBlockCalled == nil {
		return nil, nil
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
