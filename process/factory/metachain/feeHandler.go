package metachain

import "math/big"

type TransactionFeeHandler struct {
}

func (t *TransactionFeeHandler) ProcessTransactionFee(cost *big.Int) {
}

func (t *TransactionFeeHandler) IsInterfaceNil() bool {
	if t == nil {
		return true
	}
	return false
}
