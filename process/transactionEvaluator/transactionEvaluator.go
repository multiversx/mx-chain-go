package transactionEvaluator

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const dummySignature = "01010101"
const gasRemainedSplitString = "gas remained = "
const gasUsedSlitString = "gas used = "

// ArgsApiTransactionEvaluator holds the arguments required for creating a new transaction evaluator
type ArgsApiTransactionEvaluator struct {
	TxTypeHandler       process.TxTypeHandler
	FeeHandler          process.FeeHandler
	TxSimulator         facade.TransactionSimulatorProcessor
	Accounts            state.AccountsAdapterWithClean
	ShardCoordinator    sharding.Coordinator
	EnableEpochsHandler common.EnableEpochsHandler
}

type apiTransactionEvaluator struct {
	accounts            state.AccountsAdapterWithClean
	shardCoordinator    sharding.Coordinator
	txTypeHandler       process.TxTypeHandler
	feeHandler          process.FeeHandler
	txSimulator         facade.TransactionSimulatorProcessor
	enableEpochsHandler common.EnableEpochsHandler
	mutExecution        sync.RWMutex
}

// NewAPITransactionEvaluator will create a new api transaction evaluator
func NewAPITransactionEvaluator(args ArgsApiTransactionEvaluator) (*apiTransactionEvaluator, error) {
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
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.CleanUpInformativeSCRsFlag,
	})
	if err != nil {
		return nil, err
	}

	tce := &apiTransactionEvaluator{
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
func (ate *apiTransactionEvaluator) SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
	ate.mutExecution.Lock()
	defer func() {
		ate.accounts.CleanCache()
		ate.mutExecution.Unlock()
	}()

	return ate.txSimulator.ProcessTx(tx)
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (ate *apiTransactionEvaluator) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	ate.mutExecution.Lock()
	defer func() {
		ate.accounts.CleanCache()
		ate.mutExecution.Unlock()
	}()

	txTypeOnSender, txTypeOnDestination := ate.txTypeHandler.ComputeTransactionType(tx)
	if txTypeOnSender == process.MoveBalance && txTypeOnDestination == process.MoveBalance {
		return ate.computeMoveBalanceCost(tx), nil
	}

	switch txTypeOnSender {
	case process.SCDeployment, process.SCInvoking, process.BuiltInFunctionCall, process.MoveBalance:
		return ate.simulateTransactionCost(tx, txTypeOnSender)
	case process.RelayedTx, process.RelayedTxV2, process.RelayedTxV3:
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

func (ate *apiTransactionEvaluator) computeMoveBalanceCost(tx *transaction.Transaction) *transaction.CostResponse {
	gasUnits := ate.feeHandler.ComputeGasLimit(tx)

	return &transaction.CostResponse{
		GasUnits:      gasUnits,
		ReturnMessage: "",
	}
}

func (ate *apiTransactionEvaluator) simulateTransactionCost(tx *transaction.Transaction, txType process.TransactionType) (*transaction.CostResponse, error) {
	err := ate.addMissingFieldsIfNeeded(tx)
	if err != nil {
		return nil, err
	}

	costResponse := &transaction.CostResponse{}

	res, err := ate.txSimulator.ProcessTx(tx)
	if err != nil {
		costResponse.ReturnMessage = err.Error()
		return costResponse, nil
	}

	isMoveBalanceOk := txType == process.MoveBalance && res.FailReason == ""
	if isMoveBalanceOk {
		costResponse.GasUnits = ate.feeHandler.ComputeGasLimit(tx)
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
		costResponse.GasUnits = ate.computeGasUnitsBasedOnVMOutput(tx, res.VMOutput)
		return costResponse, nil
	}

	costResponse.ReturnMessage = fmt.Sprintf("%s: %s", res.VMOutput.ReturnCode.String(), returnMessageFromVMOutput)
	return costResponse, nil
}

func (ate *apiTransactionEvaluator) computeGasUnitsBasedOnVMOutput(tx *transaction.Transaction, vmOutput *vmcommon.VMOutput) uint64 {
	isTooMuchGasProvided := strings.Contains(vmOutput.ReturnMessage, smartContract.TooMuchGasProvidedMessage)
	if !isTooMuchGasProvided {
		return tx.GasLimit - vmOutput.GasRemaining
	}

	isTooMuchGasV2MsgFlagSet := ate.enableEpochsHandler.IsFlagEnabled(common.CleanUpInformativeSCRsFlag)
	if isTooMuchGasV2MsgFlagSet {
		gasNeededForProcessing := extractGasRemainedFromMessage(vmOutput.ReturnMessage, gasUsedSlitString)
		return ate.feeHandler.ComputeGasLimit(tx) + gasNeededForProcessing
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

func (ate *apiTransactionEvaluator) addMissingFieldsIfNeeded(tx *transaction.Transaction) error {
	if tx.GasPrice == 0 {
		tx.GasPrice = ate.feeHandler.MinGasPrice()
	}
	if len(tx.Signature) == 0 {
		tx.Signature = []byte(dummySignature)
	}
	if tx.GasLimit == 0 {
		var err error
		tx.GasLimit, err = ate.getTxGasLimit(tx)

		return err
	}

	return nil
}

func (ate *apiTransactionEvaluator) getTxGasLimit(tx *transaction.Transaction) (uint64, error) {
	selfShardID := ate.shardCoordinator.SelfId()
	maxGasLimitPerBlock := ate.feeHandler.MaxGasLimitPerBlock(selfShardID) - 1

	senderShardID := ate.shardCoordinator.ComputeId(tx.SndAddr)
	if ate.shardCoordinator.SelfId() != senderShardID {
		return maxGasLimitPerBlock, nil
	}

	accountHandler, err := ate.accounts.LoadAccount(tx.SndAddr)
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
	txFee := ate.feeHandler.ComputeTxFee(tx)
	if txFee.Cmp(accountSenderBalance) > 0 && big.NewInt(0).Cmp(accountSenderBalance) != 0 {
		return ate.feeHandler.ComputeGasLimitBasedOnBalance(tx, accountSenderBalance)
	}

	return maxGasLimitPerBlock, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ate *apiTransactionEvaluator) IsInterfaceNil() bool {
	return ate == nil
}
