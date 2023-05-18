package txsimulator

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const dummySignature = "01010101"
const gasRemainedSplitString = "gas remained = "
const gasUsedSlitString = "gas used = "

type ArgsTransactionCostSimulator struct {
	TxTypeHandler       process.TxTypeHandler
	FeeHandler          process.FeeHandler
	TxSimulator         facade.TransactionSimulatorProcessor
	Accounts            state.AccountsAdapterWithClean
	ShardCoordinator    sharding.Coordinator
	EnableEpochsHandler common.EnableEpochsHandler
}

type transactionCostEstimator struct {
	accounts            state.AccountsAdapterWithClean
	shardCoordinator    sharding.Coordinator
	txTypeHandler       process.TxTypeHandler
	feeHandler          process.FeeHandler
	txSimulator         facade.TransactionSimulatorProcessor
	enableEpochsHandler common.EnableEpochsHandler
	mutExecution        sync.RWMutex
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(args ArgsTransactionCostSimulator) (*transactionCostEstimator, error) {
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.FeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.TxSimulator) {
		return nil, ErrNilTxSimulatorProcessor
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	tce := &transactionCostEstimator{
		txTypeHandler:       args.TxTypeHandler,
		feeHandler:          args.FeeHandler,
		txSimulator:         args.TxSimulator,
		accounts:            args.Accounts,
		shardCoordinator:    args.ShardCoordinator,
		enableEpochsHandler: args.EnableEpochsHandler,
	}

	return tce, nil
}

// SimulateTransactionExecution will simulate a transaction's execution and will return the results
func (tce *transactionCostEstimator) SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
	tce.mutExecution.Lock()
	defer func() {
		tce.accounts.CleanCache()
		tce.mutExecution.Unlock()
	}()

	return tce.txSimulator.ProcessTx(tx)
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	tce.mutExecution.Lock()
	defer func() {
		tce.accounts.CleanCache()
		tce.mutExecution.Unlock()
	}()

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

	costResponse := &transaction.CostResponse{}

	res, err := tce.txSimulator.ProcessTx(tx)
	if err != nil {
		costResponse.ReturnMessage = err.Error()
		return costResponse, nil
	}

	isMoveBalanceOk := txType == process.MoveBalance && res.FailReason == ""
	if isMoveBalanceOk {
		costResponse.GasUnits = tce.feeHandler.ComputeGasLimit(tx)
		return costResponse, nil

	}

	returnMessageFromVMOutput := ""
	if res.VMOutput != nil {
		returnMessageFromVMOutput = res.VMOutput.ReturnMessage
	}

	if res.FailReason != "" {
		costResponse.ReturnMessage = fmt.Sprintf("%s: %s", res.FailReason, returnMessageFromVMOutput)
		return costResponse, nil
	}

	if res.VMOutput == nil {
		costResponse.ReturnMessage = process.ErrNilVMOutput.Error()
		return costResponse, nil
	}

	costResponse.SmartContractResults = res.ScResults
	costResponse.Logs = res.Logs
	if res.VMOutput.ReturnCode == vmcommon.Ok {
		costResponse.GasUnits = tce.computeGasUnitsBasedOnVMOutput(tx, res.VMOutput)
		return costResponse, nil
	}

	costResponse.ReturnMessage = fmt.Sprintf("%s: %s", res.VMOutput.ReturnCode.String(), returnMessageFromVMOutput)
	return costResponse, nil
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
