package testscommon

import (
	"errors"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-go/process"
)

// ScheduledTxsExecutionStub -
type ScheduledTxsExecutionStub struct {
	InitCalled                                   func()
	AddScheduledTxCalled                         func([]byte, data.TransactionHandler) bool
	AddScheduledMiniBlocksCalled                 func(miniBlocks block.MiniBlockSlice)
	ExecuteCalled                                func([]byte) error
	ExecuteAllCalled                             func(func() time.Duration) error
	GetScheduledIntermediateTxsCalled            func() map[block.Type][]data.TransactionHandler
	GetScheduledMiniBlocksCalled                 func() block.MiniBlockSlice
	SetScheduledInfoCalled                       func(scheduledInfo *process.ScheduledInfo)
	GetScheduledRootHashForHeaderCalled          func(headerHash []byte) ([]byte, error)
	GetScheduledRootHashForHeaderWithEpochCalled func(headerHash []byte, epoch uint32) ([]byte, error)
	RollBackToBlockCalled                        func(headerHash []byte) error
	GetScheduledRootHashCalled                   func() []byte
	GetScheduledGasAndFeesCalled                 func() scheduled.GasAndFees
	SetScheduledRootHashCalled                   func([]byte)
	SetScheduledGasAndFeesCalled                 func(gasAndFees scheduled.GasAndFees)
	SetTransactionProcessorCalled                func(process.TransactionProcessor)
	SetTransactionCoordinatorCalled              func(process.TransactionCoordinator)
	HaveScheduledTxsCalled                       func() bool
	SaveStateIfNeededCalled                      func(headerHash []byte)
	SaveStateCalled                              func(headerHash []byte, scheduledInfo *process.ScheduledInfo)
	LoadStateCalled                              func(headerHash []byte)
	IsScheduledTxCalled                          func([]byte) bool
	IsMiniBlockExecutedCalled                    func([]byte) bool
}

// Init -
func (stes *ScheduledTxsExecutionStub) Init() {
	if stes.InitCalled != nil {
		stes.InitCalled()
	}
}

// AddScheduledTx -
func (stes *ScheduledTxsExecutionStub) AddScheduledTx(txHash []byte, tx data.TransactionHandler) bool {
	if stes.AddScheduledTxCalled != nil {
		return stes.AddScheduledTxCalled(txHash, tx)
	}
	return true
}

// AddScheduledMiniBlocks -
func (stes *ScheduledTxsExecutionStub) AddScheduledMiniBlocks(miniBlocks block.MiniBlockSlice) {
	if stes.AddScheduledMiniBlocksCalled != nil {
		stes.AddScheduledMiniBlocksCalled(miniBlocks)
	}
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

// GetScheduledIntermediateTxs -
func (stes *ScheduledTxsExecutionStub) GetScheduledIntermediateTxs() map[block.Type][]data.TransactionHandler {
	if stes.GetScheduledIntermediateTxsCalled != nil {
		return stes.GetScheduledIntermediateTxsCalled()
	}
	return nil
}

// GetScheduledMiniBlocks -
func (stes *ScheduledTxsExecutionStub) GetScheduledMiniBlocks() block.MiniBlockSlice {
	if stes.GetScheduledMiniBlocksCalled != nil {
		return stes.GetScheduledMiniBlocksCalled()
	}
	return nil
}

// GetScheduledGasAndFees returns the scheduled SC calls gas and fees
func (stes *ScheduledTxsExecutionStub) GetScheduledGasAndFees() scheduled.GasAndFees {
	if stes.GetScheduledGasAndFeesCalled != nil {
		return stes.GetScheduledGasAndFeesCalled()
	}
	return scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
}

// SetScheduledInfo -
func (stes *ScheduledTxsExecutionStub) SetScheduledInfo(scheduledInfo *process.ScheduledInfo) {
	if stes.SetScheduledInfoCalled != nil {
		stes.SetScheduledInfoCalled(scheduledInfo)
	}
}

// GetScheduledRootHashForHeaderWithEpoch -
func (stes *ScheduledTxsExecutionStub) GetScheduledRootHashForHeaderWithEpoch(headerHash []byte, epoch uint32) ([]byte, error) {
	if stes.GetScheduledRootHashForHeaderWithEpochCalled != nil {
		return stes.GetScheduledRootHashForHeaderWithEpochCalled(headerHash, epoch)
	}
	return nil, errors.New("scheduled root hash for header not found")
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
	scheduledInfo *process.ScheduledInfo,
) {
	if stes.SaveStateCalled != nil {
		stes.SaveStateCalled(headerHash, scheduledInfo)
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

// SetScheduledGasAndFees -
func (stes *ScheduledTxsExecutionStub) SetScheduledGasAndFees(gasAndFees scheduled.GasAndFees) {
	if stes.SetScheduledGasAndFeesCalled != nil {
		stes.SetScheduledGasAndFeesCalled(gasAndFees)
	}
}

// SetTransactionProcessor -
func (stes *ScheduledTxsExecutionStub) SetTransactionProcessor(txProcessor process.TransactionProcessor) {
	if stes.SetTransactionProcessorCalled != nil {
		stes.SetTransactionProcessorCalled(txProcessor)
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

// IsMiniBlockExecuted -
func (stes *ScheduledTxsExecutionStub) IsMiniBlockExecuted(mbHash []byte) bool {
	if stes.IsMiniBlockExecutedCalled != nil {
		return stes.IsMiniBlockExecutedCalled(mbHash)
	}
	return false
}

// IsInterfaceNil -
func (stes *ScheduledTxsExecutionStub) IsInterfaceNil() bool {
	return stes == nil
}
