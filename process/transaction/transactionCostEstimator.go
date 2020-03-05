package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
)

type transactionCostEstimator struct {
	txTypeHandler      process.TxTypeHandler
	feeHandler         process.FeeHandler
	query              external.SCQueryService
	storePerByteCost   uint64
	compilePerByteCost uint64
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	query external.SCQueryService,
	gasSchedule map[string]map[string]uint64,
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

	compileCost, storeCost := getOperationCost(gasSchedule)

	return &transactionCostEstimator{
		txTypeHandler:      txTypeHandler,
		feeHandler:         feeHandler,
		query:              query,
		storePerByteCost:   compileCost,
		compilePerByteCost: storeCost,
	}, nil
}

func getOperationCost(gasSchedule map[string]map[string]uint64) (uint64, uint64) {
	baseOpMap, ok := gasSchedule[core.BaseOperationCost]
	if !ok {
		return 0, 0
	}

	storeCost, ok := baseOpMap["StorePerByte"]
	if !ok {
		return 0, 0
	}

	compilerCost, ok := baseOpMap["CompilePerByte"]
	if !ok {
		return 0, 0
	}

	return storeCost, compilerCost
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	txType, err := tce.txTypeHandler.ComputeTransactionType(tx)
	if err != nil {
		return 0, err
	}

	tx.GasPrice = 1

	switch txType {
	case process.MoveBalance:
		return tce.feeHandler.ComputeGasLimit(tx), nil
	case process.SCDeployment:
		gasLimit := uint64(len(tx.Data)) * (tce.storePerByteCost + tce.compilePerByteCost)
		return gasLimit, nil
	case process.SCInvoking:
		return tce.query.ComputeScCallGasLimit(tx)
	default:
		return 0, process.ErrWrongTransaction
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
