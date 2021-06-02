package transaction

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
)

const dummySignature = "01010101"

type transactionCostEstimator struct {
	txTypeHandler process.TxTypeHandler
	feeHandler    process.FeeHandler
	txSimulator   facade.TransactionSimulatorProcessor
	mutExecution  sync.RWMutex
	selfShardID   uint32
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	txSimulator facade.TransactionSimulatorProcessor,
	selfShardID uint32,
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
		selfShardID:   selfShardID,
	}, nil
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	tce.mutExecution.RLock()
	defer tce.mutExecution.RUnlock()

	txType, _ := tce.txTypeHandler.ComputeTransactionType(tx)

	switch txType {
	case process.SCDeployment, process.SCInvoking, process.BuiltInFunctionCall, process.MoveBalance:
		return tce.simulateTransactionCost(tx, txType)
	case process.RelayedTx, process.RelayedTxV2:
		// TODO implement in the next PR
		return &transaction.CostResponse{
			GasUnits:      0,
			ReturnMessage: "cannot compute cost of the relayed transaction",
		}, nil
	default:
		return &transaction.CostResponse{
			GasUnits:      0,
			ReturnMessage: process.ErrWrongTransaction.Error(),
		}, nil
	}
}

func (tce *transactionCostEstimator) simulateTransactionCost(tx *transaction.Transaction, txType process.TransactionType) (*transaction.CostResponse, error) {
	tce.addMissingFieldsIfNeeded(tx)

	res, err := tce.txSimulator.ProcessTx(tx)
	if err != nil {
		return &transaction.CostResponse{
			GasUnits:      0,
			ReturnMessage: err.Error(),
		}, nil
	}

	isMoveBalanceOk := txType == process.MoveBalance && res.FailReason == ""
	if isMoveBalanceOk {
		return &transaction.CostResponse{
			GasUnits:      tce.feeHandler.ComputeGasLimit(tx),
			ReturnMessage: "",
		}, nil

	}

	if res.FailReason != "" {
		return &transaction.CostResponse{
			GasUnits:      0,
			ReturnMessage: res.FailReason,
		}, nil
	}

	if res.VMOutput == nil {
		return &transaction.CostResponse{
			GasUnits:             0,
			ReturnMessage:        process.ErrNilVMOutput.Error(),
			SmartContractResults: nil,
		}, nil
	}

	if res.VMOutput.ReturnCode == vmcommon.Ok {
		return &transaction.CostResponse{
			GasUnits:             tx.GasLimit - res.VMOutput.GasRemaining,
			ReturnMessage:        "",
			SmartContractResults: res.ScResults,
		}, nil
	}

	return &transaction.CostResponse{
		GasUnits:             0,
		ReturnMessage:        fmt.Sprintf("%s %s", res.VMOutput.ReturnCode.String(), res.VMOutput.ReturnMessage),
		SmartContractResults: res.ScResults,
	}, nil
}

func (tce *transactionCostEstimator) addMissingFieldsIfNeeded(tx *transaction.Transaction) {
	if tx.GasLimit == 0 {
		tx.GasLimit = tce.feeHandler.MaxGasLimitPerBlock(tce.selfShardID) - 1
	}
	if tx.GasPrice == 0 {
		tx.GasPrice = tce.feeHandler.MinGasPrice()
	}

	if len(tx.Signature) == 0 {
		tx.Signature = []byte(dummySignature)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
