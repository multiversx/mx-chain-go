package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
)

type transactionCostEstimator struct {
	txTypeHandler process.TxTypeHandler
	feeHandler    process.FeeHandler
	query         external.SCQueryService
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	query external.SCQueryService,
) (*transactionCostEstimator, error) {
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(query) {
		return nil, external.ErrNilSCQueryService
	}

	return &transactionCostEstimator{
		txTypeHandler: txTypeHandler,
		feeHandler:    feeHandler,
		query:         query,
	}, nil
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	txType, err := tce.txTypeHandler.ComputeTransactionType(tx)
	if err != nil {
		return 0, err
	}

	tx.GasPrice = 1

	switch txType {
	case process.MoveBalance, process.SCDeployment:
		return tce.feeHandler.ComputeFee(tx).Uint64(), nil
	case process.SCInvoking:
		return tce.query.ComputeScCallCost(tx)
	default:
		return 0, process.ErrWrongTransaction
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
