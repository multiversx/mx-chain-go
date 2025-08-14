package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
)

// TransactionCostEstimatorMock  -
type TransactionCostEstimatorMock struct {
	ComputeTransactionGasLimitCalled   func(tx *transaction.Transaction) (*transaction.CostResponse, error)
	SimulateTransactionExecutionCalled func(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error)
	SimulateSCRExecutionCostCalled     func(scr *smartContractResult.SmartContractResult) (*transaction.CostResponse, error)
}

// SimulateSCRExecutionCost -
func (tcem *TransactionCostEstimatorMock) SimulateSCRExecutionCost(scr *smartContractResult.SmartContractResult) (*transaction.CostResponse, error) {
	if tcem.SimulateSCRExecutionCostCalled != nil {
		return tcem.SimulateSCRExecutionCostCalled(scr)
	}

	return nil, nil
}

// ComputeTransactionGasLimit -
func (tcem *TransactionCostEstimatorMock) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	if tcem.ComputeTransactionGasLimitCalled != nil {
		return tcem.ComputeTransactionGasLimitCalled(tx)
	}
	return &transaction.CostResponse{}, nil
}

// SimulateTransactionExecution -
func (tcem *TransactionCostEstimatorMock) SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
	if tcem.SimulateTransactionExecutionCalled != nil {
		return tcem.SimulateTransactionExecutionCalled(tx)
	}

	return &txSimData.SimulationResultsWithVMOutput{}, nil
}

// IsInterfaceNil -
func (tcem *TransactionCostEstimatorMock) IsInterfaceNil() bool {
	return tcem == nil
}
