package transaction

import (
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
)

type transactionCostEstimator struct {
	txTypeHandler process.TxTypeHandler
	feeHandler    process.FeeHandler
	mutExecution  sync.RWMutex
	txSimulator   facade.TransactionSimulatorProcessor
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	txSimulator facade.TransactionSimulatorProcessor,
) (*transactionCostEstimator, error) {
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(txSimulator) {
		return nil, txsimulator.ErrNilTxSimulatorProcessor
	}

	return &transactionCostEstimator{
		txTypeHandler: txTypeHandler,
		feeHandler:    feeHandler,
		txSimulator:   txSimulator,
	}, nil
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	tce.mutExecution.RLock()
	defer tce.mutExecution.RUnlock()

	txType, _ := tce.txTypeHandler.ComputeTransactionType(tx)

	switch txType {
	case process.MoveBalance:
		return &transaction.CostResponse{
			GasUnits: tce.feeHandler.ComputeGasLimit(tx),
		}, nil
	case process.SCDeployment, process.SCInvoking, process.BuiltInFunctionCall:
		return tce.simulateTransactionCost(tx)
	case process.RelayedTx, process.RelayedTxV2:
		// TODO implement in the next PR
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: "not implemented for relayed transactions",
		}, nil
	default:
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: process.ErrWrongTransaction.Error(),
		}, nil
	}
}

func (tce *transactionCostEstimator) simulateTransactionCost(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	if tx.GasLimit == 0 {
		tx.GasLimit = math.MaxUint64 - 1
	}
	if tx.GasPrice == 0 {
		tx.GasPrice = tce.feeHandler.MinGasPrice()
	}

	if len(tx.Signature) == 0 {
		tx.Signature = []byte("01010101")
	}

	res, err := tce.txSimulator.ProcessTx(tx)
	if err != nil {
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: err.Error(),
		}, nil
	}

	if res.VMOutput == nil {
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: "no vm output",
		}, nil

	}

	if res.VMOutput.ReturnCode == vmcommon.Ok {
		return &transaction.CostResponse{
			GasUnits:   tx.GasLimit - res.VMOutput.GasRemaining,
			RetMessage: "",
		}, nil
	}

	return &transaction.CostResponse{
		GasUnits:   0,
		RetMessage: res.VMOutput.ReturnMessage,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
