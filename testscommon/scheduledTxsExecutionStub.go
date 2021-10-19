package testscommon

import (
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ScheduledTxsExecutionStub -
type ScheduledTxsExecutionStub struct {
	InitCalled                            func()
	AddCalled                             func([]byte, data.TransactionHandler) bool
	ExecuteCalled                         func([]byte) error
	ExecuteAllCalled                      func(func() time.Duration) error
	GetScheduledSCRsCalled                func() map[block.Type][]data.TransactionHandler
	SetScheduledRootHashAndSCRsCalled     func(rootHash []byte, mapSCRs map[block.Type][]data.TransactionHandler)
	GetScheduledRootHashForHeaderCalled   func(headerHash []byte) ([]byte, error)
	RollBackToBlockCalled                 func(headerHash []byte) error
	GetScheduledRootHashCalled            func() []byte
	SetScheduledRootHashCalled            func([]byte)
	SetTransactionProcessorCalled         func(process.TransactionProcessor)
	SetSmartContractResultProcessorCalled func(process.SmartContractResultProcessor)
	SetTransactionCoordinatorCalled       func(process.TransactionCoordinator)
	HaveScheduledTxsCalled                func() bool
	SaveStateIfNeededCalled               func(headerHash []byte)
	SaveStateCalled                       func(headerHash []byte, scheduledRootHash []byte, mapScheduledSCRs map[block.Type][]data.TransactionHandler)
	LoadStateCalled                       func(headerHash []byte)
	IsScheduledTxCalled                   func([]byte) bool
}

// Init -
func (stes *ScheduledTxsExecutionStub) Init() {
	if stes.InitCalled != nil {
		stes.InitCalled()
	}
}

// Add -
func (stes *ScheduledTxsExecutionStub) Add(txHash []byte, tx data.TransactionHandler) bool {
	if stes.AddCalled != nil {
		return stes.AddCalled(txHash, tx)
	}
	return true
}

// Execute -
func (stes *ScheduledTxsExecutionStub) Execute(txHash []byte) error {
	if stes.ExecuteCalled != nil {
		return stes.ExecuteCalled(txHash)
	}
	return nil
}

// ExecuteAll -
func (stes *ScheduledTxsExecutionStub) ExecuteAll(haveTime func() time.Duration) error {
	if stes.ExecuteAllCalled != nil {
		return stes.ExecuteAllCalled(haveTime)
	}
	return nil
}

// GetScheduledSCRs -
func (stes *ScheduledTxsExecutionStub) GetScheduledSCRs() map[block.Type][]data.TransactionHandler {
	if stes.GetScheduledSCRsCalled != nil {
		return stes.GetScheduledSCRsCalled()
	}
	return nil
}

// SetScheduledRootHashAndSCRs -
func (stes *ScheduledTxsExecutionStub) SetScheduledRootHashAndSCRs(rootHash []byte, mapSCRs map[block.Type][]data.TransactionHandler) {
	if stes.SetScheduledRootHashAndSCRsCalled != nil {
		stes.SetScheduledRootHashAndSCRsCalled(rootHash, mapSCRs)
	}
}

// GetScheduledRootHashForHeader -
func (stes *ScheduledTxsExecutionStub) GetScheduledRootHashForHeader(headerHash []byte) ([]byte, error) {
	if stes.GetScheduledRootHashForHeaderCalled != nil {
		return stes.GetScheduledRootHashForHeaderCalled(headerHash)
	}
	return nil, errors.New("scheduled root hash for header not found")
}

// RollBackToBlock -
func (stes *ScheduledTxsExecutionStub) RollBackToBlock(headerHash []byte) error {
	if stes.RollBackToBlockCalled != nil {
		return stes.RollBackToBlockCalled(headerHash)
	}
	return nil
}

// SaveStateIfNeeded -
func (stes *ScheduledTxsExecutionStub) SaveStateIfNeeded(headerHash []byte) {
	if stes.SaveStateIfNeededCalled != nil {
		stes.SaveStateIfNeededCalled(headerHash)
	}
}

// SaveState -
func (stes *ScheduledTxsExecutionStub) SaveState(
	headerHash []byte,
	scheduledRootHash []byte,
	mapScheduledSCRs map[block.Type][]data.TransactionHandler,
) {
	if stes.SaveStateCalled != nil {
		stes.SaveStateCalled(headerHash, scheduledRootHash, mapScheduledSCRs)
	}
}

// GetScheduledRootHash -
func (stes *ScheduledTxsExecutionStub) GetScheduledRootHash() []byte {
	if stes.GetScheduledRootHashCalled != nil {
		return stes.GetScheduledRootHashCalled()
	}

	return nil
}

// SetScheduledRootHash -
func (stes *ScheduledTxsExecutionStub) SetScheduledRootHash(rootHash []byte) {
	if stes.SetScheduledRootHashCalled != nil {
		stes.SetScheduledRootHashCalled(rootHash)
	}
}

// SetTransactionProcessor -
func (stes *ScheduledTxsExecutionStub) SetTransactionProcessor(txProcessor process.TransactionProcessor) {
	if stes.SetTransactionProcessorCalled != nil {
		stes.SetTransactionProcessorCalled(txProcessor)
	}
}

// SetSmartContractResultProcessor -
func (stes *ScheduledTxsExecutionStub) SetSmartContractResultProcessor(scrProcessor process.SmartContractResultProcessor) {
	if stes.SetSmartContractResultProcessorCalled != nil {
		stes.SetSmartContractResultProcessorCalled(scrProcessor)
	}
}

// SetTransactionCoordinator -
func (stes *ScheduledTxsExecutionStub) SetTransactionCoordinator(txCoordinator process.TransactionCoordinator) {
	if stes.SetTransactionCoordinatorCalled != nil {
		stes.SetTransactionCoordinatorCalled(txCoordinator)
	}
}

// IsScheduledTx -
func (stes *ScheduledTxsExecutionStub) IsScheduledTx(txHash []byte) bool {
	if stes.IsScheduledTxCalled != nil {
		return stes.IsScheduledTxCalled(txHash)
	}
	return false
}

// IsInterfaceNil -
func (stes *ScheduledTxsExecutionStub) IsInterfaceNil() bool {
	return stes == nil
}
