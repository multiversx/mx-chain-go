package mock

import "github.com/multiversx/mx-chain-core-go/data/transaction"

// TransactionCostEstimatorMock  --
type TransactionCostEstimatorMock struct {
	ComputeTransactionGasLimitCalled func(tx *transaction.Transaction) (*transaction.CostResponse, error)
}

// ComputeTransactionGasLimit --
func (tcem *TransactionCostEstimatorMock) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	if tcem.ComputeTransactionGasLimitCalled != nil {
		return tcem.ComputeTransactionGasLimitCalled(tx)
	}
	return &transaction.CostResponse{}, nil
}

// IsInterfaceNil --
func (tcem *TransactionCostEstimatorMock) IsInterfaceNil() bool {
	return tcem == nil
}
