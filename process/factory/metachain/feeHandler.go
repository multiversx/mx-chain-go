package metachain

import "math/big"

// TransactionFeeHandler is an empty struct which implements TransactionFeeHandler interface
type TransactionFeeHandler struct {
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
