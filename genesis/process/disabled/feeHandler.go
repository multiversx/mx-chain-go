package disabled

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// FeeHandler represents a disabled fee handler implementation
type FeeHandler struct {
}

// GasPriceModifier returns 1.0
func (fh *FeeHandler) GasPriceModifier() float64 {
	return 1.0
}

// DeveloperPercentage returns 0
func (fh *FeeHandler) DeveloperPercentage() float64 {
	return 0
}

// GenesisTotalSupply returns 0
func (fh *FeeHandler) GenesisTotalSupply() *big.Int {
	return big.NewInt(0)
}

// ComputeGasLimitBasedOnBalance return 0 and nil
func (fh *FeeHandler) ComputeGasLimitBasedOnBalance(_ process.TransactionWithFeeHandler, _ *big.Int) (uint64, error) {
	return 0, nil
}

// MinGasPrice returns 0
func (fh *FeeHandler) MinGasPrice() uint64 {
	return 0
}

// MinGasLimit returns 0
func (fh *FeeHandler) MinGasLimit() uint64 {
	return 0
}

// MaxGasLimitPerBlock return max uint64
func (fh *FeeHandler) MaxGasLimitPerBlock(uint32) uint64 {
	return math.MaxUint64
}

// ComputeGasLimit returns 0
func (fh *FeeHandler) ComputeGasLimit(_ process.TransactionWithFeeHandler) uint64 {
	return 0
}

// ComputeMoveBalanceFee returns 0
func (fh *FeeHandler) ComputeMoveBalanceFee(_ process.TransactionWithFeeHandler) *big.Int {
	return big.NewInt(0)
}

// ComputeFeeForProcessing returns 0
func (fh *FeeHandler) ComputeFeeForProcessing(_ process.TransactionWithFeeHandler, _ uint64) *big.Int {
	return big.NewInt(0)
}

// ComputeTxFee returns 0
func (fh *FeeHandler) ComputeTxFee(_ process.TransactionWithFeeHandler) *big.Int {
	return big.NewInt(0)
}

// CheckValidityTxValues returns nil
func (fh *FeeHandler) CheckValidityTxValues(_ process.TransactionWithFeeHandler) error {
	return nil
}

// CreateBlockStarted does nothing
func (fh *FeeHandler) CreateBlockStarted() {
}

// GetAccumulatedFees returns 0
func (fh *FeeHandler) GetAccumulatedFees() *big.Int {
	return big.NewInt(0)
}

// ProcessTransactionFee does nothing
func (fh *FeeHandler) ProcessTransactionFee(_ *big.Int, _ *big.Int, _ []byte) {
}

// RevertFees does nothing
func (fh *FeeHandler) RevertFees(_ [][]byte) {
}

// GetDeveloperFees returns 0
func (fh *FeeHandler) GetDeveloperFees() *big.Int {
	return big.NewInt(0)
}

// GasPerDataByte returns 0
func (fh *FeeHandler) GasPerDataByte() uint64 {
	return 0
}

// SplitTxGasInCategories returns 0, 0
func (fh *FeeHandler) SplitTxGasInCategories(_ process.TransactionWithFeeHandler) (uint64, uint64) {
	return 0, 0
}

// GasPriceForProcessing return 0
func (fh *FeeHandler) GasPriceForProcessing(_ process.TransactionWithFeeHandler) uint64 {
	return 0
}

// GasPriceForMove returns 0
func (fh *FeeHandler) GasPriceForMove(_ process.TransactionWithFeeHandler) uint64 {
	return 0
}

// MinGasPriceForProcessing returns 0
func (fh *FeeHandler) MinGasPriceForProcessing() uint64 {
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValue returns 0, 0
func (fh *FeeHandler) ComputeGasUsedAndFeeBasedOnRefundValue(_ process.TransactionWithFeeHandler, _ *big.Int) (uint64, *big.Int) {
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsed returns 0
func (fh *FeeHandler) ComputeTxFeeBasedOnGasUsed(_ process.TransactionWithFeeHandler, _ uint64) *big.Int {
	return big.NewInt(0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fh *FeeHandler) IsInterfaceNil() bool {
	return fh == nil
}
