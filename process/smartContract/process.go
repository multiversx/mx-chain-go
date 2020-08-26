package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/vm"
	"math/big"
	"strings"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.SmartContractResultProcessor = (*scProcessor)(nil)
var _ process.SmartContractProcessor = (*scProcessor)(nil)

var log = logger.GetOrCreate("process/smartcontract")

const executeDurationAlarmThreshold = time.Duration(100) * time.Millisecond

// TODO: Move to vm-common.
const upgradeFunctionName = "upgradeContract"

var zero = big.NewInt(0)

type scProcessor struct {
	accounts                       state.AccountsAdapter
	blockChainHook                 process.BlockChainHookHandler
	pubkeyConv                     core.PubkeyConverter
	hasher                         hashing.Hasher
	marshalizer                    marshal.Marshalizer
	shardCoordinator               sharding.Coordinator
	vmContainer                    process.VirtualMachinesContainer
	argsParser                     process.ArgumentsParser
	builtInFunctions               process.BuiltInFunctionContainer
	deployEnableEpoch              uint32
	builtinEnableEpoch             uint32
	penalizedTooMuchGasEnableEpoch uint32
	flagDeploy                     atomic.Flag
	flagBuiltin                    atomic.Flag
	flagPenalizedTooMuchGas        atomic.Flag

	badTxForwarder process.IntermediateTransactionHandler
	scrForwarder   process.IntermediateTransactionHandler
	txFeeHandler   process.TransactionFeeHandler
	economicsFee   process.FeeHandler
	txTypeHandler  process.TxTypeHandler
	gasHandler     process.GasHandler

	asyncCallbackGasLock uint64
	asyncCallStepCost    uint64

	txLogsProcessor process.TransactionLogProcessor
}

// ArgsNewSmartContractProcessor defines the arguments needed for new smart contract processor
type ArgsNewSmartContractProcessor struct {
	VmContainer                    process.VirtualMachinesContainer
	ArgsParser                     process.ArgumentsParser
	Hasher                         hashing.Hasher
	Marshalizer                    marshal.Marshalizer
	AccountsDB                     state.AccountsAdapter
	BlockChainHook                 process.BlockChainHookHandler
	PubkeyConv                     core.PubkeyConverter
	Coordinator                    sharding.Coordinator
	ScrForwarder                   process.IntermediateTransactionHandler
	TxFeeHandler                   process.TransactionFeeHandler
	EconomicsFee                   process.FeeHandler
	TxTypeHandler                  process.TxTypeHandler
	GasHandler                     process.GasHandler
	GasSchedule                    map[string]map[string]uint64
	BuiltInFunctions               process.BuiltInFunctionContainer
	TxLogsProcessor                process.TransactionLogProcessor
	BadTxForwarder                 process.IntermediateTransactionHandler
	DeployEnableEpoch              uint32
	BuiltinEnableEpoch             uint32
	PenalizedTooMuchGasEnableEpoch uint32
	EpochNotifier                  process.EpochNotifier
}

// NewSmartContractProcessor creates a smart contract processor that creates and interprets VM data
func NewSmartContractProcessor(args ArgsNewSmartContractProcessor) (*scProcessor, error) {
	if check.IfNil(args.VmContainer) {
		return nil, process.ErrNoVM
	}
	if check.IfNil(args.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.AccountsDB) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.BlockChainHook) {
		return nil, process.ErrNilTemporaryAccountsHandler
	}
	if check.IfNil(args.PubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ScrForwarder) {
		return nil, process.ErrNilIntermediateTransactionHandler
	}
	if check.IfNil(args.TxFeeHandler) {
		return nil, process.ErrNilUnsignedTxHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.GasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if args.GasSchedule == nil {
		return nil, process.ErrNilGasSchedule
	}
	if check.IfNil(args.BuiltInFunctions) {
		return nil, process.ErrNilBuiltInFunction
	}
	if check.IfNil(args.TxLogsProcessor) {
		return nil, process.ErrNilTxLogsProcessor
	}
	if check.IfNil(args.BadTxForwarder) {
		return nil, process.ErrNilBadTxHandler
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	apiCosts := args.GasSchedule[core.ElrondAPICost]

	sc := &scProcessor{
		vmContainer:                    args.VmContainer,
		argsParser:                     args.ArgsParser,
		hasher:                         args.Hasher,
		marshalizer:                    args.Marshalizer,
		accounts:                       args.AccountsDB,
		blockChainHook:                 args.BlockChainHook,
		pubkeyConv:                     args.PubkeyConv,
		shardCoordinator:               args.Coordinator,
		scrForwarder:                   args.ScrForwarder,
		txFeeHandler:                   args.TxFeeHandler,
		economicsFee:                   args.EconomicsFee,
		txTypeHandler:                  args.TxTypeHandler,
		gasHandler:                     args.GasHandler,
		asyncCallStepCost:              apiCosts[core.AsyncCallStepField],
		asyncCallbackGasLock:           apiCosts[core.AsyncCallbackGasLockField],
		builtInFunctions:               args.BuiltInFunctions,
		txLogsProcessor:                args.TxLogsProcessor,
		badTxForwarder:                 args.BadTxForwarder,
		deployEnableEpoch:              args.DeployEnableEpoch,
		builtinEnableEpoch:             args.BuiltinEnableEpoch,
		penalizedTooMuchGasEnableEpoch: args.PenalizedTooMuchGasEnableEpoch,
	}

	args.EpochNotifier.RegisterNotifyHandler(sc)

	return sc, nil
}

func (sc *scProcessor) checkTxValidity(tx data.TransactionHandler) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := sc.pubkeyConv.Len() != len(tx.GetRcvAddr())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

func (sc *scProcessor) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRcvAddr(), make([]byte, sc.pubkeyConv.Len()))
	return isEmptyAddress
}

// ExecuteSmartContractTransaction processes the transaction, call the VM and processes the SC call output
func (sc *scProcessor) ExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	if check.IfNil(tx) {
		return 0, process.ErrNilTransaction
	}

	sw := core.NewStopWatch()
	sw.Start("execute")
	returnCode, err := sc.doExecuteSmartContractTransaction(tx, acntSnd, acntDst)
	sw.Stop("execute")
	duration := sw.GetMeasurement("execute")

	if duration > executeDurationAlarmThreshold {
		log.Debug(fmt.Sprintf("scProcessor.ExecuteSmartContractTransaction(): execution took > %s", duration), "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
	} else {
		log.Trace("scProcessor.ExecuteSmartContractTransaction()", "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
	}

	return returnCode, err
}

func (sc *scProcessor) doExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		log.Debug("process sc payment error", "error", err.Error())
		return 0, err
	}

	var txHash []byte
	txHash, err = core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return 0, err
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		log.Debug("saveAccounts error", "error", err)
		return 0, err
	}

	returnMessage := ""
	var vmOutput *vmcommon.VMOutput
	var executedBuiltIn bool
	snapshot := sc.accounts.JournalLen()

	var vmInput *vmcommon.ContractCallInput
	vmInput, err = sc.createVMCallInput(tx)
	if err != nil {
		returnMessage = "cannot create VMInput, check the transaction data field"
		log.Debug("create vm call input error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot)
	}

	err = sc.checkUpgradePermission(acntDst, vmInput)
	if err != nil {
		log.Debug("checkUpgradePermission", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(err.Error()), snapshot)
	}

	executedBuiltIn, err = sc.resolveBuiltInFunctions(txHash, tx, acntSnd, acntDst, vmInput)
	if err != nil {
		log.Trace("processed built in functions error", "error", err.Error())
		if executedBuiltIn {
			return vmcommon.UserError, sc.resolveFailedTransaction(acntSnd, tx, txHash, err.Error(), snapshot)
		}
		return vmcommon.UserError, err
	}
	if executedBuiltIn {
		return vmcommon.Ok, nil
	}

	if check.IfNil(acntDst) {
		return 0, process.ErrNilSCDestAccount
	}

	var vmExec vmcommon.VMExecutionHandler
	vmExec, err = findVMByTransaction(sc.vmContainer, tx)
	if err != nil {
		returnMessage = "cannot get vm from address"
		log.Debug("get vm from address error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot)
	}

	vmOutput, err = vmExec.RunSmartContractCall(vmInput)
	if err != nil {
		log.Debug("run smart contract call error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot)
	}

	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Debug("run smart contract call error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return 0, err
	}

	var consumedFee *big.Int
	var results []data.TransactionHandler
	results, consumedFee, err = sc.processVMOutput(vmOutput, txHash, tx, acntSnd, vmInput.CallType, vmInput.GasProvided)
	if err != nil {
		log.Trace("process vm output returned with problem ", "err", err.Error())
		return vmOutput.ReturnCode, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(vmOutput.ReturnMessage), snapshot)
	}

	finalResults := sc.deleteSCRsWithValueZeroGoingToMeta(results)
	err = sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Debug("AddIntermediateTransactions error", "error", err.Error())
		return 0, err
	}

	ignorableError := sc.txLogsProcessor.SaveLog(txHash, tx, vmOutput.Logs)
	if ignorableError != nil {
		log.Debug("txLogsProcessor.SaveLog() error", "error", ignorableError.Error())
	}

	newDeveloperReward := core.GetPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())

	acntDst, err = sc.reloadLocalAccount(acntDst)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return 0, err
	}

	acntDst.AddToDeveloperReward(newDeveloperReward)
	sc.txFeeHandler.ProcessTransactionFee(consumedFee, newDeveloperReward, txHash)

	err = sc.accounts.SaveAccount(acntDst)
	if err != nil {
		log.Debug("error saving account")
		return 0, err
	}

	return vmcommon.Ok, nil
}

func (sc *scProcessor) deleteSCRsWithValueZeroGoingToMeta(scrs []data.TransactionHandler) []data.TransactionHandler {
	if sc.shardCoordinator.SelfId() == core.MetachainShardId || len(scrs) == 0 {
		return scrs
	}

	cleanSCRs := make([]data.TransactionHandler, 0, len(scrs))
	for _, scr := range scrs {
		shardID := sc.shardCoordinator.ComputeId(scr.GetRcvAddr())
		if shardID == core.MetachainShardId && scr.GetGasLimit() == 0 && scr.GetValue().Cmp(zero) == 0 {
			continue
		}
		cleanSCRs = append(cleanSCRs, scr)
	}

	return cleanSCRs
}

func (sc *scProcessor) saveAccounts(acntSnd, acntDst state.AccountHandler) error {
	if !check.IfNil(acntSnd) {
		err := sc.accounts.SaveAccount(acntSnd)
		if err != nil {
			return err
		}
	}

	if !check.IfNil(acntDst) {
		err := sc.accounts.SaveAccount(acntDst)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *scProcessor) resolveFailedTransaction(
	acntSnd state.UserAccountHandler,
	tx data.TransactionHandler,
	txHash []byte,
	errorMessage string,
	snapshot int,
) error {

	err := sc.ProcessIfError(acntSnd, txHash, tx, errorMessage, []byte(errorMessage), snapshot)
	if err != nil {
		return err
	}

	if _, ok := tx.(*transaction.Transaction); ok {
		err = sc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx})
		if err != nil {
			return err
		}
	}

	return process.ErrFailedTransaction
}

func (sc *scProcessor) resolveBuiltInFunctions(
	txHash []byte,
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (bool, error) {

	builtIn, err := sc.builtInFunctions.Get(vmInput.Function)
	if err != nil {
		return false, nil
	}

	if !check.IfNil(acntSnd) && !sc.flagBuiltin.IsSet() {
		return true, process.ErrBuiltInFunctionsAreDisabled
	}

	lockedGas, err := sc.handleAsyncStepGas(vmInput)
	if err != nil {
		return true, err
	}

	// TODO: returned error should be protocol error - vmOutput error must be used for user errors
	// return error here only if acntSnd is not nil - so this is sender shard
	vmOutput, err := builtIn.ProcessBuiltinFunction(acntSnd, acntDst, vmInput)
	if err != nil {
		if !check.IfNil(acntSnd) {
			log.Trace("built in function error at sender", "err", err, "function", vmInput.Function)
			return true, err
		}

		vmOutput = &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: err.Error(),
			GasRemaining:  lockedGas,
		}
	}

	sc.penalizeUserIfNeeded(tx, txHash, vmInput.CallType, vmInput.GasProvided, vmOutput)

	// Restore gas set aside for the async callback, if applicable
	vmOutput.GasRemaining += lockedGas

	scrResults := make([]data.TransactionHandler, 0, len(vmOutput.OutputAccounts)+1)

	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		scTx := sc.createSmartContractResult(outAcc, tx, txHash, storageUpdates)
		scrResults = append(scrResults, scTx)
	}

	scrForSender, consumedFee := sc.createSCRForSender(
		vmOutput,
		tx,
		txHash,
		acntSnd,
		vmInput.CallType,
	)
	if !check.IfNil(acntSnd) {
		err = acntSnd.AddToBalance(scrForSender.Value)
		if err != nil {
			return false, err
		}
	}

	scrResults = append(scrResults, scrForSender)
	finalResults := sc.deleteSCRsWithValueZeroGoingToMeta(scrResults)
	err = sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Debug("AddIntermediateTransactions error", "error", err.Error())
		return false, err
	}

	if check.IfNil(acntSnd) {
		// it was already consumed in sender shard
		consumedFee = big.NewInt(0)
	}

	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)
	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ProcessIfError creates a smart contract result, consumed the gas and returns the value to the user
func (sc *scProcessor) ProcessIfError(
	acntSnd state.UserAccountHandler,
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
	snapShot int,
) error {
	err := sc.accounts.RevertToSnapshot(snapShot)
	if err != nil {
		log.Warn("revert to snapshot", "error", err.Error())
		return err
	}

	consumedFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GetGasLimit()), big.NewInt(0).SetUint64(tx.GetGasPrice()))
	scrIfError := sc.createSCRsWhenError(txHash, tx, returnCode, returnMessage)

	if check.IfNil(acntSnd) {
		moveBalanceCost := sc.economicsFee.ComputeMoveBalanceFee(tx)
		consumedFee.Sub(consumedFee, moveBalanceCost)
	}

	err = sc.addBackTxValues(acntSnd, scrIfError, tx)
	if err != nil {
		return err
	}

	err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrIfError})
	if err != nil {
		return err
	}

	err = sc.processForRelayerWhenError(tx, returnMessage)
	if err != nil {
		return err
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)

	return nil
}

func (sc *scProcessor) processForRelayerWhenError(
	originalTx data.TransactionHandler,
	returnMessage []byte,
) error {
	relayedSCR, isRelayedTx := isRelayedTx(originalTx)
	if !isRelayedTx {
		return nil
	}

	relayerAcnt, err := sc.getAccountFromAddress(relayedSCR.RelayerAddr)
	if err != nil {
		return err
	}

	if !check.IfNil(relayerAcnt) {
		err := relayerAcnt.AddToBalance(relayedSCR.RelayedValue)
		if err != nil {
			return err
		}

		err = sc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			log.Debug("error saving account")
			return err
		}
	}

	scrForRelayer := &smartContractResult.SmartContractResult{
		Nonce:          originalTx.GetNonce(),
		Value:          relayedSCR.RelayedValue,
		RcvAddr:        relayedSCR.RelayerAddr,
		SndAddr:        relayedSCR.RcvAddr,
		OriginalTxHash: relayedSCR.OriginalTxHash,
		ReturnMessage:  returnMessage,
	}

	err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrForRelayer})
	if err != nil {
		return err
	}

	return nil
}

// transaction must be of type SCR and relayed address to be set with reward value higher than 0
func isRelayedTx(tx data.TransactionHandler) (*smartContractResult.SmartContractResult, bool) {
	relayedSCR, ok := tx.(*smartContractResult.SmartContractResult)
	if !ok {
		return nil, false
	}

	if len(relayedSCR.RelayerAddr) == len(relayedSCR.SndAddr) &&
		relayedSCR.RelayedValue != nil && relayedSCR.RelayedValue.Cmp(zero) > 0 {
		return relayedSCR, true
	}

	return nil, false
}

// refunds the transaction values minus the relayed value to the sender account
// in case of failed smart contract execution - gas is consumed, value is sent back
func (sc *scProcessor) addBackTxValues(
	acntSnd state.UserAccountHandler,
	scrIfError *smartContractResult.SmartContractResult,
	originalTx data.TransactionHandler,
) error {
	valueForSnd := big.NewInt(0).Set(scrIfError.Value)

	relayedSCR, isRelayedTx := isRelayedTx(originalTx)
	if isRelayedTx {
		valueForSnd.Sub(valueForSnd, relayedSCR.RelayedValue)
		if valueForSnd.Cmp(zero) < 0 {
			return process.ErrNegativeValue
		}
		scrIfError.Value = big.NewInt(0).Set(valueForSnd)
	}

	if !check.IfNil(acntSnd) {
		err := acntSnd.AddToBalance(valueForSnd)
		if err != nil {
			return err
		}

		err = sc.accounts.SaveAccount(acntSnd)
		if err != nil {
			log.Debug("error saving account")
			return err
		}
	}

	return nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(tx data.TransactionHandler, acntSnd state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	err := sc.checkTxValidity(tx)
	if err != nil {
		log.Debug("invalid transaction", "error", err.Error())
		return 0, err
	}

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		log.Debug("wrong transaction", "error", process.ErrWrongTransaction.Error())
		return 0, process.ErrWrongTransaction
	}

	txHash, err := core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return 0, err
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return 0, err
	}

	err = sc.saveAccounts(acntSnd, nil)
	if err != nil {
		log.Debug("saveAccounts error", "error", err)
		return 0, err
	}

	var vmOutput *vmcommon.VMOutput
	snapshot := sc.accounts.JournalLen()

	if !sc.flagDeploy.IsSet() {
		log.Trace("deploy is disabled")
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, process.ErrSmartContractDeploymentIsDisabled.Error(), []byte(""), snapshot)
	}

	vmInput, vmType, err := sc.createVMDeployInput(tx)
	if err != nil {
		log.Debug("Transaction error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot)
	}

	vmExec, err := sc.vmContainer.Get(vmType)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot)
	}

	vmOutput, err = vmExec.RunSmartContractCreate(vmInput)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return 0, err
	}

	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Debug("run smart contract call error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot)
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, txHash, tx, acntSnd, vmInput.CallType, vmInput.GasProvided)
	if err != nil {
		log.Trace("Processing error", "error", err.Error())
		return vmOutput.ReturnCode, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(vmOutput.ReturnMessage), snapshot)
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("AddIntermediate Transaction error", "error", err.Error())
		return 0, err
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)
	sc.printScDeployed(vmOutput, tx)

	return 0, nil
}

func (sc *scProcessor) printScDeployed(vmOutput *vmcommon.VMOutput, tx data.TransactionHandler) {
	scGenerated := make([]string, 0, len(vmOutput.OutputAccounts))
	for _, account := range vmOutput.OutputAccounts {
		if account == nil {
			continue
		}

		addr := account.Address
		if !core.IsSmartContractAddress(addr) {
			continue
		}

		scGenerated = append(scGenerated, sc.pubkeyConv.Encode(addr))
	}

	log.Debug("SmartContract deployed",
		"owner", sc.pubkeyConv.Encode(tx.GetSndAddr()),
		"SC address(es)", strings.Join(scGenerated, ", "))
}

// taking money from sender, as VM might not have access to him because of state sharding
func (sc *scProcessor) processSCPayment(tx data.TransactionHandler, acntSnd state.UserAccountHandler) error {
	if check.IfNil(acntSnd) {
		// transaction was already processed at sender shard
		return nil
	}

	acntSnd.IncreaseNonce(1)
	err := sc.economicsFee.CheckValidityTxValues(tx)
	if err != nil {
		return err
	}

	cost := big.NewInt(0)
	cost = cost.Mul(big.NewInt(0).SetUint64(tx.GetGasPrice()), big.NewInt(0).SetUint64(tx.GetGasLimit()))
	cost = cost.Add(cost, tx.GetValue())

	if cost.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	err = acntSnd.SubFromBalance(cost)
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) processVMOutput(
	vmOutput *vmcommon.VMOutput,
	txHash []byte,
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
	callType vmcommon.CallType,
	gasProvided uint64,
) ([]data.TransactionHandler, *big.Int, error) {
	if vmOutput == nil {
		return nil, nil, process.ErrNilVMOutput
	}
	if check.IfNil(tx) {
		return nil, nil, process.ErrNilTransaction
	}

	sc.penalizeUserIfNeeded(tx, txHash, callType, gasProvided, vmOutput)

	scrForSender, consumedFee := sc.createSCRForSender(
		vmOutput,
		tx,
		txHash,
		acntSnd,
		callType,
	)

	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Trace("smart contract processing returned with error",
			"hash", txHash,
			"return code", vmOutput.ReturnCode.String(),
			"return message", vmOutput.ReturnMessage,
		)

		if callType == vmcommon.AsynchronousCall || callType == vmcommon.AsynchronousCallBack {
			return []data.TransactionHandler{scrForSender}, consumedFee, nil
		}

		return nil, nil, fmt.Errorf(vmOutput.ReturnCode.String())
	}

	outPutAccounts := process.SortVMOutputInsideData(vmOutput)

	scrTxs, err := sc.processSCOutputAccounts(outPutAccounts, tx, txHash)
	if err != nil {
		return nil, nil, err
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return nil, nil, err
	}

	totalGasConsumed := tx.GetGasLimit() - vmOutput.GasRemaining
	log.Trace("total gas consumed", "value", totalGasConsumed, "hash", txHash)

	if vmOutput.GasRefund.Cmp(big.NewInt(0)) > 0 {
		log.Trace("total gas refunded", "value", vmOutput.GasRefund.String(), "hash", txHash)
	}

	scrTxs = append(scrTxs, scrForSender)

	if !check.IfNil(acntSnd) {
		err = acntSnd.AddToBalance(scrForSender.Value)
		if err != nil {
			return nil, nil, err
		}

		err = sc.accounts.SaveAccount(acntSnd)
		if err != nil {
			return nil, nil, err
		}
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, nil, err
	}

	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)

	return scrTxs, consumedFee, nil
}

func (sc *scProcessor) penalizeUserIfNeeded(
	tx data.TransactionHandler,
	txHash []byte,
	callType vmcommon.CallType,
	gasProvided uint64,
	vmOutput *vmcommon.VMOutput,
) {
	if !sc.flagPenalizedTooMuchGas.IsSet() {
		return
	}

	isTooMuchGasProvided := callType != vmcommon.AsynchronousCall && isTooMuchGasProvided(gasProvided, vmOutput.GasRemaining)
	if !isTooMuchGasProvided {
		return
	}

	gasUsed := gasProvided - vmOutput.GasRemaining
	log.Trace("scProcessor.penalizeUserIfNeeded: too much gas provided",
		"hash", txHash,
		"nonce", tx.GetNonce(),
		"value", tx.GetValue(),
		"sender", tx.GetSndAddr(),
		"receiver", tx.GetRcvAddr(),
		"gas limit", tx.GetGasLimit(),
		"gas price", tx.GetGasPrice(),
		"gas provided", gasProvided,
		"gas remained", vmOutput.GasRemaining,
		"gas used", gasUsed,
		"return code", vmOutput.ReturnCode.String(),
		"return message", vmOutput.ReturnMessage,
	)

	vmOutput.ReturnMessage += fmt.Sprintf("too much gas provided: gas needed = %d, gas remained = %d",
		gasUsed, vmOutput.GasRemaining)
	vmOutput.GasRemaining = 0
}

func isTooMuchGasProvided(gasProvided uint64, gasRemained uint64) bool {
	if gasProvided <= gasRemained {
		return false
	}

	gasUsed := gasProvided - gasRemained
	return gasProvided > gasUsed*process.MaxGasFeeHigherFactorAccepted
}

func (sc *scProcessor) createSCRsWhenError(
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
) *smartContractResult.SmartContractResult {
	rcvAddress := tx.GetSndAddr()

	callType := determineCallType(tx)
	if callType == vmcommon.AsynchronousCallBack {
		rcvAddress = tx.GetRcvAddr()
	}

	scr := &smartContractResult.SmartContractResult{
		Nonce:         tx.GetNonce(),
		Value:         tx.GetValue(),
		RcvAddr:       rcvAddress,
		SndAddr:       tx.GetRcvAddr(),
		Data:          []byte("@" + hex.EncodeToString([]byte(returnCode)) + "@" + hex.EncodeToString(txHash)),
		PrevTxHash:    txHash,
		ReturnMessage: returnMessage,
	}

	setOriginalTxHash(scr, txHash, tx)
	if scr.Value == nil {
		scr.Value = big.NewInt(0)
	}
	if scr.Value.Cmp(zero) > 0 {
		scr.OriginalSender = tx.GetSndAddr()
	}

	return scr
}

func setOriginalTxHash(
	scr *smartContractResult.SmartContractResult,
	txHash []byte,
	tx data.TransactionHandler,
) {
	currSCR, isSCR := tx.(*smartContractResult.SmartContractResult)
	if isSCR {
		scr.OriginalTxHash = currSCR.OriginalTxHash
	} else {
		scr.OriginalTxHash = txHash
	}
}

// reloadLocalAccount will reload from current account state the sender account
// this requirement is needed because in the case of refunding the exact account that was previously
// modified in saveSCOutputToCurrentState, the modifications done there should be visible here
func (sc *scProcessor) reloadLocalAccount(acntSnd state.UserAccountHandler) (state.UserAccountHandler, error) {
	if check.IfNil(acntSnd) {
		return acntSnd, nil
	}

	isAccountFromCurrentShard := acntSnd.AddressBytes() != nil
	if !isAccountFromCurrentShard {
		return acntSnd, nil
	}

	return sc.getAccountFromAddress(acntSnd.AddressBytes())
}

func (sc *scProcessor) createSmartContractResult(
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
	storageUpdates []*vmcommon.StorageUpdate,
) *smartContractResult.SmartContractResult {
	result := &smartContractResult.SmartContractResult{}

	result.Value = outAcc.BalanceDelta
	if result.Value == nil {
		result.Value = big.NewInt(0)
	}

	result.Nonce = outAcc.Nonce
	result.RcvAddr = outAcc.Address
	result.SndAddr = tx.GetRcvAddr()
	result.Code = outAcc.Code
	result.Data = outAcc.Data
	if bytes.Equal(result.GetRcvAddr(), vm.StakingSCAddress) {
		//TODO: write directly from staking smart contract to the validator trie - it gets complicated in reverts
		// storage update for staking contract is used in stakingToPeer
		result.Data = append(result.Data, sc.argsParser.CreateDataFromStorageUpdate(storageUpdates)...)
	}
	result.GasLimit = outAcc.GasLimit
	result.GasPrice = tx.GetGasPrice()
	result.PrevTxHash = txHash
	result.CallType = outAcc.CallType
	setOriginalTxHash(result, txHash, tx)

	log.Trace("createSCR ", "data", string(result.Data))
	if result.Value.Cmp(zero) > 0 {
		result.OriginalSender = tx.GetSndAddr()
	}

	return result
}

func (sc *scProcessor) computeActualConsumedGas(
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	acntSnd state.UserAccountHandler,
) uint64 {
	consumedGas := tx.GetGasLimit() - vmOutput.GasRemaining
	if check.IfNil(acntSnd) {
		consumedGas -= sc.economicsFee.ComputeGasLimit(tx)
	}
	for _, outAcc := range vmOutput.OutputAccounts {
		consumedGas -= outAcc.GasLimit
	}

	return consumedGas
}

// createSCRForSender(vmOutput, tx, txHash, acntSnd)
// give back the user the unused gas money
func (sc *scProcessor) createSCRForSender(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte,
	acntSnd state.UserAccountHandler,
	callType vmcommon.CallType,
) (*smartContractResult.SmartContractResult, *big.Int) {
	if vmOutput.GasRefund == nil {
		vmOutput.GasRefund = big.NewInt(0)
	}

	storageFreeRefund := big.NewInt(0).Mul(vmOutput.GasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
	refundErd := core.SafeMul(vmOutput.GasRemaining, tx.GetGasPrice())

	rcvAddress := tx.GetSndAddr()
	if callType == vmcommon.AsynchronousCallBack {
		rcvAddress = tx.GetRcvAddr()
	}

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = big.NewInt(0).Set(storageFreeRefund)
	if callType != vmcommon.AsynchronousCall {
		scTx.Value.Add(scTx.Value, refundErd)
	}

	scTx.RcvAddr = rcvAddress
	scTx.SndAddr = tx.GetRcvAddr()
	scTx.Nonce = tx.GetNonce() + 1
	scTx.PrevTxHash = txHash
	scTx.GasLimit = vmOutput.GasRemaining
	scTx.GasPrice = tx.GetGasPrice()
	scTx.ReturnMessage = []byte(vmOutput.ReturnMessage)
	setOriginalTxHash(scTx, txHash, tx)

	if callType == vmcommon.AsynchronousCall {
		scTx.CallType = vmcommon.AsynchronousCallBack
		scTx.Data = []byte("@" + core.ConvertToEvenHex(int(vmOutput.ReturnCode)))
	} else {
		scTx.Data = []byte("@" + hex.EncodeToString([]byte(vmOutput.ReturnCode.String())))
	}

	for _, retData := range vmOutput.ReturnData {
		scTx.Data = append(scTx.Data, []byte("@"+hex.EncodeToString(retData))...)
	}

	log.Trace("createSCRForSender ", "data", string(scTx.Data))
	consumedFee := core.SafeMul(sc.computeActualConsumedGas(tx, vmOutput, acntSnd), tx.GetGasPrice())
	return scTx, consumedFee
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(
	outputAccounts []*vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	scResults := make([]data.TransactionHandler, 0, len(outputAccounts))

	sumOfAllDiff := big.NewInt(0)
	sumOfAllDiff.Sub(sumOfAllDiff, tx.GetValue())

	zero := big.NewInt(0)
	for _, outAcc := range outputAccounts {
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return nil, err
		}

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		scTx := sc.createSmartContractResult(outAcc, tx, txHash, storageUpdates)
		scResults = append(scResults, scTx)

		if check.IfNil(acc) {
			if outAcc.BalanceDelta != nil {
				sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
			}
			continue
		}

		for j := 0; j < len(storageUpdates); j++ {
			storeUpdate := storageUpdates[j]
			if !process.IsAllowedToSaveUnderKey(storeUpdate.Offset) {
				log.Trace("storeUpdate is not allowed", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
				continue
			}

			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			log.Trace("storeUpdate", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
		}

		sc.updateSmartContractCode(acc, outAcc, tx)
		// change nonce only if there is a change
		if outAcc.Nonce != acc.GetNonce() && outAcc.Nonce != 0 {
			if outAcc.Nonce < acc.GetNonce() {
				return nil, process.ErrWrongNonceInVMOutput
			}

			nonceDifference := outAcc.Nonce - acc.GetNonce()
			acc.IncreaseNonce(nonceDifference)
		}

		// if no change then continue
		if outAcc.BalanceDelta == nil || outAcc.BalanceDelta.Cmp(zero) == 0 {
			err = sc.accounts.SaveAccount(acc)
			if err != nil {
				return nil, err
			}

			continue
		}

		sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)

		err = acc.AddToBalance(outAcc.BalanceDelta)
		if err != nil {
			return nil, err
		}

		err = sc.accounts.SaveAccount(acc)
		if err != nil {
			return nil, err
		}
	}

	if sumOfAllDiff.Cmp(zero) != 0 {
		return nil, process.ErrOverallBalanceChangeFromSC
	}

	return scResults, nil
}

// updateSmartContractCode upgrades code for "direct" deployments & upgrades and for "indirect" deployments & upgrades
// It receives:
// 	(1) the account as found in the State
//	(2) the account as returned in VM Output
// 	(3) the transaction that, upon execution, produced the VM Output
func (sc *scProcessor) updateSmartContractCode(
	stateAccount state.UserAccountHandler,
	outputAccount *vmcommon.OutputAccount,
	tx data.TransactionHandler,
) {
	if len(outputAccount.Code) == 0 {
		return
	}
	if len(outputAccount.CodeMetadata) == 0 {
		return
	}
	if !core.IsSmartContractAddress(outputAccount.Address) {
		return
	}

	// This check is desirable (not required though) since currently both Arwen and IELE send the code in the output account even for "regular" execution
	sameCode := bytes.Equal(outputAccount.Code, stateAccount.GetCode())
	sameCodeMetadata := bytes.Equal(outputAccount.CodeMetadata, stateAccount.GetCodeMetadata())
	if sameCode && sameCodeMetadata {
		return
	}

	transactionSender := tx.GetSndAddr()
	currentOwner := stateAccount.GetOwnerAddress()
	isSenderOwner := bytes.Equal(currentOwner, transactionSender)

	noExistingCode := len(stateAccount.GetCode()) == 0
	noExistingOwner := len(currentOwner) == 0
	currentCodeMetadata := vmcommon.CodeMetadataFromBytes(stateAccount.GetCodeMetadata())
	newCodeMetadata := vmcommon.CodeMetadataFromBytes(outputAccount.CodeMetadata)
	isUpgradeable := currentCodeMetadata.Upgradeable
	isDeployment := noExistingCode && noExistingOwner
	isUpgrade := !isDeployment && isSenderOwner && isUpgradeable

	if isDeployment {
		stateAccount.SetOwnerAddress(transactionSender)
		stateAccount.SetCodeMetadata(outputAccount.CodeMetadata)
		stateAccount.SetCode(outputAccount.Code)
		log.Info("updateSmartContractCode(): created", "address", sc.pubkeyConv.Encode(outputAccount.Address), "upgradeable", newCodeMetadata.Upgradeable)
		return
	}

	if isUpgrade {
		stateAccount.SetCodeMetadata(outputAccount.CodeMetadata)
		stateAccount.SetCode(outputAccount.Code)
		log.Info("updateSmartContractCode(): upgraded", "address", sc.pubkeyConv.Encode(outputAccount.Address), "upgradeable", newCodeMetadata.Upgradeable)
		return
	}

	// TODO: change to return some error when IELE is updated. Currently IELE sends the code in output account even for normal SC RUN
}

// delete accounts - only suicide by current SC or another SC called by current SC - protected by VM
func (sc *scProcessor) deleteAccounts(deletedAccounts [][]byte) error {
	for _, value := range deletedAccounts {
		acc, err := sc.getAccountFromAddress(value)
		if err != nil {
			return err
		}

		if check.IfNil(acc) {
			//TODO: sharded Smart Contract processing
			continue
		}

		err = sc.accounts.RemoveAccount(acc.AddressBytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *scProcessor) getAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	shardForCurrentNode := sc.shardCoordinator.SelfId()
	shardForSrc := sc.shardCoordinator.ComputeId(address)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := sc.accounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	stAcc, ok := acnt.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return stAcc, nil
}

// ProcessSmartContractResult updates the account state from the smart contract result
func (sc *scProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	if check.IfNil(scr) {
		return 0, process.ErrNilSmartContractResult
	}

	log.Trace("scProcessor.ProcessSmartContractResult()", "sender", scr.GetSndAddr(), "receiver", scr.GetRcvAddr(), "data", string(scr.GetData()))

	var err error
	returnCode := vmcommon.UserError
	txHash, err := core.CalculateHash(sc.marshalizer, sc.hasher, scr)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return returnCode, err
	}

	dstAcc, err := sc.getAccountFromAddress(scr.RcvAddr)
	if err != nil {
		return returnCode, err
	}
	sndAcc, err := sc.getAccountFromAddress(scr.SndAddr)
	if err != nil {
		return returnCode, err
	}

	if check.IfNil(dstAcc) {
		err = process.ErrNilSCDestAccount
		return returnCode, err
	}

	snapshot := sc.accounts.JournalLen()
	process.DisplayProcessTxDetails(
		"ProcessSmartContractResult: receiver account details",
		dstAcc,
		scr,
		sc.pubkeyConv,
	)

	txType := sc.txTypeHandler.ComputeTransactionType(scr)

	switch txType {
	case process.MoveBalance:
		err = sc.processSimpleSCR(scr, dstAcc)
		if err != nil {
			return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot)
		}
		return vmcommon.Ok, nil
	case process.SCDeployment:
		err = process.ErrSCDeployFromSCRIsNotPermitted
		return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot)
	case process.SCInvoking:
		returnCode, err = sc.ExecuteSmartContractTransaction(scr, sndAcc, dstAcc)
		return returnCode, err
	case process.BuiltInFunctionCall:
		returnCode, err = sc.ExecuteSmartContractTransaction(scr, sndAcc, dstAcc)
		return returnCode, err
	}

	err = process.ErrWrongTransaction
	return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot)
}

func (sc *scProcessor) processSimpleSCR(
	scResult *smartContractResult.SmartContractResult,
	dstAcc state.UserAccountHandler,
) error {
	if scResult.Value.Cmp(zero) <= 0 {
		return nil
	}

	isPayable, err := sc.IsPayable(scResult.RcvAddr)
	if err != nil {
		return err
	}
	if !isPayable && !bytes.Equal(scResult.RcvAddr, scResult.OriginalSender) {
		return process.ErrAccountNotPayable
	}

	err = dstAcc.AddToBalance(scResult.Value)
	if err != nil {
		return err
	}

	return sc.accounts.SaveAccount(dstAcc)
}

func (sc *scProcessor) checkUpgradePermission(contract state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) error {
	isUpgradeCalled := vmInput.Function == upgradeFunctionName
	if !isUpgradeCalled {
		return nil
	}
	if check.IfNil(contract) {
		return process.ErrUpgradeNotAllowed
	}

	codeMetadata := vmcommon.CodeMetadataFromBytes(contract.GetCodeMetadata())
	isUpgradeable := codeMetadata.Upgradeable
	callerAddress := vmInput.CallerAddr
	ownerAddress := contract.GetOwnerAddress()
	isCallerOwner := bytes.Equal(callerAddress, ownerAddress)

	if isUpgradeable && isCallerOwner {
		return nil
	}

	return process.ErrUpgradeNotAllowed
}

// IsPayable returns if address is payable, smart contract ca set to false
func (sc *scProcessor) IsPayable(address []byte) (bool, error) {
	return sc.blockChainHook.IsPayable(address)
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (sc *scProcessor) EpochConfirmed(epoch uint32) {
	sc.flagDeploy.Toggle(epoch >= sc.deployEnableEpoch)
	log.Debug("scProcessor: deployment of SC", "enabled", sc.flagDeploy.IsSet())

	sc.flagBuiltin.Toggle(epoch >= sc.builtinEnableEpoch)
	log.Debug("scProcessor: built in functions", "enabled", sc.flagBuiltin.IsSet())

	sc.flagPenalizedTooMuchGas.Toggle(epoch >= sc.penalizedTooMuchGasEnableEpoch)
	log.Debug("scProcessor: penalized too much gas", "enabled", sc.flagPenalizedTooMuchGas.IsSet())
}

func (sc *scProcessor) handleAsyncStepGas(input *vmcommon.ContractCallInput) (uint64, error) {
	if input.CallType != vmcommon.AsynchronousCall {
		return 0, nil
	}

	// gasToLock is the amount of gas to set aside for the callback, to avoid it
	// being used by executing built-in functions; this amount will be restored
	// to the caller, so that there is sufficient gas for the async callback
	gasToLock := sc.asyncCallStepCost + sc.asyncCallbackGasLock

	// gasToDeduct also contains an extra asyncCallStepCost, apart from
	// gasToLock; asyncCallStepCost will be deducted, but not refunded, just as
	// Arwen does when executing an async call
	gasToDeduct := sc.asyncCallStepCost + gasToLock
	if input.GasProvided <= gasToDeduct {
		return 0, process.ErrNotEnoughGas
	}

	input.GasProvided -= gasToDeduct

	return gasToLock, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	return sc == nil
}
