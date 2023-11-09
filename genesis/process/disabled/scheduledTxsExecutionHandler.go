package disabled

import (
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-go/process"
)

// ScheduledTxsExecutionHandler implements ScheduledTxsExecutionHandler interface but does nothing as it is a disabled component
type ScheduledTxsExecutionHandler struct {
}

// Init does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Init() {
}

// AddScheduledTx does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) AddScheduledTx(_ []byte, _ data.TransactionHandler) bool {
	return true
}

// AddScheduledMiniBlocks does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) AddScheduledMiniBlocks(_ block.MiniBlockSlice) {
}

// Execute does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Execute(_ []byte) error {
	return nil
}

// ExecuteAll does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) ExecuteAll(_ func() time.Duration) error {
	return nil
}

// GetScheduledIntermediateTxs does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) GetScheduledIntermediateTxs() map[block.Type][]data.TransactionHandler {
	return make(map[block.Type][]data.TransactionHandler)
}

// GetScheduledMiniBlocks does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) GetScheduledMiniBlocks() block.MiniBlockSlice {
	return make(block.MiniBlockSlice, 0)
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

// SetScheduledInfo does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) SetScheduledInfo(_ *process.ScheduledInfo) {
}

// GetScheduledRootHashForHeader does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) GetScheduledRootHashForHeader(_ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// GetScheduledRootHashForHeaderWithEpoch does nothing as it is disabled
func (steh *ScheduledTxsExecutionHandler) GetScheduledRootHashForHeaderWithEpoch(_ []byte, _ uint32) ([]byte, error) {
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
	_ *process.ScheduledInfo,
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

// IsMiniBlockExecuted always returns false as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) IsMiniBlockExecuted(_ []byte) bool {
	return false
}

// IsInterfaceNil returns true if underlying object is nil
func (steh *ScheduledTxsExecutionHandler) IsInterfaceNil() bool {
	return steh == nil
}
