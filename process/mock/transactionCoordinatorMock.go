package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type TransactionCoordinatorMock struct {
	ComputeTransactionTypeCalled                         func(tx data.TransactionHandler) (process.TransactionType, error)
	RequestMiniBlocksCalled                              func(header data.HeaderHandler)
	RequestBlockTransactionsCalled                       func(body block.Body)
	IsDataPreparedForProcessingCalled                    func(haveTime func() time.Duration) error
	SaveBlockDataToStorageCalled                         func(body block.Body) error
	RestoreBlockDataFromStorageCalled                    func(body block.Body) (int, map[int][][]byte, error)
	RemoveBlockDataFromPoolCalled                        func(body block.Body) error
	ProcessBlockTransactionCalled                        func(body block.Body, round uint32, haveTime func() time.Duration) error
	CreateBlockStartedCalled                             func()
	CreateMbsAndProcessCrossShardTransactionsDstMeCalled func(header data.HeaderHandler, maxTxRemaining uint32, round uint32, haveTime func() bool) (block.MiniBlockSlice, uint32, bool)
	CreateMbsAndProcessTransactionsFromMeCalled          func(maxTxRemaining uint32, round uint32, haveTime func() bool) block.MiniBlockSlice
	CreateMarshalizedDataCalled                          func(body block.Body) (map[uint32]block.MiniBlockSlice, map[uint32][][]byte)
	GetAllCurrentUsedTxsCalled                           func(blockType block.Type) map[string]data.TransactionHandler
	GetPreprocessorCalled                                func(blockType block.Type) process.PreProcessor
}

func (tcm *TransactionCoordinatorMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	if tcm.ComputeTransactionTypeCalled == nil {
		return 0, nil
	}

	return tcm.ComputeTransactionTypeCalled(tx)
}

func (tcm *TransactionCoordinatorMock) RequestMiniBlocks(header data.HeaderHandler) {
	if tcm.RequestMiniBlocksCalled == nil {
		return
	}

	tcm.RequestMiniBlocksCalled(header)
}

func (tcm *TransactionCoordinatorMock) RequestBlockTransactions(body block.Body) {
	if tcm.RequestBlockTransactionsCalled == nil {
		return
	}

	tcm.RequestBlockTransactionsCalled(body)
}

func (tcm *TransactionCoordinatorMock) IsDataPreparedForProcessing(haveTime func() time.Duration) error {
	if tcm.IsDataPreparedForProcessingCalled == nil {
		return nil
	}

	return tcm.IsDataPreparedForProcessingCalled(haveTime)
}

func (tcm *TransactionCoordinatorMock) SaveBlockDataToStorage(body block.Body) error {
	if tcm.SaveBlockDataToStorageCalled == nil {
		return nil
	}

	return tcm.SaveBlockDataToStorageCalled(body)
}

func (tcm *TransactionCoordinatorMock) RestoreBlockDataFromStorage(body block.Body) (int, map[int][][]byte, error) {
	if tcm.RestoreBlockDataFromStorageCalled == nil {
		return 0, nil, nil
	}

	return tcm.RestoreBlockDataFromStorageCalled(body)
}

func (tcm *TransactionCoordinatorMock) RemoveBlockDataFromPool(body block.Body) error {
	if tcm.RemoveBlockDataFromPoolCalled == nil {
		return nil
	}

	return tcm.RemoveBlockDataFromPoolCalled(body)
}

func (tcm *TransactionCoordinatorMock) ProcessBlockTransaction(body block.Body, round uint32, haveTime func() time.Duration) error {
	if tcm.ProcessBlockTransactionCalled == nil {
		return nil
	}

	return tcm.ProcessBlockTransactionCalled(body, round, haveTime)
}

func (tcm *TransactionCoordinatorMock) CreateBlockStarted() {
	if tcm.CreateBlockStartedCalled == nil {
		return
	}

	tcm.CreateBlockStartedCalled()
}

func (tcm *TransactionCoordinatorMock) CreateMbsAndProcessCrossShardTransactionsDstMe(header data.HeaderHandler, maxTxRemaining uint32, round uint32, haveTime func() bool) (block.MiniBlockSlice, uint32, bool) {
	if tcm.CreateMbsAndProcessCrossShardTransactionsDstMeCalled == nil {
		return nil, 0, false
	}

	return tcm.CreateMbsAndProcessCrossShardTransactionsDstMeCalled(header, maxTxRemaining, round, haveTime)
}

func (tcm *TransactionCoordinatorMock) CreateMbsAndProcessTransactionsFromMe(maxTxRemaining uint32, round uint32, haveTime func() bool) block.MiniBlockSlice {
	if tcm.CreateMbsAndProcessTransactionsFromMeCalled == nil {
		return nil
	}

	return tcm.CreateMbsAndProcessTransactionsFromMeCalled(maxTxRemaining, round, haveTime)
}

func (tcm *TransactionCoordinatorMock) CreateMarshalizedData(body block.Body) (map[uint32]block.MiniBlockSlice, map[uint32][][]byte) {
	if tcm.CreateMarshalizedDataCalled == nil {
		return make(map[uint32]block.MiniBlockSlice), make(map[uint32][][]byte)
	}

	return tcm.CreateMarshalizedDataCalled(body)
}

func (tcm *TransactionCoordinatorMock) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	if tcm.GetAllCurrentUsedTxsCalled == nil {
		return nil
	}

	return tcm.GetAllCurrentUsedTxsCalled(blockType)
}

func (tcm *TransactionCoordinatorMock) GetPreprocessor(blockType block.Type) process.PreProcessor {
	if tcm.GetPreprocessorCalled == nil {
		return nil
	}

	return tcm.GetPreprocessorCalled(blockType)
}
