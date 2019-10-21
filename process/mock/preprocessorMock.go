package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type PreProcessorMock struct {
	CreateBlockStartedCalled              func()
	IsDataPreparedCalled                  func(requestedTxs int, haveTime func() time.Duration) error
	RemoveTxBlockFromPoolsCalled          func(body block.Body, miniBlockPool storage.Cacher) error
	RestoreTxBlockIntoPoolsCalled         func(body block.Body, miniBlockPool storage.Cacher) (int, error)
	SaveTxBlockToStorageCalled            func(body block.Body) error
	ProcessBlockTransactionsCalled        func(body block.Body, round uint64, haveTime func() bool) error
	RequestBlockTransactionsCalled        func(body block.Body) int
	CreateMarshalizedDataCalled           func(txHashes [][]byte) ([][]byte, error)
	RequestTransactionsForMiniBlockCalled func(miniBlock *block.MiniBlock) int
	ProcessMiniBlockCalled                func(miniBlock *block.MiniBlock, haveTime func() bool, round uint64) error
	CreateAndProcessMiniBlocksCalled      func(maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, round uint64, haveTime func() bool) (block.MiniBlockSlice, error)
	CreateAndProcessMiniBlockCalled       func(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint64) (*block.MiniBlock, error)
	GetAllCurrentUsedTxsCalled            func() map[string]data.TransactionHandler
}

func (ppm *PreProcessorMock) CreateBlockStarted() {
	if ppm.CreateBlockStartedCalled == nil {
		return
	}
	ppm.CreateBlockStartedCalled()
}

func (ppm *PreProcessorMock) IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error {
	if ppm.IsDataPreparedCalled == nil {
		return nil
	}
	return ppm.IsDataPreparedCalled(requestedTxs, haveTime)
}

func (ppm *PreProcessorMock) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if ppm.RemoveTxBlockFromPoolsCalled == nil {
		return nil
	}
	return ppm.RemoveTxBlockFromPoolsCalled(body, miniBlockPool)
}

func (ppm *PreProcessorMock) RestoreTxBlockIntoPools(body block.Body, miniBlockPool storage.Cacher) (int, error) {
	if ppm.RestoreTxBlockIntoPoolsCalled == nil {
		return 0, nil
	}
	return ppm.RestoreTxBlockIntoPoolsCalled(body, miniBlockPool)
}

func (ppm *PreProcessorMock) SaveTxBlockToStorage(body block.Body) error {
	if ppm.SaveTxBlockToStorageCalled == nil {
		return nil
	}
	return ppm.SaveTxBlockToStorageCalled(body)
}

func (ppm *PreProcessorMock) ProcessBlockTransactions(body block.Body, round uint64, haveTime func() bool) error {
	if ppm.ProcessBlockTransactionsCalled == nil {
		return nil
	}
	return ppm.ProcessBlockTransactionsCalled(body, round, haveTime)
}

func (ppm *PreProcessorMock) RequestBlockTransactions(body block.Body) int {
	if ppm.RequestBlockTransactionsCalled == nil {
		return 0
	}
	return ppm.RequestBlockTransactionsCalled(body)
}

func (ppm *PreProcessorMock) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	if ppm.CreateMarshalizedDataCalled == nil {
		return nil, nil
	}
	return ppm.CreateMarshalizedDataCalled(txHashes)
}

func (ppm *PreProcessorMock) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if ppm.RequestTransactionsForMiniBlockCalled == nil {
		return 0
	}
	return ppm.RequestTransactionsForMiniBlockCalled(miniBlock)
}

func (ppm *PreProcessorMock) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint64) error {
	if ppm.ProcessMiniBlockCalled == nil {
		return nil
	}
	return ppm.ProcessMiniBlockCalled(miniBlock, haveTime, round)
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (ppm *PreProcessorMock) CreateAndProcessMiniBlocks(
	maxTxSpaceRemained uint32,
	maxMbSpaceRemained uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, error) {
	if ppm.CreateAndProcessMiniBlocksCalled == nil {
		return nil, nil
	}
	return ppm.CreateAndProcessMiniBlocksCalled(maxTxSpaceRemained, maxMbSpaceRemained, round, haveTime)
}

func (ppm *PreProcessorMock) CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint64) (*block.MiniBlock, error) {
	if ppm.CreateAndProcessMiniBlockCalled == nil {
		return nil, nil
	}
	return ppm.CreateAndProcessMiniBlockCalled(sndShardId, dstShardId, spaceRemained, haveTime, round)
}

func (ppm *PreProcessorMock) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	if ppm.GetAllCurrentUsedTxsCalled == nil {
		return nil
	}
	return ppm.GetAllCurrentUsedTxsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppm *PreProcessorMock) IsInterfaceNil() bool {
	if ppm == nil {
		return true
	}
	return false
}
