package metachain

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// TransactionFeeHandler is an empty struct which implements TransactionFeeHandler interface
type TransactionFeeHandler struct {
}

// MaxGasLimitPerBlock will return the maximum gas limit to exist in a block
func (t *TransactionFeeHandler) MaxGasLimitPerBlock() uint64 {
	return 1500000000
}

// ComputeGasLimit will return 0
func (t *TransactionFeeHandler) ComputeGasLimit(_ process.TransactionWithFeeHandler) uint64 {
	return 0
}

// ComputeFee will return 0
func (t *TransactionFeeHandler) ComputeFee(_ process.TransactionWithFeeHandler) *big.Int {
	return big.NewInt(0)
}

// CheckValidityTxValues will return nil for any tx
func (t *TransactionFeeHandler) CheckValidityTxValues(_ process.TransactionWithFeeHandler) error {
	return nil
}

// ProcessTransactionFee empty cost processing for metachain
func (t *TransactionFeeHandler) ProcessTransactionFee(_ *big.Int) {
}

// IsInterfaceNil verifies if underlying struct is nil
func (t *TransactionFeeHandler) IsInterfaceNil() bool {
	return t == nil
}
