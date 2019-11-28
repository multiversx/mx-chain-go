package metachain

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// TransactionFeeHandler is an empty struct which implements TransactionFeeHandler interface
type TransactionFeeHandler struct {
}

func (t *TransactionFeeHandler) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	return 0
}

func (t *TransactionFeeHandler) ComputeFee(tx process.TransactionWithFeeHandler) *big.Int {
	return big.NewInt(0)
}

func (t *TransactionFeeHandler) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	return nil
}

// ProcessTransactionFee empty cost processing for metachain
func (t *TransactionFeeHandler) ProcessTransactionFee(cost *big.Int) {
}

// IsInterfaceNil verifies if underlying struct is nil
func (t *TransactionFeeHandler) IsInterfaceNil() bool {
	if t == nil {
		return true
	}
	return false
}
