package disabled

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ScheduledTxsExecutionHandler implements ScheduledTxsExecutionHandler interface but does nothing as it is a disabled component
type ScheduledTxsExecutionHandler struct {
}

// Init does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Init() {
}

// Add does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Add(_ []byte, _ data.TransactionHandler) bool {
	return true
}

// AddMiniBlocks does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) AddMiniBlocks(_ block.MiniBlockSlice) {
}

// Execute does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Execute(_ []byte) error {
	return nil
}

// ExecuteAll does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) ExecuteAll(_ func() time.Duration) error {
	return nil
}

// GetScheduledSCRs does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) GetScheduledSCRs() map[block.Type][]data.TransactionHandler {
	return make(map[block.Type][]data.TransactionHandler)
}

// GetScheduledGasAndFees returns a zero value structure for the gas and fees
func (steh *ScheduledTxsExecutionHandler) GetScheduledGasAndFees() scheduled.GasAndFees {
	return scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
}

// SetScheduledRootHashSCRsGasAndFees does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) SetScheduledRootHashSCRsGasAndFees(_ []byte, _ map[block.Type][]data.TransactionHandler, _ scheduled.GasAndFees) {
}

// GetScheduledRootHashForHeader does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) GetScheduledRootHashForHeader(_ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// RollBackToBlock does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) RollBackToBlock(_ []byte) error {
	return nil
}

// SaveStateIfNeeded does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) SaveStateIfNeeded(_ []byte) {
}

// SaveState does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) SaveState(
	_ []byte,
	_ []byte,
	_ map[block.Type][]data.TransactionHandler,
	_ scheduled.GasAndFees,
) {
}

// GetScheduledRootHash does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) GetScheduledRootHash() []byte {
	return nil
}

// SetScheduledRootHash does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) SetScheduledRootHash(_ []byte) {
}

// SetScheduledGasAndFees does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) SetScheduledGasAndFees(_ scheduled.GasAndFees) {
}

// SetTransactionProcessor does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) SetTransactionProcessor(_ process.TransactionProcessor) {
}

// SetTransactionCoordinator does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) SetTransactionCoordinator(_ process.TransactionCoordinator) {
}

// IsScheduledTx always returns false as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) IsScheduledTx(_ []byte) bool {
	return false
}

// IsInterfaceNil returns true if underlying object is nil
func (steh *ScheduledTxsExecutionHandler) IsInterfaceNil() bool {
	return steh == nil
}
