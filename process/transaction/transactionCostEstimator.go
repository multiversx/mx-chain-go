package transaction

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const dummySignature = "01010101"

type transactionCostEstimator struct {
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
	txTypeHandler    process.TxTypeHandler
	feeHandler       process.FeeHandler
	txSimulator      facade.TransactionSimulatorProcessor
	mutExecution     sync.RWMutex
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	txSimulator facade.TransactionSimulatorProcessor,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
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
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}

	return &transactionCostEstimator{
		txTypeHandler:    txTypeHandler,
		feeHandler:       feeHandler,
		txSimulator:      txSimulator,
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
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
			GasUnits:             computeGasUnitsBasedOnVMOutput(tx, res.VMOutput),
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

func computeGasUnitsBasedOnVMOutput(tx *transaction.Transaction, vmOutput *vmcommon.VMOutput) uint64 {
	isTooMuchGasProvided := strings.Contains(vmOutput.ReturnMessage, smartContract.TooMuchGasProvidedMessage)
	if !isTooMuchGasProvided {
		return tx.GasLimit - vmOutput.GasRemaining
	}

	return tx.GasLimit - extractGasRemainedFromMessage(vmOutput.ReturnMessage)
}

func extractGasRemainedFromMessage(message string) uint64 {
	splitMessage := strings.Split(message, "gas remained = ")
	if len(splitMessage) < 2 {
		return 0
	}

	gasValue, err := strconv.ParseUint(splitMessage[1], 10, 64)
	if err != nil {
		return 0
	}

	return gasValue
}

func (tce *transactionCostEstimator) addMissingFieldsIfNeeded(tx *transaction.Transaction) {
	if tx.GasLimit == 0 {
		tx.GasLimit = tce.getTxGasLimit(tx.SndAddr)
	}
	if tx.GasPrice == 0 {
		tx.GasPrice = tce.feeHandler.MinGasPrice()
	}

	if len(tx.Signature) == 0 {
		tx.Signature = []byte(dummySignature)
	}
}

func (tce *transactionCostEstimator) getTxGasLimit(txSender []byte) (gasLimit uint64) {
	selfShardID := tce.shardCoordinator.SelfId()
	maxGasLimitPerBlock := tce.feeHandler.MaxGasLimitPerBlock(selfShardID) - 1

	gasLimit = maxGasLimitPerBlock

	senderShardID := tce.shardCoordinator.ComputeId(txSender)
	if tce.shardCoordinator.SelfId() != senderShardID {
		return
	}

	accountHandler, err := tce.accounts.LoadAccount(txSender)
	if err != nil || accountHandler == nil {
		return
	}

	accountSender, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return
	}

	maxGasLimitPerBlockBig := big.NewInt(0).SetUint64(maxGasLimitPerBlock)
	accountSenderBalance := accountSender.GetBalance()
	if accountSenderBalance.Cmp(maxGasLimitPerBlockBig) < 0 {
		gasLimit = accountSenderBalance.Uint64()
		return
	}

	return
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
