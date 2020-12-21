package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
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
	isGenesisProcessing            bool

	badTxForwarder process.IntermediateTransactionHandler
	scrForwarder   process.IntermediateTransactionHandler
	txFeeHandler   process.TransactionFeeHandler
	economicsFee   process.FeeHandler
	txTypeHandler  process.TxTypeHandler
	gasHandler     process.GasHandler

	asyncCallbackGasLock uint64
	asyncCallStepCost    uint64
	esdtTransferCost     uint64
	mutGasLock           sync.RWMutex

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
	GasSchedule                    core.GasScheduleNotifier
	BuiltInFunctions               process.BuiltInFunctionContainer
	TxLogsProcessor                process.TransactionLogProcessor
	BadTxForwarder                 process.IntermediateTransactionHandler
	DeployEnableEpoch              uint32
	BuiltinEnableEpoch             uint32
	PenalizedTooMuchGasEnableEpoch uint32
	EpochNotifier                  process.EpochNotifier
	IsGenesisProcessing            bool
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
	if check.IfNil(args.GasSchedule) || args.GasSchedule.LatestGasSchedule() == nil {
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

	apiCosts := args.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	builtInFuncCost := args.GasSchedule.LatestGasSchedule()[core.BuiltInCost]
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
		esdtTransferCost:               builtInFuncCost[core.BuiltInFunctionESDTTransfer],
		builtInFunctions:               args.BuiltInFunctions,
		txLogsProcessor:                args.TxLogsProcessor,
		badTxForwarder:                 args.BadTxForwarder,
		deployEnableEpoch:              args.DeployEnableEpoch,
		builtinEnableEpoch:             args.BuiltinEnableEpoch,
		penalizedTooMuchGasEnableEpoch: args.PenalizedTooMuchGasEnableEpoch,
		isGenesisProcessing:            args.IsGenesisProcessing,
	}

	args.EpochNotifier.RegisterNotifyHandler(sc)
	args.GasSchedule.RegisterNotifyHandler(sc)

	return sc, nil
}

// GasScheduleChange sets the new gas schedule where it is needed
func (sc *scProcessor) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	sc.mutGasLock.Lock()
	defer sc.mutGasLock.Unlock()

	apiCosts := gasSchedule[core.ElrondAPICost]
	if apiCosts == nil {
		return
	}

	builtInFuncCost := gasSchedule[core.BuiltInCost]
	if builtInFuncCost == nil {
		return
	}

	sc.asyncCallStepCost = apiCosts[core.AsyncCallStepField]
	sc.asyncCallbackGasLock = apiCosts[core.AsyncCallbackGasLockField]
	sc.esdtTransferCost = builtInFuncCost[core.BuiltInFunctionESDTTransfer]
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

func (sc *scProcessor) prepareExecution(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	builtInFuncCall bool,
) (vmcommon.ReturnCode, *vmcommon.ContractCallInput, []byte, error) {
	err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		log.Debug("process sc payment error", "error", err.Error())
		return 0, nil, nil, err
	}

	var txHash []byte
	txHash, err = core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return 0, nil, nil, err
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		log.Debug("saveAccounts error", "error", err)
		return 0, nil, nil, err
	}

	snapshot := sc.accounts.JournalLen()

	var vmInput *vmcommon.ContractCallInput
	vmInput, err = sc.createVMCallInput(tx, txHash, builtInFuncCall)
	if err != nil {
		returnMessage := "cannot create VMInput, check the transaction data field"
		log.Debug("create vm call input error", "error", err.Error())
		return vmcommon.UserError, vmInput, txHash, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot, 0)
	}

	err = sc.checkUpgradePermission(acntDst, vmInput)
	if err != nil {
		log.Debug("checkUpgradePermission", "error", err.Error())
		return vmcommon.UserError, vmInput, txHash, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(err.Error()), snapshot, vmInput.GasLocked)
	}

	return vmcommon.Ok, vmInput, txHash, nil
}

func (sc *scProcessor) doExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	returnCode, vmInput, txHash, err := sc.prepareExecution(tx, acntSnd, acntDst, false)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}

	snapshot := sc.accounts.JournalLen()

	var vmOutput *vmcommon.VMOutput
	vmOutput, err = sc.executeSmartContractCall(vmInput, tx, txHash, snapshot, acntSnd, acntDst)
	if err != nil {
		return returnCode, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode, nil
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte("gas consumed exceeded"), snapshot, vmInput.GasLocked)
	}

	var results []data.TransactionHandler
	results, err = sc.processVMOutput(vmOutput, txHash, tx, vmInput.CallType, vmInput.GasProvided)
	if err != nil {
		log.Trace("process vm output returned with problem ", "err", err.Error())
		return vmOutput.ReturnCode, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	return sc.finishSCExecution(results, txHash, tx, vmOutput, 0)
}

func (sc *scProcessor) executeSmartContractCall(
	vmInput *vmcommon.ContractCallInput,
	tx data.TransactionHandler,
	txHash []byte,
	snapshot int,
	acntSnd, acntDst state.UserAccountHandler,
) (*vmcommon.VMOutput, error) {
	if check.IfNil(acntDst) {
		return nil, process.ErrNilSCDestAccount
	}

	userErrorVmOutput := &vmcommon.VMOutput{
		ReturnCode: vmcommon.UserError,
	}
	vmExec, err := findVMByTransaction(sc.vmContainer, tx)
	if err != nil {
		returnMessage := "cannot get vm from address"
		log.Debug("get vm from address error", "error", err.Error())
		return userErrorVmOutput, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot, vmInput.GasLocked)
	}

	var vmOutput *vmcommon.VMOutput
	vmOutput, err = vmExec.RunSmartContractCall(vmInput)
	if err != nil {
		log.Debug("run smart contract call error", "error", err.Error())
		return userErrorVmOutput, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}
	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Debug("run smart contract call error", "error", err.Error())
		return userErrorVmOutput, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}
	vmOutput.GasRemaining += vmInput.GasLocked

	if vmOutput.ReturnCode != vmcommon.Ok {
		return userErrorVmOutput, sc.ProcessIfError(acntSnd, txHash, tx, vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return nil, err
	}

	ignorableError := sc.txLogsProcessor.SaveLog(txHash, tx, vmOutput.Logs)
	if ignorableError != nil {
		log.Debug("txLogsProcessor.SaveLog() error", "error", ignorableError.Error())
	}

	return vmOutput, nil
}

func (sc *scProcessor) finishSCExecution(
	results []data.TransactionHandler,
	txHash []byte,
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) (vmcommon.ReturnCode, error) {
	finalResults := sc.deleteSCRsWithValueZeroGoingToMeta(results)
	err := sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Error("AddIntermediateTransactions error", "error", err.Error())
		return 0, err
	}

	err = sc.updateDeveloperRewardsProxy(tx, vmOutput, builtInFuncGasUsed)
	if err != nil {
		log.Error("updateDeveloperRewardsProxy", "error", err.Error())
		return 0, err
	}

	totalConsumedFee, totalDevRwd := sc.computeTotalConsumedFeeAndDevRwd(tx, vmOutput, builtInFuncGasUsed)
	sc.txFeeHandler.ProcessTransactionFee(totalConsumedFee, totalDevRwd, txHash)
	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)

	return vmcommon.Ok, nil
}

func (sc *scProcessor) updateDeveloperRewardsV2(
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) error {
	usedGasByMainSC, err := core.SafeSubUint64(tx.GetGasLimit(), vmOutput.GasRemaining)
	if err != nil {
		return err
	}
	usedGasByMainSC, err = core.SafeSubUint64(usedGasByMainSC, builtInFuncGasUsed)
	if err != nil {
		return err
	}

	for _, outAcc := range vmOutput.OutputAccounts {
		if bytes.Equal(tx.GetRcvAddr(), outAcc.Address) {
			continue
		}

		sentGas := uint64(0)
		for _, outTransfer := range outAcc.OutputTransfers {
			sentGas, err = core.SafeAddUint64(sentGas, outTransfer.GasLimit)
			if err != nil {
				return err
			}
		}

		usedGasByMainSC, err = core.SafeSubUint64(usedGasByMainSC, sentGas)
		if err != nil {
			return err
		}
		usedGasByMainSC, err = core.SafeSubUint64(usedGasByMainSC, outAcc.GasUsed)
		if err != nil {
			return err
		}

		if outAcc.GasUsed > 0 && sc.isSelfShard(outAcc.Address) {
			err = sc.addToDevRewardsV2(outAcc.Address, outAcc.GasUsed, tx)
			if err != nil {
				return err
			}
		}
	}

	moveBalanceGasLimit := sc.economicsFee.ComputeGasLimit(tx)
	if !sc.flagDeploy.IsSet() && !sc.isSelfShard(tx.GetSndAddr()) {
		usedGasByMainSC, err = core.SafeSubUint64(usedGasByMainSC, moveBalanceGasLimit)
		if err != nil {
			return err
		}
	} else if !isSmartContractResult(tx) {
		usedGasByMainSC, err = core.SafeSubUint64(usedGasByMainSC, moveBalanceGasLimit)
		if err != nil {
			return err
		}
	}

	err = sc.addToDevRewardsV2(tx.GetRcvAddr(), usedGasByMainSC, tx)
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) addToDevRewardsV2(address []byte, gasUsed uint64, tx data.TransactionHandler) error {
	if core.IsEmptyAddress(address) || !core.IsSmartContractAddress(address) {
		return nil
	}

	consumedFee := sc.economicsFee.ComputeFeeForProcessing(tx, gasUsed)
	devRwd := core.GetPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())

	userAcc, err := sc.getAccountFromAddress(address)
	if err != nil {
		return err
	}

	if check.IfNil(userAcc) {
		return nil
	}

	userAcc.AddToDeveloperReward(devRwd)
	err = sc.accounts.SaveAccount(userAcc)
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) isSelfShard(address []byte) bool {
	return sc.shardCoordinator.ComputeId(address) == sc.shardCoordinator.SelfId()
}

func (sc *scProcessor) gasConsumedChecks(
	tx data.TransactionHandler,
	gasProvided uint64,
	gasLocked uint64,
	vmOutput *vmcommon.VMOutput,
) error {
	if tx.GetGasLimit() == 0 && sc.shardCoordinator.ComputeId(tx.GetSndAddr()) == core.MetachainShardId {
		// special case for issuing and minting ESDT tokens for normal users
		return nil
	}

	totalGasProvided, err := core.SafeAddUint64(gasProvided, gasLocked)
	if err != nil {
		return err
	}

	totalGasInVMOutput := vmOutput.GasRemaining
	for _, outAcc := range vmOutput.OutputAccounts {
		for _, outTransfer := range outAcc.OutputTransfers {
			transferGas := uint64(0)
			transferGas, err = core.SafeAddUint64(outTransfer.GasLocked, outTransfer.GasLimit)
			if err != nil {
				return err
			}

			totalGasInVMOutput, err = core.SafeAddUint64(totalGasInVMOutput, transferGas)
			if err != nil {
				return err
			}
		}
		totalGasInVMOutput, err = core.SafeAddUint64(totalGasInVMOutput, outAcc.GasUsed)
		if err != nil {
			return err
		}
	}

	if totalGasInVMOutput > totalGasProvided {
		return process.ErrMoreGasConsumedThanProvided
	}

	return nil
}

func (sc *scProcessor) computeTotalConsumedFeeAndDevRwd(
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) (*big.Int, *big.Int) {
	if tx.GetGasLimit() == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	senderInSelfShard := sc.isSelfShard(tx.GetSndAddr())
	consumedGas, err := core.SafeSubUint64(tx.GetGasLimit(), vmOutput.GasRemaining)
	log.LogIfError(err, "computeTotalConsumedFeeAndDevRwd", "vmOutput.GasRemaining")
	if !senderInSelfShard {
		consumedGas, err = core.SafeSubUint64(consumedGas, builtInFuncGasUsed)
		log.LogIfError(err, "computeTotalConsumedFeeAndDevRwd", "builtInFuncGasUsed")
	}

	for _, outAcc := range vmOutput.OutputAccounts {
		if !sc.isSelfShard(outAcc.Address) {
			sentGas := uint64(0)
			for _, outTransfer := range outAcc.OutputTransfers {
				sentGas += outTransfer.GasLimit
			}
			consumedGas, err = core.SafeSubUint64(consumedGas, sentGas)
			log.LogIfError(err, "computeTotalConsumedFeeAndDevRwd", "outTransfer.GasLimit")
		}
	}

	moveBalanceGasLimit := sc.economicsFee.ComputeGasLimit(tx)
	if !isSmartContractResult(tx) {
		consumedGas, err = core.SafeSubUint64(consumedGas, moveBalanceGasLimit)
		log.LogIfError(err, "computeTotalConsumedFeeAndDevRwd", "computeGasLimit")
	}

	consumedGasWithoutBuiltin := consumedGas
	if senderInSelfShard {
		consumedGasWithoutBuiltin, err = core.SafeSubUint64(consumedGasWithoutBuiltin, builtInFuncGasUsed)
		log.LogIfError(err, "computeTotalConsumedFeeAndDevRwd", "consumedWithoutBuiltin")
	}

	totalFee := sc.economicsFee.ComputeFeeForProcessing(tx, consumedGas)
	totalFeeMinusBuiltIn := sc.economicsFee.ComputeFeeForProcessing(tx, consumedGasWithoutBuiltin)
	totalDevRwd := core.GetPercentageOfValue(totalFeeMinusBuiltIn, sc.economicsFee.DeveloperPercentage())

	if !isSmartContractResult(tx) && senderInSelfShard {
		totalFee.Add(totalFee, sc.economicsFee.ComputeMoveBalanceFee(tx))
	}

	if !sc.flagDeploy.IsSet() {
		totalDevRwd = core.GetPercentageOfValue(totalFee, sc.economicsFee.DeveloperPercentage())
	}

	return totalFee, totalDevRwd
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

	err := sc.ProcessIfError(acntSnd, txHash, tx, errorMessage, []byte(errorMessage), snapshot, 0)
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

func (sc *scProcessor) computeBuiltInFuncGasUsed(
	tx data.TransactionHandler,
	gasProvided uint64,
	gasRemaining uint64,
) (uint64, error) {
	_, txTypeOnDst := sc.txTypeHandler.ComputeTransactionType(tx)
	if txTypeOnDst != process.SCInvoking {
		return core.SafeSubUint64(gasProvided, gasRemaining)
	}

	sc.mutGasLock.RLock()
	builtInFuncGasUsed := sc.esdtTransferCost
	sc.mutGasLock.RUnlock()

	return builtInFuncGasUsed, nil
}

// ExecuteBuiltInFunction  processes the transaction, executes the built in function call and subsequent results
func (sc *scProcessor) ExecuteBuiltInFunction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	returnCode, vmInput, txHash, err := sc.prepareExecution(tx, acntSnd, acntDst, true)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}

	snapshot := sc.accounts.JournalLen()
	if !sc.flagBuiltin.IsSet() {
		return vmcommon.UserError, sc.resolveFailedTransaction(acntSnd, tx, txHash, process.ErrBuiltInFunctionsAreDisabled.Error(), snapshot)
	}

	vmOutput, err := sc.resolveBuiltInFunctions(acntSnd, acntDst, vmInput)
	if err != nil {
		log.Debug("processed built in functions error", "error", err.Error())
		return 0, err
	}
	builtInFuncGasUsed, err := sc.computeBuiltInFuncGasUsed(tx, vmInput.GasProvided, vmOutput.GasRemaining)
	log.LogIfError(err, "ExecuteBultInFunction", "computeBuiltInFuncGasUSed")

	vmOutput.GasRemaining += vmInput.GasLocked
	if vmOutput.ReturnCode != vmcommon.Ok {
		if !check.IfNil(acntSnd) {
			return vmcommon.UserError, sc.resolveFailedTransaction(acntSnd, tx, txHash, vmOutput.ReturnMessage, snapshot)
		}
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte("gas consumed exceeded"), snapshot, vmInput.GasLocked)
	}

	createdAsyncCallback := false
	scrResults := make([]data.TransactionHandler, 0, len(vmOutput.OutputAccounts)+1)
	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		tmpCreatedAsyncCallback, scTxs := sc.createSmartContractResults(vmOutput, vmInput.CallType, outAcc, tx, txHash)
		createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback
		if len(scTxs) > 0 {
			scrResults = append(scrResults, scTxs...)
		}
	}

	isSCCall, newVMOutput, newVMInput, err := sc.treatExecutionAfterBuiltInFunc(tx, vmInput, vmOutput, acntSnd, acntDst, snapshot)
	if err != nil {
		log.Debug("treat execution after built in function", "error", err.Error())
		return 0, err
	}
	if newVMOutput.ReturnCode != vmcommon.Ok {
		return newVMOutput.ReturnCode, nil
	}

	if isSCCall {
		err = sc.gasConsumedChecks(tx, newVMInput.GasProvided, newVMInput.GasLocked, newVMOutput)
		if err != nil {
			log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
			return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte("gas consumed exceeded"), snapshot, vmInput.GasLocked)
		}

		outPutAccounts := process.SortVMOutputInsideData(newVMOutput)
		var newSCRTxs []data.TransactionHandler
		tmpCreatedAsyncCallback := false
		tmpCreatedAsyncCallback, newSCRTxs, err = sc.processSCOutputAccounts(newVMOutput, vmInput.CallType, outPutAccounts, tx, txHash)
		if err != nil {
			return 0, err
		}
		createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback

		if len(newSCRTxs) > 0 {
			scrResults = append(scrResults, newSCRTxs...)
		}
	}

	scrForSender, scrForRelayer, err := sc.processSCRForSender(tx, txHash, vmInput, newVMOutput)
	if err != nil {
		return 0, err
	}

	if !createdAsyncCallback && vmInput.CallType == vmcommon.AsynchronousCall {
		asyncCallBackSCR := createAsyncCallBackSCRFromVMOutput(newVMOutput, tx, txHash)
		scrResults = append(scrResults, asyncCallBackSCR)
	} else if !createdAsyncCallback {
		scrResults = append(scrResults, scrForSender)
	}

	if !check.IfNil(scrForRelayer) {
		scrResults = append(scrResults, scrForRelayer)
	}

	return sc.finishSCExecution(scrResults, txHash, tx, newVMOutput, builtInFuncGasUsed)
}

func (sc *scProcessor) processSCRForSender(
	tx data.TransactionHandler,
	txHash []byte,
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
) (*smartContractResult.SmartContractResult, *smartContractResult.SmartContractResult, error) {
	sc.penalizeUserIfNeeded(tx, txHash, vmInput.CallType, vmInput.GasProvided, vmOutput)
	scrForSender, scrForRelayer := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmInput.CallType,
	)

	err := sc.addToBalanceIfInShard(scrForSender.RcvAddr, scrForSender.Value)
	if err != nil {
		return nil, nil, err
	}

	if !check.IfNil(scrForRelayer) {
		err = sc.addToBalanceIfInShard(scrForRelayer.RcvAddr, scrForRelayer.Value)
		if err != nil {
			return nil, nil, err
		}
	}
	return scrForSender, scrForRelayer, nil
}

func (sc *scProcessor) resolveBuiltInFunctions(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {

	vmOutput := &vmcommon.VMOutput{
		ReturnCode: vmcommon.UserError,
	}

	builtIn, err := sc.builtInFunctions.Get(vmInput.Function)
	if err != nil {
		vmOutput.ReturnMessage = err.Error()
		return vmOutput, nil
	}

	vmOutput, err = builtIn.ProcessBuiltinFunction(acntSnd, acntDst, vmInput)
	if err != nil {
		vmOutput = &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: err.Error(),
			GasRemaining:  0,
		}
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		return nil, err
	}

	return vmOutput, nil
}

func (sc *scProcessor) treatExecutionAfterBuiltInFunc(
	tx data.TransactionHandler,
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
	acntSnd state.UserAccountHandler,
	acntDst state.UserAccountHandler,
	snapshot int,
) (bool, *vmcommon.VMOutput, *vmcommon.ContractCallInput, error) {
	isSCCall, newVMInput, err := sc.isSCExecutionAfterBuiltInFunc(tx, vmInput, vmOutput, acntDst)
	if !isSCCall {
		return false, vmOutput, vmInput, nil
	}

	userErrorVmOutput := &vmcommon.VMOutput{
		ReturnCode: vmcommon.UserError,
	}
	if err != nil {
		return true, userErrorVmOutput, newVMInput, sc.ProcessIfError(acntSnd, vmInput.CurrentTxHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	err = sc.checkUpgradePermission(acntDst, newVMInput)
	if err != nil {
		log.Debug("checkUpgradePermission", "error", err.Error())
		return true, userErrorVmOutput, newVMInput, sc.ProcessIfError(acntSnd, vmInput.CurrentTxHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	newVMOutput, err := sc.executeSmartContractCall(newVMInput, tx, newVMInput.CurrentTxHash, snapshot, acntSnd, acntDst)
	if err != nil {
		return true, userErrorVmOutput, newVMInput, err
	}
	if newVMOutput.ReturnCode != vmcommon.Ok {
		return true, newVMOutput, newVMInput, nil
	}

	return true, newVMOutput, newVMInput, nil
}

func (sc *scProcessor) isSCExecutionAfterBuiltInFunc(
	tx data.TransactionHandler,
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
	acntDst state.UserAccountHandler,
) (bool, *vmcommon.ContractCallInput, error) {
	if vmOutput.ReturnCode != vmcommon.Ok {
		return false, nil, nil
	}
	if check.IfNil(acntDst) {
		return false, nil, nil
	}
	if !core.IsSmartContractAddress(vmInput.RecipientAddr) {
		return false, nil, nil
	}

	outAcc, ok := vmOutput.OutputAccounts[string(vmInput.RecipientAddr)]
	if !ok {
		return false, nil, nil
	}
	if len(outAcc.OutputTransfers) != 1 {
		return false, nil, nil
	}

	callType := determineCallType(tx)
	txData := prependCallbackToTxDataIfAsyncCallBack(outAcc.OutputTransfers[0].Data, callType)

	function, arguments, err := sc.argsParser.ParseCallData(string(txData))
	if err != nil {
		return true, nil, err
	}

	newVMInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     vmInput.CallerAddr,
			Arguments:      arguments,
			CallValue:      big.NewInt(0),
			CallType:       callType,
			GasPrice:       vmInput.GasPrice,
			GasProvided:    vmOutput.GasRemaining - vmInput.GasLocked,
			GasLocked:      vmInput.GasLocked,
			OriginalTxHash: vmInput.OriginalTxHash,
			CurrentTxHash:  vmInput.CurrentTxHash,
		},
		RecipientAddr:     vmInput.RecipientAddr,
		Function:          function,
		AllowInitFunction: false,
	}

	fillWithESDTValue(vmInput, newVMInput)

	return true, newVMInput, nil
}

func fillWithESDTValue(fullVMInput *vmcommon.ContractCallInput, newVMInput *vmcommon.ContractCallInput) {
	if fullVMInput.Function != core.BuiltInFunctionESDTTransfer {
		return
	}

	newVMInput.ESDTTokenName = fullVMInput.Arguments[0]
	newVMInput.ESDTValue = big.NewInt(0).SetBytes(fullVMInput.Arguments[1])
}

func (sc *scProcessor) isCrossShardESDTTransfer(tx data.TransactionHandler) bool {
	sndShardID := sc.shardCoordinator.ComputeId(tx.GetSndAddr())
	if sndShardID == sc.shardCoordinator.SelfId() {
		return false
	}

	dstShardID := sc.shardCoordinator.ComputeId(tx.GetRcvAddr())
	if dstShardID == sndShardID {
		return false
	}

	function, _, err := sc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return false
	}

	return function == core.BuiltInFunctionESDTTransfer
}

// ProcessIfError creates a smart contract result, consumed the gas and returns the value to the user
func (sc *scProcessor) ProcessIfError(
	acntSnd state.UserAccountHandler,
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
	snapshot int,
	gasLocked uint64,
) error {
	err := sc.accounts.RevertToSnapshot(snapshot)
	if err != nil {
		log.Warn("revert to snapshot", "error", err.Error())
		return err
	}

	scrIfError, consumedFee := sc.createSCRsWhenError(acntSnd, txHash, tx, returnCode, returnMessage, gasLocked)
	err = sc.addBackTxValues(acntSnd, scrIfError, tx)
	if err != nil {
		return err
	}

	err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrIfError})
	if err != nil {
		return err
	}

	err = sc.processForRelayerWhenError(tx, txHash, returnMessage)
	if err != nil {
		return err
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)

	return nil
}

func (sc *scProcessor) processForRelayerWhenError(
	originalTx data.TransactionHandler,
	txHash []byte,
	returnMessage []byte,
) error {
	relayedSCR, isRelayed := isRelayedTx(originalTx)
	if !isRelayed {
		return nil
	}
	if relayedSCR.Value.Cmp(zero) == 0 {
		return nil
	}

	relayerAcnt, err := sc.getAccountFromAddress(relayedSCR.RelayerAddr)
	if err != nil {
		return err
	}

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.AddToBalance(relayedSCR.RelayedValue)
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
		PrevTxHash:     txHash,
		ReturnMessage:  returnMessage,
	}

	err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrForRelayer})
	if err != nil {
		return err
	}

	return nil
}

// transaction must be of type SCR and relayed address to be set with relayed value higher than 0
func isRelayedTx(tx data.TransactionHandler) (*smartContractResult.SmartContractResult, bool) {
	relayedSCR, ok := tx.(*smartContractResult.SmartContractResult)
	if !ok {
		return nil, false
	}

	if len(relayedSCR.RelayerAddr) == len(relayedSCR.SndAddr) {
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

	relayedSCR, isRelayed := isRelayedTx(originalTx)
	if isRelayed {
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

	shouldAllowDeploy := sc.flagDeploy.IsSet() || sc.isGenesisProcessing
	if !shouldAllowDeploy {
		log.Trace("deploy is disabled")
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, process.ErrSmartContractDeploymentIsDisabled.Error(), []byte(""), snapshot, 0)
	}

	vmInput, vmType, err := sc.createVMDeployInput(tx)
	if err != nil {
		log.Debug("Transaction error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, 0)
	}

	vmExec, err := sc.vmContainer.Get(vmType)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	vmOutput, err = vmExec.RunSmartContractCreate(vmInput)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Debug("run smart contract call error", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}
	vmOutput.GasRemaining += vmInput.GasLocked
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte("gas consumed exceeded"), snapshot, vmInput.GasLocked)
	}

	results, err := sc.processVMOutput(vmOutput, txHash, tx, vmInput.CallType, vmInput.GasProvided)
	if err != nil {
		log.Trace("Processing error", "error", err.Error())
		return vmOutput.ReturnCode, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return 0, err
	}
	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("AddIntermediate Transaction error", "error", err.Error())
		return 0, err
	}

	err = sc.updateDeveloperRewardsProxy(tx, vmOutput, 0)
	if err != nil {
		log.Debug("updateDeveloperRewardsProxy", "error", err.Error())
		return 0, err
	}

	totalConsumedFee, totalDevRwd := sc.computeTotalConsumedFeeAndDevRwd(tx, vmOutput, 0)
	sc.txFeeHandler.ProcessTransactionFee(totalConsumedFee, totalDevRwd, txHash)
	sc.printScDeployed(vmOutput, tx)
	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)

	return 0, nil
}

func (sc *scProcessor) updateDeveloperRewardsProxy(
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) error {
	if !sc.flagDeploy.IsSet() {
		return sc.updateDeveloperRewardsV1(tx, vmOutput, builtInFuncGasUsed)
	}

	return sc.updateDeveloperRewardsV2(tx, vmOutput, builtInFuncGasUsed)
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

	cost := sc.economicsFee.ComputeTxFee(tx)
	if !sc.flagPenalizedTooMuchGas.IsSet() {
		cost = core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}
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
	callType vmcommon.CallType,
	gasProvided uint64,
) ([]data.TransactionHandler, error) {

	sc.penalizeUserIfNeeded(tx, txHash, callType, gasProvided, vmOutput)
	scrForSender, scrForRelayer := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		callType,
	)

	outPutAccounts := process.SortVMOutputInsideData(vmOutput)
	createdAsyncCallback, scrTxs, err := sc.processSCOutputAccounts(vmOutput, callType, outPutAccounts, tx, txHash)
	if err != nil {
		return nil, err
	}

	if !check.IfNil(scrForRelayer) {
		scrTxs = append(scrTxs, scrForRelayer)
		err = sc.addToBalanceIfInShard(scrForRelayer.RcvAddr, scrForRelayer.Value)
		if err != nil {
			return nil, err
		}
	}

	if !createdAsyncCallback && callType == vmcommon.AsynchronousCall {
		asyncCallBackSCR := createAsyncCallBackSCRFromVMOutput(vmOutput, tx, txHash)
		scrTxs = append(scrTxs, asyncCallBackSCR)
	} else if !createdAsyncCallback {
		scrTxs = append(scrTxs, scrForSender)
	}

	err = sc.addToBalanceIfInShard(scrForSender.RcvAddr, scrForSender.Value)
	if err != nil {
		return nil, err
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, err
	}

	return scrTxs, nil
}

func (sc *scProcessor) addToBalanceIfInShard(address []byte, value *big.Int) error {
	userAcc, err := sc.getAccountFromAddress(address)
	if err != nil {
		return err
	}

	if check.IfNil(userAcc) {
		return nil
	}

	err = userAcc.AddToBalance(value)
	if err != nil {
		return err
	}

	return sc.accounts.SaveAccount(userAcc)
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
	if callType == vmcommon.AsynchronousCall {
		return
	}

	isTooMuchProvided := isTooMuchGasProvided(gasProvided, vmOutput.GasRemaining)
	if !isTooMuchProvided {
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
	acntSnd state.UserAccountHandler,
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
	gasLocked uint64,
) (*smartContractResult.SmartContractResult, *big.Int) {
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
		PrevTxHash:    txHash,
		ReturnMessage: returnMessage,
	}

	accumulatedSCRData := ""
	if sc.isCrossShardESDTTransfer(tx) {
		accumulatedSCRData += string(tx.GetData())
	}

	consumedFee := sc.economicsFee.ComputeTxFee(tx)
	if !sc.flagPenalizedTooMuchGas.IsSet() {
		consumedFee = core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	if !sc.flagDeploy.IsSet() {
		accumulatedSCRData += "@" + hex.EncodeToString([]byte(returnCode)) + "@" + hex.EncodeToString(txHash)
		if check.IfNil(acntSnd) {
			moveBalanceCost := sc.economicsFee.ComputeMoveBalanceFee(tx)
			consumedFee.Sub(consumedFee, moveBalanceCost)
		}
	} else {
		if callType == vmcommon.AsynchronousCall {
			scr.CallType = vmcommon.AsynchronousCallBack
			scr.GasPrice = tx.GetGasPrice()
			if tx.GetGasLimit() >= gasLocked {
				scr.GasLimit = gasLocked
				consumedFee = sc.economicsFee.ComputeFeeForProcessing(tx, tx.GetGasLimit()-gasLocked)
			}
			accumulatedSCRData += "@" + core.ConvertToEvenHex(int(vmcommon.UserError))
		} else {
			accumulatedSCRData += "@" + hex.EncodeToString([]byte(returnCode))
			if check.IfNil(acntSnd) {
				moveBalanceCost := sc.economicsFee.ComputeMoveBalanceFee(tx)
				consumedFee.Sub(consumedFee, moveBalanceCost)
			}
		}
	}

	scr.Data = []byte(accumulatedSCRData)
	setOriginalTxHash(scr, txHash, tx)
	if scr.Value == nil {
		scr.Value = big.NewInt(0)
	}
	if scr.Value.Cmp(zero) > 0 {
		scr.OriginalSender = tx.GetSndAddr()
	}

	return scr, consumedFee
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

func createBaseSCR(
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	result := &smartContractResult.SmartContractResult{}

	result.Value = big.NewInt(0)
	result.Nonce = outAcc.Nonce
	result.RcvAddr = outAcc.Address
	result.SndAddr = tx.GetRcvAddr()
	result.Code = outAcc.Code
	result.GasPrice = tx.GetGasPrice()
	result.PrevTxHash = txHash
	result.CallType = vmcommon.DirectCall
	setOriginalTxHash(result, txHash, tx)

	relayedTx, isRelayed := isRelayedTx(tx)
	if isRelayed {
		result.RelayedValue = big.NewInt(0)
		result.RelayerAddr = relayedTx.RelayerAddr
	}

	return result
}

func addVMOutputResultsToSCR(vmOutput *vmcommon.VMOutput, result *smartContractResult.SmartContractResult) {
	result.CallType = vmcommon.AsynchronousCallBack
	result.GasLimit = vmOutput.GasRemaining
	result.Data = []byte("@" + core.ConvertToEvenHex(int(vmOutput.ReturnCode)))
	addReturnDataToSCR(vmOutput, result)
}

func createAsyncCallBackSCRFromVMOutput(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	scr := &smartContractResult.SmartContractResult{
		Value:          big.NewInt(0),
		RcvAddr:        tx.GetSndAddr(),
		SndAddr:        tx.GetRcvAddr(),
		PrevTxHash:     txHash,
		GasPrice:       tx.GetGasPrice(),
		ReturnMessage:  []byte(vmOutput.ReturnMessage),
		OriginalSender: tx.GetSndAddr(),
	}
	setOriginalTxHash(scr, txHash, tx)
	relayedTx, isRelayed := isRelayedTx(tx)
	if isRelayed {
		scr.RelayedValue = big.NewInt(0)
		scr.RelayerAddr = relayedTx.RelayerAddr
	}

	addVMOutputResultsToSCR(vmOutput, scr)

	return scr
}

func (sc *scProcessor) createSmartContractResults(
	vmOutput *vmcommon.VMOutput,
	callType vmcommon.CallType,
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) (bool, []data.TransactionHandler) {

	if bytes.Equal(outAcc.Address, vm.StakingSCAddress) {
		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		result := createBaseSCR(outAcc, tx, txHash)
		result.Data = append(result.Data, sc.argsParser.CreateDataFromStorageUpdate(storageUpdates)...)

		return false, []data.TransactionHandler{result}
	}

	lenOutTransfers := len(outAcc.OutputTransfers)
	if lenOutTransfers == 0 {
		if callType == vmcommon.AsynchronousCall && bytes.Equal(outAcc.Address, tx.GetSndAddr()) {
			result := createBaseSCR(outAcc, tx, txHash)
			addVMOutputResultsToSCR(vmOutput, result)
			return true, []data.TransactionHandler{result}
		}

		if !sc.flagDeploy.IsSet() {
			result := createBaseSCR(outAcc, tx, txHash)
			result.Code = outAcc.Code
			result.Value.Set(outAcc.BalanceDelta)
			if result.Value.Cmp(zero) > 0 {
				result.OriginalSender = tx.GetSndAddr()
			}

			return false, []data.TransactionHandler{result}
		}

		return false, nil
	}

	createdAsyncCallBack := false
	var result *smartContractResult.SmartContractResult
	scResults := make([]data.TransactionHandler, 0, len(outAcc.OutputTransfers))
	for i, outputTransfer := range outAcc.OutputTransfers {
		result = createBaseSCR(outAcc, tx, txHash)

		if outputTransfer.Value != nil {
			result.Value.Set(outputTransfer.Value)
		}
		result.Data = outputTransfer.Data
		result.GasLimit = outputTransfer.GasLimit
		result.CallType = outputTransfer.CallType
		setOriginalTxHash(result, txHash, tx)
		if result.Value.Cmp(zero) > 0 {
			result.OriginalSender = tx.GetSndAddr()
		}

		outputTransferCopy := outputTransfer
		isLastOutTransfer := i == lenOutTransfers-1
		if isLastOutTransfer &&
			sc.useLastTransferAsAsyncCallBackWhenNeeded(callType, outAcc, &outputTransferCopy, vmOutput, tx, result) {
			createdAsyncCallBack = true
		}

		if result.CallType == vmcommon.AsynchronousCall {
			result.GasLimit += outputTransfer.GasLocked
			lastArgAsGasLocked := "@" + hex.EncodeToString(big.NewInt(0).SetUint64(outputTransfer.GasLocked).Bytes())
			result.Data = append(result.Data, []byte(lastArgAsGasLocked)...)
		}

		scResults = append(scResults, result)
	}

	return createdAsyncCallBack, scResults
}

func (sc *scProcessor) useLastTransferAsAsyncCallBackWhenNeeded(
	callType vmcommon.CallType,
	outAcc *vmcommon.OutputAccount,
	outputTransfer *vmcommon.OutputTransfer,
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	result *smartContractResult.SmartContractResult,
) bool {
	if len(vmOutput.ReturnData) > 0 {
		return false
	}

	isAsyncTransferBackToSender := callType == vmcommon.AsynchronousCall &&
		bytes.Equal(outAcc.Address, tx.GetSndAddr())
	if !isAsyncTransferBackToSender {
		return false
	}

	if !sc.isTransferWithNoDataOrBuiltInCall(outputTransfer.Data) {
		return false
	}

	result.CallType = vmcommon.AsynchronousCallBack
	result.GasLimit += vmOutput.GasRemaining

	return true
}

func (sc *scProcessor) isTransferWithNoDataOrBuiltInCall(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	function, _, err := sc.argsParser.ParseCallData(string(data))
	if err != nil {
		return false
	}

	_, err = sc.builtInFunctions.Get(function)
	return err == nil
}

// createSCRForSender(vmOutput, tx, txHash, acntSnd)
// give back the user the unused gas money
func (sc *scProcessor) createSCRForSenderAndRelayer(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte,
	callType vmcommon.CallType,
) (*smartContractResult.SmartContractResult, *smartContractResult.SmartContractResult) {
	if vmOutput.GasRefund == nil {
		// TODO: compute gas refund with reduced gasPrice if we need to activate this
		vmOutput.GasRefund = big.NewInt(0)
	}

	gasRefund := sc.economicsFee.ComputeFeeForProcessing(tx, vmOutput.GasRemaining)
	gasRemaining := uint64(0)
	storageFreeRefund := big.NewInt(0)
	// backward compatibility - there should be no refund as the storage pay was already distributed among validators
	// this would only create additional inflation
	// backward compatibility - direct smart contract results were created with gasLimit - there is no need for them
	if !sc.flagDeploy.IsSet() {
		storageFreeRefund = big.NewInt(0).Mul(vmOutput.GasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
		gasRemaining = vmOutput.GasRemaining
	}

	rcvAddress := tx.GetSndAddr()
	if callType == vmcommon.AsynchronousCallBack {
		rcvAddress = tx.GetRcvAddr()
	}

	var refundGasToRelayerSCR *smartContractResult.SmartContractResult
	relayedSCR, isRelayed := isRelayedTx(tx)
	shouldRefundGasToRelayerSCR := isRelayed && callType != vmcommon.AsynchronousCall && gasRefund.Cmp(zero) > 0
	if shouldRefundGasToRelayerSCR {
		senderForRelayerRefund := tx.GetRcvAddr()
		if !sc.isSelfShard(tx.GetRcvAddr()) {
			senderForRelayerRefund = tx.GetSndAddr()
		}

		refundGasToRelayerSCR = &smartContractResult.SmartContractResult{
			Nonce:          relayedSCR.Nonce + 1,
			Value:          big.NewInt(0).Set(gasRefund),
			RcvAddr:        relayedSCR.RelayerAddr,
			SndAddr:        senderForRelayerRefund,
			PrevTxHash:     txHash,
			OriginalTxHash: relayedSCR.OriginalTxHash,
			GasPrice:       tx.GetGasPrice(),
			CallType:       vmcommon.DirectCall,
			ReturnMessage:  []byte("gas refund for relayer"),
			OriginalSender: relayedSCR.OriginalSender,
		}
		gasRemaining = 0
	}

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = big.NewInt(0).Set(storageFreeRefund)
	if callType != vmcommon.AsynchronousCall && check.IfNil(refundGasToRelayerSCR) {
		scTx.Value.Add(scTx.Value, gasRefund)
	}

	scTx.RcvAddr = rcvAddress
	scTx.SndAddr = tx.GetRcvAddr()
	scTx.Nonce = tx.GetNonce() + 1
	scTx.PrevTxHash = txHash
	scTx.GasLimit = gasRemaining
	scTx.GasPrice = tx.GetGasPrice()
	scTx.ReturnMessage = []byte(vmOutput.ReturnMessage)
	scTx.CallType = vmcommon.DirectCall
	setOriginalTxHash(scTx, txHash, tx)
	scTx.Data = []byte("@" + hex.EncodeToString([]byte(vmOutput.ReturnCode.String())))

	// when asynchronous call - the callback is created by combining the last output transfer with the returnData
	if callType != vmcommon.AsynchronousCall {
		addReturnDataToSCR(vmOutput, scTx)
	}

	log.Trace("createSCRForSenderAndRelayer ", "data", string(scTx.Data), "snd", scTx.SndAddr, "rcv", scTx.RcvAddr)
	return scTx, refundGasToRelayerSCR
}

func addReturnDataToSCR(vmOutput *vmcommon.VMOutput, scTx *smartContractResult.SmartContractResult) {
	for _, retData := range vmOutput.ReturnData {
		scTx.Data = append(scTx.Data, []byte("@"+hex.EncodeToString(retData))...)
	}
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(
	vmOutput *vmcommon.VMOutput,
	callType vmcommon.CallType,
	outputAccounts []*vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) (bool, []data.TransactionHandler, error) {
	scResults := make([]data.TransactionHandler, 0, len(outputAccounts))

	sumOfAllDiff := big.NewInt(0)
	sumOfAllDiff.Sub(sumOfAllDiff, tx.GetValue())

	createdAsyncCallback := false
	for _, outAcc := range outputAccounts {
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return false, nil, err
		}

		tmpCreatedAsyncCallback, newScrs := sc.createSmartContractResults(vmOutput, callType, outAcc, tx, txHash)
		createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback

		if len(newScrs) != 0 {
			scResults = append(scResults, newScrs...)
		}
		if check.IfNil(acc) {
			if outAcc.BalanceDelta != nil {
				sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
			}
			continue
		}

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for j := 0; j < len(storageUpdates); j++ {
			storeUpdate := storageUpdates[j]
			if !process.IsAllowedToSaveUnderKey(storeUpdate.Offset) {
				log.Trace("storeUpdate is not allowed", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
				continue
			}

			err = acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				log.Warn("saveKeyValue", "error", err)
				return false, nil, err
			}
			log.Trace("storeUpdate", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
		}

		sc.updateSmartContractCode(acc, outAcc)
		// change nonce only if there is a change
		if outAcc.Nonce != acc.GetNonce() && outAcc.Nonce != 0 {
			if outAcc.Nonce < acc.GetNonce() {
				return false, nil, process.ErrWrongNonceInVMOutput
			}

			nonceDifference := outAcc.Nonce - acc.GetNonce()
			acc.IncreaseNonce(nonceDifference)
		}

		// if no change then continue
		if outAcc.BalanceDelta == nil || outAcc.BalanceDelta.Cmp(zero) == 0 {
			err = sc.accounts.SaveAccount(acc)
			if err != nil {
				return false, nil, err
			}

			continue
		}

		sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)

		err = acc.AddToBalance(outAcc.BalanceDelta)
		if err != nil {
			return false, nil, err
		}

		err = sc.accounts.SaveAccount(acc)
		if err != nil {
			return false, nil, err
		}
	}

	if sumOfAllDiff.Cmp(zero) != 0 {
		return false, nil, process.ErrOverallBalanceChangeFromSC
	}

	return createdAsyncCallback, scResults, nil
}

// updateSmartContractCode upgrades code for "direct" deployments & upgrades and for "indirect" deployments & upgrades
// It receives:
// 	(1) the account as found in the State
//	(2) the account as returned in VM Output
// 	(3) the transaction that, upon execution, produced the VM Output
func (sc *scProcessor) updateSmartContractCode(
	stateAccount state.UserAccountHandler,
	outputAccount *vmcommon.OutputAccount,
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

	currentOwner := stateAccount.GetOwnerAddress()
	isCodeDeployerSet := len(outputAccount.CodeDeployerAddress) > 0
	isCodeDeployerOwner := bytes.Equal(currentOwner, outputAccount.CodeDeployerAddress) && isCodeDeployerSet

	noExistingCode := len(stateAccount.GetCode()) == 0
	noExistingOwner := len(currentOwner) == 0
	currentCodeMetadata := vmcommon.CodeMetadataFromBytes(stateAccount.GetCodeMetadata())
	newCodeMetadata := vmcommon.CodeMetadataFromBytes(outputAccount.CodeMetadata)
	isUpgradeable := currentCodeMetadata.Upgradeable
	isDeployment := noExistingCode && noExistingOwner
	isUpgrade := !isDeployment && isCodeDeployerOwner && isUpgradeable

	if isDeployment {
		// At this point, we are under the condition "noExistingOwner"
		stateAccount.SetOwnerAddress(outputAccount.CodeDeployerAddress)
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

	gasLocked := sc.getGasLockedFromSCR(scr)

	txType, _ := sc.txTypeHandler.ComputeTransactionType(scr)
	switch txType {
	case process.MoveBalance:
		err = sc.processSimpleSCR(scr, dstAcc)
		if err != nil {
			return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot, gasLocked)
		}
		return vmcommon.Ok, nil
	case process.SCDeployment:
		err = process.ErrSCDeployFromSCRIsNotPermitted
		return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot, gasLocked)
	case process.SCInvoking:
		returnCode, err = sc.ExecuteSmartContractTransaction(scr, sndAcc, dstAcc)
		return returnCode, err
	case process.BuiltInFunctionCall:
		returnCode, err = sc.ExecuteBuiltInFunction(scr, sndAcc, dstAcc)
		return returnCode, err
	}

	err = process.ErrWrongTransaction
	return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot, gasLocked)
}

func (sc *scProcessor) getGasLockedFromSCR(scr *smartContractResult.SmartContractResult) uint64 {
	if scr.CallType != vmcommon.AsynchronousCall {
		return 0
	}

	_, arguments, err := sc.argsParser.ParseCallData(string(scr.Data))
	if err != nil {
		return 0
	}
	_, gasLocked := sc.getAsyncCallGasLockFromTxData(scr.CallType, arguments)
	return gasLocked
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

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	return sc == nil
}
