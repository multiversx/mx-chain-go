package testscommon

import (
	"errors"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ScheduledTxsExecutionStub -
type ScheduledTxsExecutionStub struct {
	InitCalled                               func()
	AddCalled                                func([]byte, data.TransactionHandler) bool
	AddMiniBlocksCalled                      func(miniBlocks block.MiniBlockSlice)
	ExecuteCalled                            func([]byte) error
	ExecuteAllCalled                         func(func() time.Duration) error
	GetScheduledSCRsCalled                   func() map[block.Type][]data.TransactionHandler
	SetScheduledRootHashSCRsGasAndFeesCalled func(rootHash []byte, mapSCRs map[block.Type][]data.TransactionHandler, gasAndFees scheduled.GasAndFees)
	GetScheduledRootHashForHeaderCalled      func(headerHash []byte) ([]byte, error)
	RollBackToBlockCalled                    func(headerHash []byte) error
	GetScheduledRootHashCalled               func() []byte
	GetScheduledGasAndFeesCalled             func() scheduled.GasAndFees
	SetScheduledRootHashCalled               func([]byte)
	SetScheduledGasAndFeesCalled             func(gasAndFees scheduled.GasAndFees)
	SetTransactionProcessorCalled            func(process.TransactionProcessor)
	SetTransactionCoordinatorCalled          func(process.TransactionCoordinator)
	HaveScheduledTxsCalled                   func() bool
	SaveStateIfNeededCalled                  func(headerHash []byte)
	SaveStateCalled                          func(headerHash []byte, scheduledRootHash []byte, mapScheduledSCRs map[block.Type][]data.TransactionHandler, gasAndFees scheduled.GasAndFees)
	LoadStateCalled                          func(headerHash []byte)
	IsScheduledTxCalled                      func([]byte) bool
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

// AddMiniBlocks -
func (stes *ScheduledTxsExecutionStub) AddMiniBlocks(miniBlocks block.MiniBlockSlice) {
	if stes.AddMiniBlocksCalled != nil {
		stes.AddMiniBlocksCalled(miniBlocks)
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

// GetScheduledSCRs -
func (stes *ScheduledTxsExecutionStub) GetScheduledSCRs() map[block.Type][]data.TransactionHandler {
	if stes.GetScheduledSCRsCalled != nil {
		return stes.GetScheduledSCRsCalled()
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

// SetScheduledRootHashSCRsGasAndFees -
func (stes *ScheduledTxsExecutionStub) SetScheduledRootHashSCRsGasAndFees(rootHash []byte, mapSCRs map[block.Type][]data.TransactionHandler, gasAndFees scheduled.GasAndFees) {
	if stes.SetScheduledRootHashSCRsGasAndFeesCalled != nil {
		stes.SetScheduledRootHashSCRsGasAndFeesCalled(rootHash, mapSCRs, gasAndFees)
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
	gasAndFees scheduled.GasAndFees,
) {
	if stes.SaveStateCalled != nil {
		stes.SaveStateCalled(headerHash, scheduledRootHash, mapScheduledSCRs, gasAndFees)
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

// IsInterfaceNil -
func (stes *ScheduledTxsExecutionStub) IsInterfaceNil() bool {
	return stes == nil
}
