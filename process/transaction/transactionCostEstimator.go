package transaction

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-go/common"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/txsimulator"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const dummySignature = "01010101"
const gasRemainedSplitString = "gas remained = "
const gasUsedSlitString = "gas used = "

type transactionCostEstimator struct {
	accounts            state.AccountsAdapter
	shardCoordinator    sharding.Coordinator
	txTypeHandler       process.TxTypeHandler
	feeHandler          process.FeeHandler
	txSimulator         facade.TransactionSimulatorProcessor
	enableEpochsHandler common.EnableEpochsHandler
	mutExecution        sync.RWMutex
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	txSimulator facade.TransactionSimulatorProcessor,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	enableEpochsHandler common.EnableEpochsHandler,
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
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	tce := &transactionCostEstimator{
		txTypeHandler:       txTypeHandler,
		feeHandler:          feeHandler,
		txSimulator:         txSimulator,
		accounts:            accounts,
		shardCoordinator:    shardCoordinator,
		enableEpochsHandler: enableEpochsHandler,
	}

	return tce, nil
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	tce.mutExecution.RLock()
	defer tce.mutExecution.RUnlock()

	txTypeOnSender, txTypeOnDestination := tce.txTypeHandler.ComputeTransactionType(tx)
	if txTypeOnSender == process.MoveBalance && txTypeOnDestination == process.MoveBalance {
		return tce.computeMoveBalanceCost(tx), nil
	}

	switch txTypeOnSender {
	case process.SCDeployment, process.SCInvoking, process.BuiltInFunctionCall, process.MoveBalance:
		return tce.simulateTransactionCost(tx, txTypeOnSender)
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

func (tce *transactionCostEstimator) computeMoveBalanceCost(tx *transaction.Transaction) *transaction.CostResponse {
	gasUnits := tce.feeHandler.ComputeGasLimit(tx)

	return &transaction.CostResponse{
		GasUnits:      gasUnits,
		ReturnMessage: "",
	}
}

func (tce *transactionCostEstimator) simulateTransactionCost(tx *transaction.Transaction, txType process.TransactionType) (*transaction.CostResponse, error) {
	err := tce.addMissingFieldsIfNeeded(tx)
	if err != nil {
		return nil, err
	}

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
			GasUnits:             tce.computeGasUnitsBasedOnVMOutput(tx, res.VMOutput),
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

func (tce *transactionCostEstimator) computeGasUnitsBasedOnVMOutput(tx *transaction.Transaction, vmOutput *vmcommon.VMOutput) uint64 {
	isTooMuchGasProvided := strings.Contains(vmOutput.ReturnMessage, smartContract.TooMuchGasProvidedMessage)
	if !isTooMuchGasProvided {
		return tx.GasLimit - vmOutput.GasRemaining
	}

	isTooMuchGasV2MsgFlagSet := tce.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabled()
	if isTooMuchGasV2MsgFlagSet {
		gasNeededForProcessing := extractGasRemainedFromMessage(vmOutput.ReturnMessage, gasUsedSlitString)
		return tce.feeHandler.ComputeGasLimit(tx) + gasNeededForProcessing
	}

	return tx.GasLimit - extractGasRemainedFromMessage(vmOutput.ReturnMessage, gasRemainedSplitString)
}

func extractGasRemainedFromMessage(message string, splitString string) uint64 {
	splitMessage := strings.Split(message, splitString)
	if len(splitMessage) < 2 {
		log.Warn("extractGasRemainedFromMessage", "error", "cannot split message", "message", message)
		return 0
	}

	gasValue, err := strconv.ParseUint(splitMessage[1], 10, 64)
	if err != nil {
		log.Warn("extractGasRemainedFromMessage", "error", err.Error())
		return 0
	}

	return gasValue
}

func (tce *transactionCostEstimator) addMissingFieldsIfNeeded(tx *transaction.Transaction) error {
	if tx.GasPrice == 0 {
		tx.GasPrice = tce.feeHandler.MinGasPrice()
	}
	if len(tx.Signature) == 0 {
		tx.Signature = []byte(dummySignature)
	}
	if tx.GasLimit == 0 {
		var err error
		tx.GasLimit, err = tce.getTxGasLimit(tx)

		return err
	}

	return nil
}

func (tce *transactionCostEstimator) getTxGasLimit(tx *transaction.Transaction) (uint64, error) {
	selfShardID := tce.shardCoordinator.SelfId()
	maxGasLimitPerBlock := tce.feeHandler.MaxGasLimitPerBlock(selfShardID) - 1

	senderShardID := tce.shardCoordinator.ComputeId(tx.SndAddr)
	if tce.shardCoordinator.SelfId() != senderShardID {
		return maxGasLimitPerBlock, nil
	}

	accountHandler, err := tce.accounts.LoadAccount(tx.SndAddr)
	if err != nil {
		return 0, err
	}

	if check.IfNil(accountHandler) {
		return 0, process.ErrInsufficientFee
	}

	accountSender, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return 0, process.ErrInsufficientFee
	}

	accountSenderBalance := accountSender.GetBalance()
	tx.GasLimit = maxGasLimitPerBlock
	txFee := tce.feeHandler.ComputeTxFee(tx)
	if txFee.Cmp(accountSenderBalance) > 0 && big.NewInt(0).Cmp(accountSenderBalance) != 0 {
		return tce.feeHandler.ComputeGasLimitBasedOnBalance(tx, accountSenderBalance)
	}

	return maxGasLimitPerBlock, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
