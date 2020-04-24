package disabled

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// DisabledFeeHandler represents a disabled fee handler implementation
type DisabledFeeHandler struct {
}

// DeveloperPercentage returns 0
func (dfh *DisabledFeeHandler) DeveloperPercentage() float64 {
	return 0
}

// MinGasPrice returns 0
func (dfh *DisabledFeeHandler) MinGasPrice() uint64 {
	return 0
}

// MaxGasLimitPerBlock return max uint64
func (dfh *DisabledFeeHandler) MaxGasLimitPerBlock() uint64 {
	return math.MaxUint64
}

// ComputeGasLimit returns 0
func (dfh *DisabledFeeHandler) ComputeGasLimit(_ process.TransactionWithFeeHandler) uint64 {
	return 0
}

// ComputeFee returns 0
func (dfh *DisabledFeeHandler) ComputeFee(_ process.TransactionWithFeeHandler) *big.Int {
	return big.NewInt(0)
}

// CheckValidityTxValues returns nil
func (dfh *DisabledFeeHandler) CheckValidityTxValues(_ process.TransactionWithFeeHandler) error {
	return nil
}

// CreateBlockStarted does nothing
func (dfh *DisabledFeeHandler) CreateBlockStarted() {
}

// GetAccumulatedFees returns 0
func (dfh *DisabledFeeHandler) GetAccumulatedFees() *big.Int {
	return big.NewInt(0)
}

// ProcessTransactionFee does nothing
func (dfh *DisabledFeeHandler) ProcessTransactionFee(_ *big.Int, _ []byte) {
}

// RevertFees does nothing
func (dfh *DisabledFeeHandler) RevertFees(_ [][]byte) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (dfh *DisabledFeeHandler) IsInterfaceNil() bool {
	return dfh == nil
}
