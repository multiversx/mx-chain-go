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

// GenesisTotalSupply -
func (fh *FeeHandler) GenesisTotalSupply() *big.Int {
	return big.NewInt(0)
}

// MinGasPrice returns 0
func (fh *FeeHandler) MinGasPrice() uint64 {
	return 0
}

// MinGasLimit will return min gas limit
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

// GetDeveloperFees -
func (fh *FeeHandler) GetDeveloperFees() *big.Int {
	return big.NewInt(0)
}

// GenesisTotalSupply -
func (fh *FeeHandler) GenesisTotalSupply() *big.Int {
	return big.NewInt(0)
}

// GasPerDataByte -
func (fh *FeeHandler) GasPerDataByte() uint64 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (fh *FeeHandler) IsInterfaceNil() bool {
	return fh == nil
}
