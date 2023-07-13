package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

var _ process.SmartContractResultProcessor = (*scProcessor)(nil)
var _ process.SmartContractProcessor = (*scProcessor)(nil)

var log = logger.GetOrCreate("process/smartcontract")
var logCounters = logger.GetOrCreate("process/smartcontract.blockchainHookCounters")

const maxTotalSCRsSize = 3 * (1 << 18) // 768KB

const (
	// TooMuchGasProvidedMessage is the message for the too much gas provided error
	TooMuchGasProvidedMessage = "too much gas provided"

	executeDurationAlarmThreshold = time.Duration(100) * time.Millisecond

	// TODO: Move to vm-common.
	upgradeFunctionName = "upgradeContract"
	returnOkData        = "@6f6b"
)

var zero = big.NewInt(0)

type scProcessor struct {
	accounts           state.AccountsAdapter
	blockChainHook     process.BlockChainHookHandler
	pubkeyConv         core.PubkeyConverter
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
	shardCoordinator   sharding.Coordinator
	vmContainer        process.VirtualMachinesContainer
	argsParser         process.ArgumentsParser
	esdtTransferParser vmcommon.ESDTTransferParser
	builtInFunctions   vmcommon.BuiltInFunctionContainer
	wasmVMChangeLocker common.Locker

	enableRoundsHandler process.EnableRoundsHandler
	enableEpochsHandler common.EnableEpochsHandler
	badTxForwarder      process.IntermediateTransactionHandler
	scrForwarder        process.IntermediateTransactionHandler
	txFeeHandler        process.TransactionFeeHandler
	economicsFee        process.FeeHandler
	txTypeHandler       process.TxTypeHandler
	gasHandler          process.GasHandler

	builtInGasCosts     map[string]uint64
	persistPerByte      uint64
	storePerByte        uint64
	mutGasLock          sync.RWMutex
	txLogsProcessor     process.TransactionLogProcessor
	vmOutputCacher      storage.Cacher
	isGenesisProcessing bool

	executableCheckers    map[string]scrCommon.ExecutableChecker
	mutExecutableCheckers sync.RWMutex
}

// ArgsNewSmartContractProcessor defines the arguments needed for new smart contract processor
type ArgsNewSmartContractProcessor struct {
	VmContainer         process.VirtualMachinesContainer
	ArgsParser          process.ArgumentsParser
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	AccountsDB          state.AccountsAdapter
	BlockChainHook      process.BlockChainHookHandler
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	PubkeyConv          core.PubkeyConverter
	ShardCoordinator    sharding.Coordinator
	ScrForwarder        process.IntermediateTransactionHandler
	TxFeeHandler        process.TransactionFeeHandler
	EconomicsFee        process.FeeHandler
	TxTypeHandler       process.TxTypeHandler
	GasHandler          process.GasHandler
	GasSchedule         core.GasScheduleNotifier
	TxLogsProcessor     process.TransactionLogProcessor
	EnableEpochsHandler common.EnableEpochsHandler
	BadTxForwarder      process.IntermediateTransactionHandler
	VMOutputCacher      storage.Cacher
	WasmVMChangeLocker  common.Locker
	IsGenesisProcessing bool
}

// NewSmartContractProcessor creates a smart contract processor that creates and interprets VM data
func NewSmartContractProcessor(args scrCommon.ArgsNewSmartContractProcessor) (*scProcessor, error) {
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
	if check.IfNil(args.ShardCoordinator) {
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
	if check.IfNil(args.TxLogsProcessor) {
		return nil, process.ErrNilTxLogsProcessor
	}
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.BadTxForwarder) {
		return nil, process.ErrNilBadTxHandler
	}
	if check.IfNilReflect(args.WasmVMChangeLocker) {
		return nil, process.ErrNilLocker
	}
	if check.IfNil(args.VMOutputCacher) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(args.BuiltInFunctions) {
		return nil, process.ErrNilBuiltInFunction
	}

	builtInFuncCost := args.GasSchedule.LatestGasSchedule()[common.BuiltInCost]
	baseOperationCost := args.GasSchedule.LatestGasSchedule()[common.BaseOperationCost]
	sc := &scProcessor{
		vmContainer:         args.VmContainer,
		argsParser:          args.ArgsParser,
		hasher:              args.Hasher,
		marshalizer:         args.Marshalizer,
		accounts:            args.AccountsDB,
		blockChainHook:      args.BlockChainHook,
		pubkeyConv:          args.PubkeyConv,
		shardCoordinator:    args.ShardCoordinator,
		scrForwarder:        args.ScrForwarder,
		txFeeHandler:        args.TxFeeHandler,
		economicsFee:        args.EconomicsFee,
		txTypeHandler:       args.TxTypeHandler,
		gasHandler:          args.GasHandler,
		builtInGasCosts:     builtInFuncCost,
		txLogsProcessor:     args.TxLogsProcessor,
		enableRoundsHandler: args.EnableRoundsHandler,
		enableEpochsHandler: args.EnableEpochsHandler,
		badTxForwarder:      args.BadTxForwarder,
		builtInFunctions:    args.BuiltInFunctions,
		isGenesisProcessing: args.IsGenesisProcessing,
		wasmVMChangeLocker:  args.WasmVMChangeLocker,
		vmOutputCacher:      args.VMOutputCacher,
		storePerByte:        baseOperationCost["StorePerByte"],
		persistPerByte:      baseOperationCost["PersistPerByte"],
		executableCheckers:  scrCommon.CreateExecutableCheckersMap(args.BuiltInFunctions),
	}

	var err error
	sc.esdtTransferParser, err = parsers.NewESDTTransferParser(args.Marshalizer)
	if err != nil {
		return nil, err
	}

	args.GasSchedule.RegisterNotifyHandler(sc)

	return sc, nil
}

// GasScheduleChange sets the new gas schedule where it is needed
// Warning: do not use flags in this function as it will raise backward compatibility issues because the GasScheduleChange
// is not called on each epoch change
func (sc *scProcessor) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	sc.mutGasLock.Lock()
	defer sc.mutGasLock.Unlock()

	builtInFuncCost := gasSchedule[common.BuiltInCost]
	if builtInFuncCost == nil {
		return
	}

	sc.builtInGasCosts = builtInFuncCost
	sc.storePerByte = gasSchedule[common.BaseOperationCost]["StorePerByte"]
	sc.persistPerByte = gasSchedule[common.BaseOperationCost]["PersistPerByte"]
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
		txHash := sc.computeTxHashUnsafe(tx)
		log.Debug(fmt.Sprintf("scProcessor.ExecuteSmartContractTransaction(): execution took > %s", executeDurationAlarmThreshold), "tx hash", txHash, "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
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

	vmOutput, err := sc.executeSmartContractCall(vmInput, tx, txHash, snapshot, acntSnd, acntDst, nil)
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
	results, err = sc.processVMOutput(&vmInput.VMInput, vmOutput, txHash, tx)
	if err != nil {
		if core.IsGetNodeFromDBError(err) {
			return vmcommon.ExecutionFailed, err
		}
		log.Trace("process vm output returned with problem ", "err", err.Error())
		return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	return sc.finishSCExecution(results, txHash, tx, vmOutput, 0)
}

func (sc *scProcessor) executeSmartContractCall(
	vmInput *vmcommon.ContractCallInput,
	tx data.TransactionHandler,
	txHash []byte,
	snapshot int,
	acntSnd, acntDst state.UserAccountHandler,
	prevVmOutput *vmcommon.VMOutput,
) (*vmcommon.VMOutput, error) {
	if check.IfNil(acntDst) {
		return nil, process.ErrNilSCDestAccount
	}

	sc.wasmVMChangeLocker.RLock()

	userErrorVmOutput := &vmcommon.VMOutput{
		ReturnCode: vmcommon.UserError,
	}
	vmExec, _, err := scrCommon.FindVMByScAddress(sc.vmContainer, vmInput.RecipientAddr)
	if err != nil {
		sc.wasmVMChangeLocker.RUnlock()
		returnMessage := "cannot get vm from address"
		log.Trace("get vm from address error", "error", err.Error())
		return userErrorVmOutput, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot, vmInput.GasLocked)
	}

	sc.blockChainHook.ResetCounters()
	defer sc.printBlockchainHookCounters(tx)

	var vmOutput *vmcommon.VMOutput
	vmOutput, err = vmExec.RunSmartContractCall(vmInput)

	sc.wasmVMChangeLocker.RUnlock()
	if err != nil {
		if core.IsGetNodeFromDBError(err) {
			return nil, err
		}
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
		return userErrorVmOutput, sc.processIfErrorWithAddedLogs(acntSnd, txHash, tx, vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked, prevVmOutput, vmOutput.Logs)
	}
	acntSnd, err = sc.reloadLocalAccount(acntSnd) // nolint
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return nil, err
	}

	return vmOutput, nil
}

func (sc *scProcessor) isInformativeTxHandler(txHandler data.TransactionHandler) bool {
	if txHandler.GetValue().Cmp(zero) > 0 {
		return false
	}

	scr, ok := txHandler.(*smartContractResult.SmartContractResult)
	if ok && scr.CallType == vmData.AsynchronousCallBack {
		return false
	}

	function, _, err := sc.argsParser.ParseCallData(string(txHandler.GetData()))
	if err != nil {
		return true
	}

	if core.IsSmartContractAddress(txHandler.GetRcvAddr()) {
		return false
	}

	_, err = sc.builtInFunctions.Get(function)
	return err != nil
}

func (sc *scProcessor) printBlockchainHookCounters(tx data.TransactionHandler) {
	if logCounters.GetLevel() > logger.LogTrace {
		return
	}

	logCounters.Trace("blockchain hook counters",
		"counters", sc.getBlockchainHookCountersString(),
		"tx hash", sc.computeTxHashUnsafe(tx),
		"receiver", sc.pubkeyConv.SilentEncode(tx.GetRcvAddr(), log),
		"sender", sc.pubkeyConv.SilentEncode(tx.GetSndAddr(), log),
		"value", tx.GetValue().String(),
		"data", tx.GetData(),
	)
}

func (sc *scProcessor) getBlockchainHookCountersString() string {
	counters := sc.blockChainHook.GetCounterValues()
	keys := make([]string, len(counters))

	idx := 0
	for key := range counters {
		keys[idx] = key
		idx++
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	lines := make([]string, 0, len(counters))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %d", key, counters[key]))
	}

	return strings.Join(lines, ", ")
}

func (sc *scProcessor) cleanInformativeOnlySCRs(scrs []data.TransactionHandler) ([]data.TransactionHandler, []*vmcommon.LogEntry) {
	cleanedUPSCrs := make([]data.TransactionHandler, 0)
	logsFromSCRs := make([]*vmcommon.LogEntry, 0)

	for _, scr := range scrs {
		if sc.isInformativeTxHandler(scr) {
			logsFromSCRs = append(logsFromSCRs, createNewLogFromSCR(scr))
			continue
		}

		cleanedUPSCrs = append(cleanedUPSCrs, scr)
	}

	if !sc.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabled() {
		return scrs, logsFromSCRs
	}

	return cleanedUPSCrs, logsFromSCRs
}

func (sc *scProcessor) finishSCExecution(
	results []data.TransactionHandler,
	txHash []byte,
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) (vmcommon.ReturnCode, error) {
	resultWithoutMeta := sc.deleteSCRsWithValueZeroGoingToMeta(results)
	finalResults, logsFromSCRs := sc.cleanInformativeOnlySCRs(resultWithoutMeta)

	err := sc.updateDeveloperRewardsProxy(tx, vmOutput, builtInFuncGasUsed)
	if err != nil {
		log.Error("updateDeveloperRewardsProxy", "error", err.Error())
		return 0, err
	}

	err = sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Error("AddIntermediateTransactions error", "error", err.Error())
		return 0, err
	}

	vmOutput.Logs = append(vmOutput.Logs, logsFromSCRs...)
	completedTxLog := sc.createCompleteEventLogIfNoMoreAction(tx, txHash, finalResults)
	if completedTxLog != nil {
		vmOutput.Logs = append(vmOutput.Logs, completedTxLog)
	}

	ignorableError := sc.txLogsProcessor.SaveLog(txHash, tx, vmOutput.Logs)
	if ignorableError != nil {
		log.Debug("scProcessor.finishSCExecution txLogsProcessor.SaveLog()", "error", ignorableError.Error())
	}

	totalConsumedFee, totalDevRwd := sc.computeTotalConsumedFeeAndDevRwd(tx, vmOutput, builtInFuncGasUsed)
	sc.txFeeHandler.ProcessTransactionFee(totalConsumedFee, totalDevRwd, txHash)
	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)

	sc.vmOutputCacher.Put(txHash, vmOutput, 0)

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

			sentGas, err = core.SafeAddUint64(sentGas, outTransfer.GasLocked)
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
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) && !sc.isSelfShard(tx.GetSndAddr()) {
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
	var devRwd *big.Int
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if sc.enableEpochsHandler.IsStakingV2FlagEnabledAfterEpoch(currentEpoch) {
		devRwd = core.GetIntTrimmedPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())
	} else {
		devRwd = core.GetApproximatePercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())
	}
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
	addressShardID := sc.shardCoordinator.ComputeId(address)
	if !sc.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabled() && core.IsEmptyAddress(address) {
		addressShardID = 0
	}

	return addressShardID == sc.shardCoordinator.SelfId()
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

	accumulatedGasUsedForOtherShard := uint64(0)
	for _, outAcc := range vmOutput.OutputAccounts {
		if !sc.isSelfShard(outAcc.Address) {
			sentGas := uint64(0)
			for _, outTransfer := range outAcc.OutputTransfers {
				sentGas, _ = core.SafeAddUint64(sentGas, outTransfer.GasLimit)
				sentGas, _ = core.SafeAddUint64(sentGas, outTransfer.GasLocked)
			}
			displayConsumedGas := consumedGas
			consumedGas, err = core.SafeSubUint64(consumedGas, sentGas)
			log.LogIfError(err,
				"function", "computeTotalConsumedFeeAndDevRwd",
				"resulted consumedGas", consumedGas,
				"consumedGas", displayConsumedGas,
				"sentGas", sentGas,
			)
			displayConsumedGas = consumedGas
			consumedGas, err = core.SafeAddUint64(consumedGas, outAcc.GasUsed)
			log.LogIfError(err,
				"function", "computeTotalConsumedFeeAndDevRwd",
				"resulted consumedGas", consumedGas,
				"consumedGas", displayConsumedGas,
				"outAcc.GasUsed", outAcc.GasUsed,
			)
			displayAccumulatedGasUsedForOtherShard := accumulatedGasUsedForOtherShard
			accumulatedGasUsedForOtherShard, err = core.SafeAddUint64(accumulatedGasUsedForOtherShard, outAcc.GasUsed)
			log.LogIfError(err,
				"function", "computeTotalConsumedFeeAndDevRwd",
				"resulted accumulatedGasUsedForOtherShard", accumulatedGasUsedForOtherShard,
				"accumulatedGasUsedForOtherShard", displayAccumulatedGasUsedForOtherShard,
				"outAcc.GasUsed", outAcc.GasUsed,
			)
		}
	}

	moveBalanceGasLimit := sc.economicsFee.ComputeGasLimit(tx)
	if !isSmartContractResult(tx) {
		displayConsumedGas := consumedGas
		consumedGas, err = core.SafeSubUint64(consumedGas, moveBalanceGasLimit)
		log.LogIfError(err,
			"function", "computeTotalConsumedFeeAndDevRwd",
			"resulted consumedGas", consumedGas,
			"consumedGas", displayConsumedGas,
			"moveBalanceGasLimit", moveBalanceGasLimit,
		)
	}

	consumedGasWithoutBuiltin, err := core.SafeSubUint64(consumedGas, accumulatedGasUsedForOtherShard)
	log.LogIfError(err,
		"function", "computeTotalConsumedFeeAndDevRwd",
		"resulted consumedGasWithoutBuiltin", consumedGasWithoutBuiltin,
		"consumedGas", consumedGas,
		"accumulatedGasUsedForOtherShard", accumulatedGasUsedForOtherShard,
	)
	if senderInSelfShard {
		displayConsumedGasWithoutBuiltin := consumedGasWithoutBuiltin
		consumedGasWithoutBuiltin, err = core.SafeSubUint64(consumedGasWithoutBuiltin, builtInFuncGasUsed)
		log.LogIfError(err,
			"function", "computeTotalConsumedFeeAndDevRwd",
			"resulted consumedGasWithoutBuiltin", consumedGasWithoutBuiltin,
			"consumedGasWithoutBuiltin", displayConsumedGasWithoutBuiltin,
			"builtInFuncGasUsed", builtInFuncGasUsed,
		)
	}

	totalFee := sc.economicsFee.ComputeFeeForProcessing(tx, consumedGas)
	totalFeeMinusBuiltIn := sc.economicsFee.ComputeFeeForProcessing(tx, consumedGasWithoutBuiltin)

	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	var totalDevRwd *big.Int
	if sc.enableEpochsHandler.IsStakingV2FlagEnabledAfterEpoch(currentEpoch) {
		totalDevRwd = core.GetIntTrimmedPercentageOfValue(totalFeeMinusBuiltIn, sc.economicsFee.DeveloperPercentage())
	} else {
		totalDevRwd = core.GetApproximatePercentageOfValue(totalFeeMinusBuiltIn, sc.economicsFee.DeveloperPercentage())
	}

	if !isSmartContractResult(tx) && senderInSelfShard {
		totalFee.Add(totalFee, sc.economicsFee.ComputeMoveBalanceFee(tx))
	}

	if !sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		totalDevRwd = core.GetApproximatePercentageOfValue(totalFee, sc.economicsFee.DeveloperPercentage())
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
			_, err := sc.getESDTParsedTransfers(scr.GetSndAddr(), scr.GetRcvAddr(), scr.GetData())
			if err != nil {
				continue
			}
		}
		cleanSCRs = append(cleanSCRs, scr)
	}

	return cleanSCRs
}

func (sc *scProcessor) saveAccounts(acntSnd, acntDst vmcommon.AccountHandler) error {
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
	txTypeOnDst process.TransactionType,
	function string,
	gasProvided uint64,
	gasRemaining uint64,
	isCrossShard bool,
) (uint64, error) {
	if txTypeOnDst != process.SCInvoking {
		return core.SafeSubUint64(gasProvided, gasRemaining)
	}

	isFixAsyncCallBackArgumentsParserFlagSet := sc.enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabled()
	if isFixAsyncCallBackArgumentsParserFlagSet && isCrossShard {
		return 0, nil
	}

	sc.mutGasLock.RLock()
	builtInFuncGasUsed := sc.builtInGasCosts[function]
	sc.mutGasLock.RUnlock()

	return builtInFuncGasUsed, nil
}

// ExecuteBuiltInFunction  processes the transaction, executes the built in function call and subsequent results
func (sc *scProcessor) ExecuteBuiltInFunction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	sw := core.NewStopWatch()
	sw.Start("executeBuiltIn")
	returnCode, err := sc.doExecuteBuiltInFunction(tx, acntSnd, acntDst)
	sw.Stop("executeBuiltIn")
	duration := sw.GetMeasurement("executeBuiltIn")

	if duration > executeDurationAlarmThreshold {
		txHash := sc.computeTxHashUnsafe(tx)
		log.Debug(fmt.Sprintf("scProcessor.ExecuteBuiltInFunction(): execution took > %s", executeDurationAlarmThreshold), "tx hash", txHash, "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
	} else {
		log.Trace("scProcessor.ExecuteBuiltInFunction()", "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
	}

	return returnCode, err
}

func (sc *scProcessor) doExecuteBuiltInFunction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	sc.blockChainHook.ResetCounters()
	defer sc.printBlockchainHookCounters(tx)

	returnCode, vmInput, txHash, err := sc.prepareExecution(tx, acntSnd, acntDst, true)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}

	snapshot := sc.accounts.JournalLen()
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsBuiltInFunctionsFlagEnabledInEpoch(currentEpoch) {
		return vmcommon.UserError, sc.resolveFailedTransaction(acntSnd, tx, txHash, process.ErrBuiltInFunctionsAreDisabled.Error(), snapshot)
	}

	var vmOutput *vmcommon.VMOutput
	vmOutput, err = sc.resolveBuiltInFunctions(vmInput)
	if err != nil {
		log.Debug("processed built in functions error", "error", err.Error())
		return 0, err
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return 0, err
	}

	if vmInput.ReturnCallAfterError && vmInput.CallType != vmData.AsynchronousCallBack {
		return sc.finishSCExecution(make([]data.TransactionHandler, 0), txHash, tx, vmOutput, 0)
	}

	_, txTypeOnDst := sc.txTypeHandler.ComputeTransactionType(tx)
	builtInFuncGasUsed, err := sc.computeBuiltInFuncGasUsed(txTypeOnDst, vmInput.Function, vmInput.GasProvided, vmOutput.GasRemaining, check.IfNil(acntSnd))
	log.LogIfError(err, "function", "ExecuteBuiltInFunction.computeBuiltInFuncGasUsed")

	if txTypeOnDst != process.SCInvoking {
		vmOutput.GasRemaining += vmInput.GasLocked
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		if !check.IfNil(acntSnd) {
			return vmcommon.UserError, sc.resolveFailedTransaction(acntSnd, tx, txHash, vmOutput.ReturnMessage, snapshot)
		}
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	if vmInput.CallType == vmData.AsynchronousCallBack {
		// in case of asynchronous callback - the process of built-in function is a must
		snapshot = sc.accounts.JournalLen()
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
		tmpCreatedAsyncCallback, scTxs, err := sc.createSmartContractResults(&vmInput.VMInput, vmOutput, outAcc, tx, txHash)
		if err != nil {
			return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(err.Error()), snapshot, vmInput.GasLocked)
		}
		createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback
		if len(scTxs) > 0 {
			scrResults = append(scrResults, scTxs...)
		}
	}

	isSCCallSelfShard, newVMOutput, newVMInput, err := sc.treatExecutionAfterBuiltInFunc(tx, vmInput, vmOutput, acntSnd, snapshot)
	if err != nil {
		log.Debug("treat execution after built in function", "error", err.Error())
		return 0, err
	}
	if newVMOutput.ReturnCode != vmcommon.Ok {
		return vmcommon.UserError, nil
	}

	if isSCCallSelfShard {
		err = sc.gasConsumedChecks(tx, newVMInput.GasProvided, newVMInput.GasLocked, newVMOutput)
		if err != nil {
			log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
			return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte("gas consumed exceeded"), snapshot, vmInput.GasLocked)
		}

		if sc.enableEpochsHandler.IsRepairCallbackFlagEnabledInEpoch(currentEpoch) {
			sc.penalizeUserIfNeeded(tx, txHash, newVMInput.CallType, newVMInput.GasProvided, newVMOutput)
		}

		outPutAccounts := process.SortVMOutputInsideData(newVMOutput)
		var newSCRTxs []data.TransactionHandler
		tmpCreatedAsyncCallback := false
		tmpCreatedAsyncCallback, newSCRTxs, err = sc.processSCOutputAccounts(&vmInput.VMInput, newVMOutput, outPutAccounts, tx, txHash)
		if err != nil {
			if core.IsGetNodeFromDBError(err) {
				return vmcommon.ExecutionFailed, err
			}
			return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(err.Error()), snapshot, vmInput.GasLocked)
		}
		createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback

		if len(newSCRTxs) > 0 {
			scrResults = append(scrResults, newSCRTxs...)
		}

		mergeVMOutputLogs(newVMOutput, vmOutput)
	}

	isSCCallCrossShard := !isSCCallSelfShard && txTypeOnDst == process.SCInvoking
	if !isSCCallCrossShard {
		if sc.enableEpochsHandler.IsRepairCallbackFlagEnabledInEpoch(currentEpoch) {
			sc.penalizeUserIfNeeded(tx, txHash, newVMInput.CallType, newVMInput.GasProvided, newVMOutput)
		}

		scrForSender, scrForRelayer, errCreateSCR := sc.processSCRForSenderAfterBuiltIn(tx, txHash, vmInput, newVMOutput)
		if errCreateSCR != nil {
			return 0, errCreateSCR
		}

		if !createdAsyncCallback {
			if vmInput.CallType == vmData.AsynchronousCall {
				asyncCallBackSCR := sc.createAsyncCallBackSCRFromVMOutput(newVMOutput, tx, txHash)
				scrResults = append(scrResults, asyncCallBackSCR)
			} else {
				scrResults = append(scrResults, scrForSender)
			}
		}

		if !check.IfNil(scrForRelayer) {
			scrResults = append(scrResults, scrForRelayer)
		}
	}

	if sc.enableEpochsHandler.IsSCRSizeInvariantOnBuiltInResultFlagEnabled() {
		errCheck := sc.checkSCRSizeInvariant(scrResults)
		if errCheck != nil {
			return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, errCheck.Error(), []byte(errCheck.Error()), snapshot, vmInput.GasLocked)
		}
	}

	return sc.finishSCExecution(scrResults, txHash, tx, newVMOutput, builtInFuncGasUsed)
}

func mergeVMOutputLogs(newVMOutput *vmcommon.VMOutput, vmOutput *vmcommon.VMOutput) {
	if vmOutput == nil || len(vmOutput.Logs) == 0 {
		return
	}

	if newVMOutput.Logs == nil {
		newVMOutput.Logs = make([]*vmcommon.LogEntry, 0, len(vmOutput.Logs))
	}

	newVMOutput.Logs = append(vmOutput.Logs, newVMOutput.Logs...)
}

func (sc *scProcessor) processSCRForSenderAfterBuiltIn(
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

	err := sc.addGasRefundIfInShard(scrForSender.RcvAddr, scrForSender.Value)
	if err != nil {
		return nil, nil, err
	}

	if !check.IfNil(scrForRelayer) {
		err = sc.addGasRefundIfInShard(scrForRelayer.RcvAddr, scrForRelayer.Value)
		if err != nil {
			return nil, nil, err
		}
	}

	return scrForSender, scrForRelayer, nil
}

func (sc *scProcessor) resolveBuiltInFunctions(
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {

	vmOutput, err := sc.blockChainHook.ProcessBuiltInFunction(vmInput)
	if err != nil {
		vmOutput = &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: err.Error(),
			GasRemaining:  0,
		}

		if core.IsGetNodeFromDBError(err) {
			return nil, err
		}

		return vmOutput, nil
	}

	return vmOutput, nil
}

func (sc *scProcessor) treatExecutionAfterBuiltInFunc(
	tx data.TransactionHandler,
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
	acntSnd state.UserAccountHandler,
	snapshot int,
) (bool, *vmcommon.VMOutput, *vmcommon.ContractCallInput, error) {
	isSCCall, newVMInput, err := sc.isSCExecutionAfterBuiltInFunc(tx, vmInput, vmOutput)
	if !isSCCall {
		return false, vmOutput, vmInput, nil
	}

	userErrorVmOutput := &vmcommon.VMOutput{
		ReturnCode: vmcommon.UserError,
	}
	if err != nil {
		return true, userErrorVmOutput, newVMInput, sc.ProcessIfError(acntSnd, vmInput.CurrentTxHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	newDestSC, err := sc.getAccountFromAddress(vmInput.RecipientAddr)
	if err != nil {
		return true, userErrorVmOutput, newVMInput, sc.ProcessIfError(acntSnd, vmInput.CurrentTxHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}
	err = sc.checkUpgradePermission(newDestSC, newVMInput)
	if err != nil {
		log.Debug("checkUpgradePermission", "error", err.Error())
		return true, userErrorVmOutput, newVMInput, sc.ProcessIfError(acntSnd, vmInput.CurrentTxHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	newVMOutput, err := sc.executeSmartContractCall(newVMInput, tx, newVMInput.CurrentTxHash, snapshot, acntSnd, newDestSC, vmOutput)
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
) (bool, *vmcommon.ContractCallInput, error) {
	if vmOutput.ReturnCode != vmcommon.Ok {
		return false, nil, nil
	}

	parsedTransfer, err := sc.esdtTransferParser.ParseESDTTransfers(vmInput.CallerAddr, vmInput.RecipientAddr, vmInput.Function, vmInput.Arguments)
	if err != nil {
		return false, nil, nil
	}
	if !core.IsSmartContractAddress(parsedTransfer.RcvAddr) {
		return false, nil, nil
	}

	if sc.shardCoordinator.ComputeId(parsedTransfer.RcvAddr) != sc.shardCoordinator.SelfId() {
		return false, nil, nil
	}

	callType := determineCallType(tx)
	if callType == vmData.AsynchronousCallBack {
		newVMInput := sc.createVMInputWithAsyncCallBackAfterBuiltIn(vmInput, vmOutput, parsedTransfer)
		return true, newVMInput, nil
	}

	if len(parsedTransfer.CallFunction) == 0 {
		return false, nil, nil
	}

	outAcc, ok := vmOutput.OutputAccounts[string(parsedTransfer.RcvAddr)]
	if !ok {
		return false, nil, nil
	}
	if len(outAcc.OutputTransfers) != 1 {
		return false, nil, nil
	}

	scExecuteOutTransfer := outAcc.OutputTransfers[0]
	if !sc.enableEpochsHandler.IsIncrementSCRNonceInMultiTransferFlagEnabled() {
		_, _, err = sc.argsParser.ParseCallData(string(scExecuteOutTransfer.Data))
		if err != nil {
			return true, nil, err
		}
	}

	newVMInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:           vmInput.CallerAddr,
			AsyncArguments:       vmInput.AsyncArguments,
			Arguments:            parsedTransfer.CallArgs,
			CallValue:            big.NewInt(0),
			CallType:             callType,
			GasPrice:             vmInput.GasPrice,
			GasProvided:          scExecuteOutTransfer.GasLimit,
			GasLocked:            vmInput.GasLocked,
			OriginalTxHash:       vmInput.OriginalTxHash,
			CurrentTxHash:        vmInput.CurrentTxHash,
			ReturnCallAfterError: vmInput.ReturnCallAfterError,
		},
		RecipientAddr:     parsedTransfer.RcvAddr,
		Function:          parsedTransfer.CallFunction,
		AllowInitFunction: false,
	}
	newVMInput.ESDTTransfers = parsedTransfer.ESDTTransfers

	return true, newVMInput, nil
}

func (sc *scProcessor) createVMInputWithAsyncCallBackAfterBuiltIn(
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
	parsedTransfer *vmcommon.ParsedESDTTransfers,
) *vmcommon.ContractCallInput {
	arguments := [][]byte{big.NewInt(int64(vmOutput.ReturnCode)).Bytes()}
	gasLimit := vmOutput.GasRemaining

	outAcc, ok := vmOutput.OutputAccounts[string(vmInput.RecipientAddr)]
	if ok && len(outAcc.OutputTransfers) == 1 {
		isDeleteWrongArgAsyncAfterBuiltInFlagEnabled := sc.enableEpochsHandler.IsManagedCryptoAPIsFlagEnabled()
		if isDeleteWrongArgAsyncAfterBuiltInFlagEnabled {
			arguments = [][]byte{}
		}

		gasLimit = outAcc.OutputTransfers[0].GasLimit

		isFixAsyncCallBackArgumentsParserFlagSet := sc.enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabled()
		if isFixAsyncCallBackArgumentsParserFlagSet {
			args, err := sc.argsParser.ParseArguments(string(outAcc.OutputTransfers[0].Data))
			log.LogIfError(err, "function", "createVMInputWithAsyncCallBackAfterBuiltIn.ParseArguments")
			arguments = append(arguments, args...)
		} else {
			function, args, err := sc.argsParser.ParseCallData(string(outAcc.OutputTransfers[0].Data))
			log.LogIfError(err, "function", "createVMInputWithAsyncCallBackAfterBuiltIn.ParseCallData")
			if len(function) > 0 {
				arguments = append(arguments, []byte(function))
			}

			arguments = append(arguments, args...)
		}
	}

	newVMInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:           vmInput.CallerAddr,
			Arguments:            arguments,
			AsyncArguments:       vmInput.AsyncArguments,
			CallValue:            big.NewInt(0),
			CallType:             vmData.AsynchronousCallBack,
			GasPrice:             vmInput.GasPrice,
			GasProvided:          gasLimit,
			GasLocked:            vmInput.GasLocked,
			OriginalTxHash:       vmInput.OriginalTxHash,
			CurrentTxHash:        vmInput.CurrentTxHash,
			ReturnCallAfterError: vmInput.ReturnCallAfterError,
		},
		RecipientAddr:     vmInput.RecipientAddr,
		Function:          "callBack",
		AllowInitFunction: false,
	}
	newVMInput.ESDTTransfers = parsedTransfer.ESDTTransfers

	return newVMInput
}

// isCrossShardESDTTransfer is called when return is created out of the esdt transfer as of failed transaction
func (sc *scProcessor) isCrossShardESDTTransfer(sender []byte, receiver []byte, data []byte) (string, bool) {
	sndShardID := sc.shardCoordinator.ComputeId(sender)
	if sndShardID == sc.shardCoordinator.SelfId() {
		return "", false
	}

	dstShardID := sc.shardCoordinator.ComputeId(receiver)
	if dstShardID == sndShardID {
		return "", false
	}

	function, args, err := sc.argsParser.ParseCallData(string(data))
	if err != nil {
		return "", false
	}

	if len(args) < 2 {
		return "", false
	}

	parsedTransfer, err := sc.esdtTransferParser.ParseESDTTransfers(sender, receiver, function, args)
	if err != nil {
		return "", false
	}

	returnData := ""
	returnData += function + "@"

	if function == core.BuiltInFunctionESDTTransfer {
		returnData += hex.EncodeToString(args[0]) + "@"
		returnData += hex.EncodeToString(args[1])

		return returnData, true
	}

	if function == core.BuiltInFunctionESDTNFTTransfer {
		if len(args) < 4 {
			return "", false
		}

		returnData += hex.EncodeToString(args[0]) + "@"
		returnData += hex.EncodeToString(args[1]) + "@"
		returnData += hex.EncodeToString(args[2]) + "@"
		returnData += hex.EncodeToString(args[3])

		return returnData, true
	}

	if function == core.BuiltInFunctionMultiESDTNFTTransfer {
		numTransferArgs := len(args)
		if len(parsedTransfer.CallFunction) > 0 {
			numTransferArgs -= len(parsedTransfer.CallArgs) + 1
		}
		if numTransferArgs < 4 {
			return "", false
		}

		for i := 0; i < numTransferArgs-1; i++ {
			returnData += hex.EncodeToString(args[i]) + "@"
		}
		returnData += hex.EncodeToString(args[numTransferArgs-1])
		return returnData, true
	}

	return "", false
}

func (sc *scProcessor) getOriginalTxHashIfIntraShardRelayedSCR(
	tx data.TransactionHandler,
	txHash []byte) []byte {
	relayedSCR, isRelayed := isRelayedTx(tx)
	if !isRelayed {
		return txHash
	}

	sndShardID := sc.shardCoordinator.ComputeId(relayedSCR.SndAddr)
	rcvShardID := sc.shardCoordinator.ComputeId(relayedSCR.RcvAddr)
	if sndShardID != rcvShardID {
		return txHash
	}

	return relayedSCR.OriginalTxHash
}

// ProcessIfError creates a smart contract result, consumes the gas and returns the value to the user
func (sc *scProcessor) ProcessIfError(
	acntSnd state.UserAccountHandler,
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
	snapshot int,
	gasLocked uint64,
) error {
	return sc.processIfErrorWithAddedLogs(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked, nil, nil)
}

func (sc *scProcessor) processIfErrorWithAddedLogs(
	acntSnd state.UserAccountHandler,
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
	snapshot int,
	gasLocked uint64,
	prevVmOutput *vmcommon.VMOutput,
	internalVMLogs []*vmcommon.LogEntry,
) error {
	err := sc.accounts.RevertToSnapshot(snapshot)
	if err != nil {
		if !core.IsClosingError(err) {
			log.Warn("revert to snapshot", "error", err.Error())
		}

		return err
	}

	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if len(returnMessage) == 0 && sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		returnMessage = []byte(returnCode)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return err
	}

	sc.setEmptyRoothashOnErrorIfSaveKeyValue(tx, acntSnd)

	scrIfError, consumedFee := sc.createSCRsWhenError(acntSnd, txHash, tx, returnCode, returnMessage, gasLocked)
	err = sc.addBackTxValues(acntSnd, scrIfError, tx)
	if err != nil {
		return err
	}

	userErrorLog := createNewLogFromSCRIfError(scrIfError)

	if !sc.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabled() || !sc.isInformativeTxHandler(scrIfError) {
		err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrIfError})
		if err != nil {
			return err
		}
	}

	relayerLog, err := sc.processForRelayerWhenError(tx, txHash, returnMessage)
	if err != nil {
		return err
	}

	processIfErrorLogs := make([]*vmcommon.LogEntry, 0)
	if prevVmOutput != nil && len(prevVmOutput.Logs) > 0 {
		processIfErrorLogs = append(processIfErrorLogs, prevVmOutput.Logs...)
	}

	processIfErrorLogs = append(processIfErrorLogs, userErrorLog)
	if relayerLog != nil {
		processIfErrorLogs = append(processIfErrorLogs, relayerLog)
	}
	if len(internalVMLogs) > 0 {
		processIfErrorLogs = append(processIfErrorLogs, internalVMLogs...)
	}

	logsTxHash := sc.getOriginalTxHashIfIntraShardRelayedSCR(tx, txHash)
	ignorableError := sc.txLogsProcessor.SaveLog(logsTxHash, tx, processIfErrorLogs)
	if ignorableError != nil {
		log.Debug("scProcessor.ProcessIfError() txLogsProcessor.SaveLog()", "error", ignorableError.Error())
	}

	sc.vmOutputCacher.Put(txHash, &vmcommon.VMOutput{
		ReturnCode:    vmcommon.SimulateFailed,
		ReturnMessage: string(returnMessage),
		Logs:          processIfErrorLogs,
	}, 0)

	txType, _ := sc.txTypeHandler.ComputeTransactionType(tx)
	isCrossShardMoveBalance := txType == process.MoveBalance && check.IfNil(acntSnd)
	if isCrossShardMoveBalance && sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		// move balance was already consumed in sender shard
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)

	if sc.enableEpochsHandler.IsOptimizeNFTStoreFlagEnabled() {
		err = sc.blockChainHook.SaveNFTMetaDataToSystemAccount(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *scProcessor) setEmptyRoothashOnErrorIfSaveKeyValue(tx data.TransactionHandler, account state.UserAccountHandler) {
	if !sc.enableEpochsHandler.IsBackwardCompSaveKeyValueFlagEnabled() {
		return
	}
	if sc.shardCoordinator.SelfId() == core.MetachainShardId {
		return
	}
	if check.IfNil(account) {
		return
	}
	if !bytes.Equal(tx.GetSndAddr(), tx.GetRcvAddr()) {
		return
	}
	if account.GetRootHash() != nil {
		return
	}

	function, args, err := sc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return
	}
	if function != core.BuiltInFunctionSaveKeyValue {
		return
	}
	if len(args) < 3 {
		return
	}

	txGasProvided, err := sc.prepareGasProvided(tx)
	if err != nil {
		return
	}

	lenVal := len(args[1])
	lenKeyVal := len(args[0]) + len(args[1])
	gasToUseForOneSave := sc.builtInGasCosts[function] + sc.persistPerByte*uint64(lenKeyVal) + sc.storePerByte*uint64(lenVal)
	if txGasProvided < gasToUseForOneSave {
		return
	}

	account.SetRootHash(make([]byte, 32))
}

func (sc *scProcessor) processForRelayerWhenError(
	originalTx data.TransactionHandler,
	txHash []byte,
	returnMessage []byte,
) (*vmcommon.LogEntry, error) {
	relayedSCR, isRelayed := isRelayedTx(originalTx)
	if !isRelayed {
		return nil, nil
	}
	if relayedSCR.Value.Cmp(zero) == 0 {
		return nil, nil
	}

	relayerAcnt, err := sc.getAccountFromAddress(relayedSCR.RelayerAddr)
	if err != nil {
		return nil, err
	}

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.AddToBalance(relayedSCR.RelayedValue)
		if err != nil {
			return nil, err
		}

		err = sc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			log.Debug("error saving account")
			return nil, err
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

	if !sc.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabled() || scrForRelayer.Value.Cmp(zero) > 0 {
		err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrForRelayer})
		if err != nil {
			return nil, err
		}
	}

	newLog := createNewLogFromSCRIfError(scrForRelayer)

	return newLog, nil
}

func createNewLogFromSCR(txHandler data.TransactionHandler) *vmcommon.LogEntry {
	returnMessage := make([]byte, 0)
	scr, ok := txHandler.(*smartContractResult.SmartContractResult)
	if ok {
		returnMessage = scr.ReturnMessage
	}

	newLog := &vmcommon.LogEntry{
		Identifier: []byte(core.WriteLogIdentifier),
		Address:    txHandler.GetSndAddr(),
		Topics:     [][]byte{txHandler.GetRcvAddr()},
		Data:       txHandler.GetData(),
	}
	if len(returnMessage) > 0 {
		newLog.Topics = append(newLog.Topics, returnMessage)
	}

	return newLog
}

func createNewLogFromSCRIfError(txHandler data.TransactionHandler) *vmcommon.LogEntry {
	returnMessage := make([]byte, 0)
	scr, ok := txHandler.(*smartContractResult.SmartContractResult)
	if ok {
		returnMessage = scr.ReturnMessage
	}

	newLog := &vmcommon.LogEntry{
		Identifier: []byte(core.SignalErrorOperation),
		Address:    txHandler.GetSndAddr(),
		Topics:     [][]byte{txHandler.GetRcvAddr(), returnMessage},
		Data:       txHandler.GetData(),
	}

	return newLog
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

	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	isOriginalTxAsyncCallBack := sc.enableEpochsHandler.IsSenderInOutTransferFlagEnabledInEpoch(currentEpoch) &&
		determineCallType(originalTx) == vmData.AsynchronousCallBack &&
		sc.shardCoordinator.SelfId() == sc.shardCoordinator.ComputeId(originalTx.GetRcvAddr())
	if isOriginalTxAsyncCallBack {
		destAcc, err := sc.getAccountFromAddress(originalTx.GetRcvAddr())
		if err != nil {
			return err
		}

		err = destAcc.AddToBalance(valueForSnd)
		if err != nil {
			return err
		}

		return sc.accounts.SaveAccount(destAcc)
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

// DeploySmartContract processes the transaction, then deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(tx data.TransactionHandler, acntSnd state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	err := sc.checkTxValidity(tx)
	if err != nil {
		log.Debug("invalid transaction", "error", err.Error())
		return 0, err
	}

	sw := core.NewStopWatch()
	sw.Start("deploy")
	returnCode, err := sc.doDeploySmartContract(tx, acntSnd)
	sw.Stop("deploy")
	duration := sw.GetMeasurement("deploy")

	if duration > executeDurationAlarmThreshold {
		txHash := sc.computeTxHashUnsafe(tx)
		log.Debug(fmt.Sprintf("scProcessor.DeploySmartContract(): execution took > %s", executeDurationAlarmThreshold), "tx hash", txHash, "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
	} else {
		log.Trace("scProcessor.DeploySmartContract()", "sc", tx.GetRcvAddr(), "duration", duration, "returnCode", returnCode, "err", err, "data", string(tx.GetData()))
	}

	return returnCode, err
}

func (sc *scProcessor) doDeploySmartContract(
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	sc.blockChainHook.ResetCounters()
	defer sc.printBlockchainHookCounters(tx)

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		log.Debug("wrong transaction - not empty address", "error", process.ErrWrongTransaction.Error())
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
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	shouldAllowDeploy := sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) || sc.isGenesisProcessing
	if !shouldAllowDeploy {
		log.Trace("deploy is disabled")
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, process.ErrSmartContractDeploymentIsDisabled.Error(), []byte(""), snapshot, 0)
	}

	vmInput, vmType, err := sc.createVMDeployInput(tx)
	if err != nil {
		log.Trace("Transaction data invalid", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, 0)
	}

	sc.wasmVMChangeLocker.RLock()
	vmExec, err := sc.vmContainer.Get(vmType)
	if err != nil {
		sc.wasmVMChangeLocker.RUnlock()
		log.Trace("VM not found", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	vmOutput, err = vmExec.RunSmartContractCreate(vmInput)
	sc.wasmVMChangeLocker.RUnlock()
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		if core.IsGetNodeFromDBError(err) {
			return vmcommon.ExecutionFailed, err
		}
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}

	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Trace("run smart contract create", "error", err.Error())
		return vmcommon.UserError, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(""), snapshot, vmInput.GasLocked)
	}
	vmOutput.GasRemaining += vmInput.GasLocked
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmcommon.UserError, sc.processIfErrorWithAddedLogs(acntSnd, txHash, tx, vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked, nil, vmOutput.Logs)
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte("gas consumed exceeded"), snapshot, vmInput.GasLocked)
	}

	results, err := sc.processVMOutput(&vmInput.VMInput, vmOutput, txHash, tx)
	if err != nil {
		log.Trace("Processing error", "error", err.Error())
		if core.IsGetNodeFromDBError(err) {
			return vmcommon.ExecutionFailed, err
		}
		return vmcommon.ExecutionFailed, sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(vmOutput.ReturnMessage), snapshot, vmInput.GasLocked)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd) // nolint
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return 0, err
	}
	finalResults, logsFromSCRs := sc.cleanInformativeOnlySCRs(results)

	vmOutput.Logs = append(vmOutput.Logs, logsFromSCRs...)
	err = sc.updateDeveloperRewardsProxy(tx, vmOutput, 0)
	if err != nil {
		log.Debug("updateDeveloperRewardsProxy", "error", err.Error())
		return 0, err
	}

	err = sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Debug("AddIntermediate Transaction error", "error", err.Error())
		return 0, err
	}

	totalConsumedFee, totalDevRwd := sc.computeTotalConsumedFeeAndDevRwd(tx, vmOutput, 0)
	sc.txFeeHandler.ProcessTransactionFee(totalConsumedFee, totalDevRwd, txHash)
	sc.printScDeployed(vmOutput, tx)
	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)

	sc.vmOutputCacher.Put(txHash, vmOutput, 0)

	ignorableError := sc.txLogsProcessor.SaveLog(txHash, tx, vmOutput.Logs)
	if ignorableError != nil {
		log.Debug("scProcessor.DeploySmartContract() txLogsProcessor.SaveLog()", "error", ignorableError.Error())
	}

	return 0, nil
}

func (sc *scProcessor) updateDeveloperRewardsProxy(
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) error {
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
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

		scAddress := sc.pubkeyConv.SilentEncode(addr, log)
		scGenerated = append(scGenerated, scAddress)
	}

	ownerAddress := sc.pubkeyConv.SilentEncode(tx.GetSndAddr(), log)

	log.Debug("SmartContract deployed",
		"owner", ownerAddress,
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
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsPenalizedTooMuchGasFlagEnabledInEpoch(currentEpoch) {
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
	vmInput *vmcommon.VMInput,
	vmOutput *vmcommon.VMOutput,
	txHash []byte,
	tx data.TransactionHandler,
) ([]data.TransactionHandler, error) {

	sc.penalizeUserIfNeeded(tx, txHash, vmInput.CallType, vmInput.GasProvided, vmOutput)
	scrForSender, scrForRelayer := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmInput.CallType,
	)

	outPutAccounts := process.SortVMOutputInsideData(vmOutput)
	createdAsyncCallback, scrTxs, err := sc.processSCOutputAccounts(vmInput, vmOutput, outPutAccounts, tx, txHash)
	if err != nil {
		return nil, err
	}

	if !check.IfNil(scrForRelayer) {
		scrTxs = append(scrTxs, scrForRelayer)
		err = sc.addGasRefundIfInShard(scrForRelayer.RcvAddr, scrForRelayer.Value)
		if err != nil {
			return nil, err
		}
	}

	if !createdAsyncCallback && vmInput.CallType == vmData.AsynchronousCall {
		asyncCallBackSCR := sc.createAsyncCallBackSCRFromVMOutput(vmOutput, tx, txHash)
		scrTxs = append(scrTxs, asyncCallBackSCR)
	} else if !createdAsyncCallback {
		scrTxs = append(scrTxs, scrForSender)
	}

	if !createdAsyncCallback {
		err = sc.addGasRefundIfInShard(scrForSender.RcvAddr, scrForSender.Value)
		if err != nil {
			return nil, err
		}
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, err
	}

	err = sc.checkSCRSizeInvariant(scrTxs)
	if err != nil {
		return nil, err
	}

	return scrTxs, nil
}

func (sc *scProcessor) checkSCRSizeInvariant(scrTxs []data.TransactionHandler) error {
	if !sc.enableEpochsHandler.IsSCRSizeInvariantCheckFlagEnabled() {
		return nil
	}

	for _, scrHandler := range scrTxs {
		scr, ok := scrHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		lenTotalData := len(scr.Data) + len(scr.ReturnMessage) + len(scr.Code)
		if lenTotalData > maxTotalSCRsSize {
			return process.ErrResultingSCRIsTooBig
		}
	}

	return nil
}

func (sc *scProcessor) addGasRefundIfInShard(address []byte, value *big.Int) error {
	userAcc, err := sc.getAccountFromAddress(address)
	if err != nil {
		return err
	}

	if check.IfNil(userAcc) {
		return nil
	}

	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) && core.IsSmartContractAddress(address) {
		userAcc.AddToDeveloperReward(value)
	} else {
		err = userAcc.AddToBalance(value)
		if err != nil {
			return err
		}
	}

	return sc.accounts.SaveAccount(userAcc)
}

func (sc *scProcessor) penalizeUserIfNeeded(
	tx data.TransactionHandler,
	txHash []byte,
	callType vmData.CallType,
	gasProvidedForProcessing uint64,
	vmOutput *vmcommon.VMOutput,
) {
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsPenalizedTooMuchGasFlagEnabledInEpoch(currentEpoch) {
		return
	}
	if callType == vmData.AsynchronousCall {
		return
	}

	isTooMuchProvided := isTooMuchGasProvided(gasProvidedForProcessing, vmOutput.GasRemaining)
	if !isTooMuchProvided {
		return
	}

	gasUsed := gasProvidedForProcessing - vmOutput.GasRemaining
	log.Trace("scProcessor.penalizeUserIfNeeded: too much gas provided",
		"hash", txHash,
		"nonce", tx.GetNonce(),
		"value", tx.GetValue(),
		"sender", tx.GetSndAddr(),
		"receiver", tx.GetRcvAddr(),
		"gas limit", tx.GetGasLimit(),
		"gas price", tx.GetGasPrice(),
		"gas provided", gasProvidedForProcessing,
		"gas remained", vmOutput.GasRemaining,
		"gas used", gasUsed,
		"return code", vmOutput.ReturnCode.String(),
		"return message", vmOutput.ReturnMessage,
	)

	if sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		vmOutput.ReturnMessage += "@"
		if !isSmartContractResult(tx) {
			gasUsed += sc.economicsFee.ComputeGasLimit(tx)
		}
	}

	if sc.enableEpochsHandler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() {
		sc.gasHandler.SetGasPenalized(vmOutput.GasRemaining, txHash)
	}

	if !sc.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabled() {
		vmOutput.ReturnMessage += fmt.Sprintf("%s: gas needed = %d, gas remained = %d",
			TooMuchGasProvidedMessage, gasUsed, vmOutput.GasRemaining)
	} else {
		vmOutput.ReturnMessage += fmt.Sprintf("%s for processing: gas provided = %d, gas used = %d",
			TooMuchGasProvidedMessage, gasProvidedForProcessing, gasProvidedForProcessing-vmOutput.GasRemaining)
	}

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
	if callType == vmData.AsynchronousCallBack {
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
	esdtReturnData, isCrossShardESDTCall := sc.isCrossShardESDTTransfer(tx.GetSndAddr(), tx.GetRcvAddr(), tx.GetData())
	if callType != vmData.AsynchronousCallBack && isCrossShardESDTCall {
		accumulatedSCRData += esdtReturnData
	}

	consumedFee := sc.economicsFee.ComputeTxFee(tx)
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsPenalizedTooMuchGasFlagEnabledInEpoch(currentEpoch) {
		consumedFee = core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	if !sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		accumulatedSCRData += "@" + hex.EncodeToString([]byte(returnCode)) + "@" + hex.EncodeToString(txHash)
		if check.IfNil(acntSnd) {
			moveBalanceCost := sc.economicsFee.ComputeMoveBalanceFee(tx)
			consumedFee.Sub(consumedFee, moveBalanceCost)
		}
	} else {
		if callType == vmData.AsynchronousCall {
			scr.CallType = vmData.AsynchronousCallBack
			scr.GasPrice = tx.GetGasPrice()
			if tx.GetGasLimit() >= gasLocked {
				scr.GasLimit = gasLocked
				consumedFee = sc.economicsFee.ComputeFeeForProcessing(tx, tx.GetGasLimit()-gasLocked)
			}

			accumulatedSCRData += "@" + core.ConvertToEvenHex(int(vmcommon.UserError))
			if sc.enableEpochsHandler.IsRepairCallbackFlagEnabledInEpoch(currentEpoch) {
				accumulatedSCRData += "@" + hex.EncodeToString(returnMessage)
			}
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
	transferNonce uint64,
) *smartContractResult.SmartContractResult {
	result := &smartContractResult.SmartContractResult{}

	result.Value = big.NewInt(0)
	result.Nonce = outAcc.Nonce + transferNonce
	result.RcvAddr = outAcc.Address
	result.SndAddr = tx.GetRcvAddr()
	result.Code = outAcc.Code
	result.GasPrice = tx.GetGasPrice()
	result.PrevTxHash = txHash
	result.CallType = vmData.DirectCall
	setOriginalTxHash(result, txHash, tx)

	relayedTx, isRelayed := isRelayedTx(tx)
	if isRelayed {
		result.RelayedValue = big.NewInt(0)
		result.RelayerAddr = relayedTx.RelayerAddr
	}

	return result
}

func (sc *scProcessor) addVMOutputResultsToSCR(vmOutput *vmcommon.VMOutput, result *smartContractResult.SmartContractResult) {
	result.CallType = vmData.AsynchronousCallBack
	result.GasLimit = vmOutput.GasRemaining
	result.Data = []byte("@" + core.ConvertToEvenHex(int(vmOutput.ReturnCode)))

	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if vmOutput.ReturnCode != vmcommon.Ok && sc.enableEpochsHandler.IsRepairCallbackFlagEnabledInEpoch(currentEpoch) {
		encodedReturnMessage := "@" + hex.EncodeToString([]byte(vmOutput.ReturnMessage))
		result.Data = append(result.Data, encodedReturnMessage...)
	}

	addReturnDataToSCR(vmOutput, result)
}

func (sc *scProcessor) createAsyncCallBackSCRFromVMOutput(
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

	sc.addVMOutputResultsToSCR(vmOutput, scr)

	return scr
}

func (sc *scProcessor) createSCRFromStakingSC(
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	if !bytes.Equal(outAcc.Address, vm.StakingSCAddress) {
		return nil
	}

	storageUpdates := process.GetSortedStorageUpdates(outAcc)
	result := createBaseSCR(outAcc, tx, txHash, 0)
	result.Data = append(result.Data, sc.argsParser.CreateDataFromStorageUpdate(storageUpdates)...)
	return result
}

func (sc *scProcessor) createSCRIfNoOutputTransfer(
	vmOutput *vmcommon.VMOutput,
	callType vmData.CallType,
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) (bool, []data.TransactionHandler) {
	if callType == vmData.AsynchronousCall && bytes.Equal(outAcc.Address, tx.GetSndAddr()) {
		result := createBaseSCR(outAcc, tx, txHash, 0)
		sc.addVMOutputResultsToSCR(vmOutput, result)
		return true, []data.TransactionHandler{result}
	}

	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		result := createBaseSCR(outAcc, tx, txHash, 0)
		result.Code = outAcc.Code
		result.Value.Set(outAcc.BalanceDelta)
		if result.Value.Cmp(zero) > 0 {
			result.OriginalSender = tx.GetSndAddr()
		}

		return false, []data.TransactionHandler{result}
	}

	return false, nil
}

func (sc *scProcessor) preprocessOutTransferToSCR(
	index int,
	outputTransfer vmcommon.OutputTransfer,
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	transferNonce := uint64(0)
	if sc.enableEpochsHandler.IsIncrementSCRNonceInMultiTransferFlagEnabled() {
		transferNonce = uint64(index)
	}
	result := createBaseSCR(outAcc, tx, txHash, transferNonce)

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
	return result
}

func (sc *scProcessor) createSmartContractResults(
	vmInput *vmcommon.VMInput,
	vmOutput *vmcommon.VMOutput,
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) (bool, []data.TransactionHandler, error) {

	result := sc.createSCRFromStakingSC(outAcc, tx, txHash)
	if !check.IfNil(result) {
		return false, []data.TransactionHandler{result}, nil
	}

	lenOutTransfers := len(outAcc.OutputTransfers)
	if lenOutTransfers == 0 {
		createdAsyncCallBack, scResults := sc.createSCRIfNoOutputTransfer(vmOutput, vmInput.CallType, outAcc, tx, txHash)
		return createdAsyncCallBack, scResults, nil
	}

	createdAsyncCallBack := false
	scResults := make([]data.TransactionHandler, 0, len(outAcc.OutputTransfers))
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	for i, outputTransfer := range outAcc.OutputTransfers {
		result = sc.preprocessOutTransferToSCR(i, outputTransfer, outAcc, tx, txHash)

		isCrossShard := sc.shardCoordinator.ComputeId(outAcc.Address) != sc.shardCoordinator.SelfId()

		if isCrossShard && result.CallType == vmData.AsynchronousCall &&
			sc.enableRoundsHandler.IsDisableAsyncCallV1Enabled() {
			return false, nil, process.ErrAsyncCallsDisabled
		}

		if result.CallType == vmData.AsynchronousCallBack {
			isCreatedCallBackCrossShardOnlyFlagSet := sc.enableEpochsHandler.IsMultiESDTTransferFixOnCallBackFlagEnabled()
			if !isCreatedCallBackCrossShardOnlyFlagSet || isCrossShard {
				// backward compatibility
				createdAsyncCallBack = true
				result.GasLimit, _ = core.SafeAddUint64(result.GasLimit, vmOutput.GasRemaining)
			}
		}

		useSenderAddressFromOutTransfer := sc.enableEpochsHandler.IsSenderInOutTransferFlagEnabledInEpoch(currentEpoch) &&
			len(outputTransfer.SenderAddress) == len(tx.GetSndAddr()) &&
			sc.shardCoordinator.ComputeId(outputTransfer.SenderAddress) == sc.shardCoordinator.SelfId()
		if useSenderAddressFromOutTransfer {
			result.SndAddr = outputTransfer.SenderAddress
		}

		isOutTransferTxRcvAddr := bytes.Equal(result.SndAddr, tx.GetRcvAddr())
		outputTransferCopy := outputTransfer
		isLastOutTransfer := i == lenOutTransfers-1
		if !createdAsyncCallBack && isLastOutTransfer && isOutTransferTxRcvAddr &&
			sc.useLastTransferAsAsyncCallBackWhenNeeded(vmInput, outAcc, &outputTransferCopy, vmOutput, tx, result, isCrossShard) {
			createdAsyncCallBack = true
		}

		if result.CallType == vmData.AsynchronousCall {
			isCreatedCallBackCrossShardOnlyFlagSet := sc.enableEpochsHandler.IsMultiESDTTransferFixOnCallBackFlagEnabled()
			if !isCreatedCallBackCrossShardOnlyFlagSet || isCrossShard {
				result.GasLimit += outputTransfer.GasLocked
				lastArgAsGasLocked := "@" + hex.EncodeToString(big.NewInt(0).SetUint64(outputTransfer.GasLocked).Bytes())
				result.Data = append(result.Data, []byte(lastArgAsGasLocked)...)
			}
		}

		scResults = append(scResults, result)
	}

	return createdAsyncCallBack, scResults, nil
}

func (sc *scProcessor) useLastTransferAsAsyncCallBackWhenNeeded(
	vmInput *vmcommon.VMInput,
	outAcc *vmcommon.OutputAccount,
	outputTransfer *vmcommon.OutputTransfer,
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	result *smartContractResult.SmartContractResult,
	isCrossShard bool,
) bool {
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if len(vmOutput.ReturnData) > 0 && !sc.enableEpochsHandler.IsReturnDataToLastTransferFlagEnabledAfterEpoch(currentEpoch) {
		return false
	}

	isAsyncTransferBackToSender := vmInput.CallType == vmData.AsynchronousCall &&
		bytes.Equal(outAcc.Address, tx.GetSndAddr())
	if !isAsyncTransferBackToSender {
		return false
	}

	isCreatedCallBackCrossShardOnlyFlagSet := sc.enableEpochsHandler.IsMultiESDTTransferFixOnCallBackFlagEnabled()
	if isCreatedCallBackCrossShardOnlyFlagSet && !isCrossShard {
		return false
	}

	if !sc.isTransferWithNoAdditionalData(outputTransfer.SenderAddress, outAcc.Address, outputTransfer.Data) {
		return false
	}

	if sc.enableEpochsHandler.IsFixAsyncCallBackArgsListFlagEnabled() {
		result.Data = append(result.Data, []byte("@"+core.ConvertToEvenHex(int(vmOutput.ReturnCode)))...)
	}

	addReturnDataToSCR(vmOutput, result)
	result.CallType = vmData.AsynchronousCallBack
	result.GasLimit, _ = core.SafeAddUint64(result.GasLimit, vmOutput.GasRemaining)

	return true
}

func (sc *scProcessor) getESDTParsedTransfers(sndAddr []byte, dstAddr []byte, data []byte,
) (*vmcommon.ParsedESDTTransfers, error) {
	function, args, err := sc.argsParser.ParseCallData(string(data))
	if err != nil {
		return nil, err
	}

	parsedTransfer, err := sc.esdtTransferParser.ParseESDTTransfers(sndAddr, dstAddr, function, args)
	if err != nil {
		return nil, err
	}

	return parsedTransfer, nil
}

func (sc *scProcessor) isTransferWithNoAdditionalData(sndAddr []byte, dstAddr []byte, data []byte) bool {
	if len(data) == 0 {
		return true
	}

	parsedTransfer, err := sc.getESDTParsedTransfers(sndAddr, dstAddr, data)
	if err != nil {
		return false
	}

	return len(parsedTransfer.CallFunction) == 0
}

// createSCRForSender(vmOutput, tx, txHash, acntSnd)
// give back the user the unused gas money
func (sc *scProcessor) createSCRForSenderAndRelayer(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte,
	callType vmData.CallType,
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
	currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
	if !sc.enableEpochsHandler.IsSCDeployFlagEnabledInEpoch(currentEpoch) {
		storageFreeRefund = big.NewInt(0).Mul(vmOutput.GasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
		gasRemaining = vmOutput.GasRemaining
	}

	rcvAddress := tx.GetSndAddr()
	if callType == vmData.AsynchronousCallBack {
		rcvAddress = tx.GetRcvAddr()
	}

	var refundGasToRelayerSCR *smartContractResult.SmartContractResult
	relayedSCR, isRelayed := isRelayedTx(tx)
	shouldRefundGasToRelayerSCR := isRelayed && callType != vmData.AsynchronousCall && gasRefund.Cmp(zero) > 0
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
			CallType:       vmData.DirectCall,
			ReturnMessage:  []byte(core.GasRefundForRelayerMessage),
			OriginalSender: relayedSCR.OriginalSender,
		}
		gasRemaining = 0
	}

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = big.NewInt(0).Set(storageFreeRefund)
	if callType != vmData.AsynchronousCall && check.IfNil(refundGasToRelayerSCR) {
		scTx.Value.Add(scTx.Value, gasRefund)
	}

	scTx.RcvAddr = rcvAddress
	scTx.SndAddr = tx.GetRcvAddr()
	scTx.Nonce = tx.GetNonce() + 1
	scTx.PrevTxHash = txHash
	scTx.GasLimit = gasRemaining
	scTx.GasPrice = tx.GetGasPrice()
	scTx.ReturnMessage = []byte(vmOutput.ReturnMessage)
	scTx.CallType = vmData.DirectCall
	setOriginalTxHash(scTx, txHash, tx)
	scTx.Data = []byte("@" + hex.EncodeToString([]byte(vmOutput.ReturnCode.String())))
	isDeleteWrongArgAsyncAfterBuiltInFlagEnabled := sc.enableEpochsHandler.IsManagedCryptoAPIsFlagEnabled()
	if isDeleteWrongArgAsyncAfterBuiltInFlagEnabled && callType == vmData.AsynchronousCall {
		scTx.Data = []byte("@" + core.ConvertToEvenHex(int(vmOutput.ReturnCode)))
	}

	// when asynchronous call - the callback is created by combining the last output transfer with the returnData
	if callType != vmData.AsynchronousCall {
		addReturnDataToSCR(vmOutput, scTx)
	}

	log.Trace("createSCRForSenderAndRelayer ", "data", string(scTx.Data), "snd", scTx.SndAddr, "rcv", scTx.RcvAddr, "gasRemaining", vmOutput.GasRemaining)
	return scTx, refundGasToRelayerSCR
}

func addReturnDataToSCR(vmOutput *vmcommon.VMOutput, scTx *smartContractResult.SmartContractResult) {
	for _, retData := range vmOutput.ReturnData {
		scTx.Data = append(scTx.Data, []byte("@"+hex.EncodeToString(retData))...)
	}
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(
	vmInput *vmcommon.VMInput,
	vmOutput *vmcommon.VMOutput,
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

		tmpCreatedAsyncCallback, newScrs, err := sc.createSmartContractResults(vmInput, vmOutput, outAcc, tx, txHash)
		if err != nil {
			return false, nil, err
		}
		createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback

		if len(newScrs) != 0 {
			scResults = append(scResults, newScrs...)
		}
		if check.IfNil(acc) {
			if outAcc.BalanceDelta != nil {
				if outAcc.BalanceDelta.Cmp(zero) < 0 {
					return false, nil, process.ErrNegativeBalanceDeltaOnCrossShardAccount
				}
				sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
			}
			continue
		}

		for _, storeUpdate := range outAcc.StorageUpdates {
			if !process.IsAllowedToSaveUnderKey(storeUpdate.Offset) {
				log.Trace("storeUpdate is not allowed", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
				isSaveKeyValueUnderProtectedErrorFlagSet := sc.enableEpochsHandler.IsRemoveNonUpdatedStorageFlagEnabled()
				if isSaveKeyValueUnderProtectedErrorFlagSet {
					return false, nil, process.ErrNotAllowedToWriteUnderProtectedKey
				}

				continue
			}

			err = acc.SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				log.Warn("saveKeyValue", "error", err)
				return false, nil, err
			}
			log.Trace("storeUpdate", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
		}

		err = sc.updateSmartContractCode(vmOutput, acc, outAcc)
		if err != nil {
			return false, nil, err
		}

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
	vmOutput *vmcommon.VMOutput,
	stateAccount state.UserAccountHandler,
	outputAccount *vmcommon.OutputAccount,
) error {
	if len(outputAccount.Code) == 0 {
		return nil
	}
	if len(outputAccount.CodeMetadata) == 0 {
		return nil
	}
	if !core.IsSmartContractAddress(outputAccount.Address) {
		return nil
	}

	outputAccountCodeMetadataBytes, err := sc.blockChainHook.FilterCodeMetadataForUpgrade(outputAccount.CodeMetadata)
	if err != nil {
		return err
	}

	// This check is desirable (not required though) since currently both Wasm VM and IELE send the code in the output account even for "regular" execution
	sameCode := bytes.Equal(outputAccount.Code, sc.accounts.GetCode(stateAccount.GetCodeHash()))
	sameCodeMetadata := bytes.Equal(outputAccountCodeMetadataBytes, stateAccount.GetCodeMetadata())
	if sameCode && sameCodeMetadata {
		return nil
	}

	currentOwner := stateAccount.GetOwnerAddress()
	isCodeDeployerSet := len(outputAccount.CodeDeployerAddress) > 0
	isCodeDeployerOwner := bytes.Equal(currentOwner, outputAccount.CodeDeployerAddress) && isCodeDeployerSet

	noExistingCode := len(sc.accounts.GetCode(stateAccount.GetCodeHash())) == 0
	noExistingOwner := len(currentOwner) == 0
	currentCodeMetadata := vmcommon.CodeMetadataFromBytes(stateAccount.GetCodeMetadata())
	newCodeMetadata := vmcommon.CodeMetadataFromBytes(outputAccountCodeMetadataBytes)
	isUpgradeable := currentCodeMetadata.Upgradeable
	isDeployment := noExistingCode && noExistingOwner
	isUpgrade := !isDeployment && isCodeDeployerOwner && isUpgradeable

	entry := &vmcommon.LogEntry{
		Address: stateAccount.AddressBytes(),
		Topics: [][]byte{
			outputAccount.Address, outputAccount.CodeDeployerAddress,
		},
	}

	encodedOutputAccountAddr := sc.pubkeyConv.SilentEncode(outputAccount.Address, log)

	if isDeployment {
		// At this point, we are under the condition "noExistingOwner"
		stateAccount.SetOwnerAddress(outputAccount.CodeDeployerAddress)
		stateAccount.SetCodeMetadata(outputAccountCodeMetadataBytes)
		stateAccount.SetCode(outputAccount.Code)
		log.Debug("updateSmartContractCode(): created", "address", encodedOutputAccountAddr, "upgradeable", newCodeMetadata.Upgradeable)

		entry.Identifier = []byte(core.SCDeployIdentifier)
		vmOutput.Logs = append(vmOutput.Logs, entry)
		return nil
	}

	if isUpgrade {
		stateAccount.SetCodeMetadata(outputAccountCodeMetadataBytes)
		stateAccount.SetCode(outputAccount.Code)
		log.Debug("updateSmartContractCode(): upgraded", "address", encodedOutputAccountAddr, "upgradeable", newCodeMetadata.Upgradeable)

		entry.Identifier = []byte(core.SCUpgradeIdentifier)
		vmOutput.Logs = append(vmOutput.Logs, entry)
		return nil
	}

	return nil
}

// delete accounts - only suicide by current SC or another SC called by current SC - protected by VM
func (sc *scProcessor) deleteAccounts(deletedAccounts [][]byte) error {
	for _, value := range deletedAccounts {
		acc, err := sc.getAccountFromAddress(value)
		if err != nil {
			return err
		}

		if check.IfNil(acc) {
			// TODO: sharded Smart Contract processing
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
		txHash,
		sc.pubkeyConv,
	)

	gasLocked := sc.getGasLockedFromSCR(scr)

	txType, _ := sc.txTypeHandler.ComputeTransactionType(scr)
	switch txType {
	case process.MoveBalance:
		err = sc.processSimpleSCR(scr, txHash, dstAcc)
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
		if sc.shardCoordinator.SelfId() == core.MetachainShardId && !sc.enableEpochsHandler.IsBuiltInFunctionOnMetaFlagEnabled() {
			returnCode, err = sc.ExecuteSmartContractTransaction(scr, sndAcc, dstAcc)
			return returnCode, err
		}
		returnCode, err = sc.ExecuteBuiltInFunction(scr, sndAcc, dstAcc)
		return returnCode, err
	}

	err = process.ErrWrongTransaction
	return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot, gasLocked)
}

func (sc *scProcessor) getGasLockedFromSCR(scr *smartContractResult.SmartContractResult) uint64 {
	if scr.CallType != vmData.AsynchronousCall {
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
	txHash []byte,
	dstAcc state.UserAccountHandler,
) error {
	if scResult.Value.Cmp(zero) <= 0 {
		return nil
	}

	isPayable, err := sc.IsPayable(scResult.SndAddr, scResult.RcvAddr)
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

	err = sc.accounts.SaveAccount(dstAcc)
	if err != nil {
		return err
	}

	if isReturnOKTxHandler(scResult) {
		completedTxLog := createCompleteEventLog(scResult, txHash)
		ignorableError := sc.txLogsProcessor.SaveLog(txHash, scResult, []*vmcommon.LogEntry{completedTxLog})
		if ignorableError != nil {
			log.Debug("scProcessor.finishSCExecution txLogsProcessor.SaveLog()", "error", ignorableError.Error())
		}
	}

	return nil
}

// CheckBuiltinFunctionIsExecutable validates the builtin function arguments and tx fields without executing it
func (sc *scProcessor) CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction string, tx data.TransactionHandler) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	functionName, arguments, err := sc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return err
	}

	if expectedBuiltinFunction != functionName {
		return process.ErrBuiltinFunctionMismatch
	}

	gasProvided, err := sc.prepareGasProvided(tx)
	if err != nil {
		return err
	}

	sc.mutExecutableCheckers.RLock()
	executableChecker, ok := sc.executableCheckers[functionName]
	sc.mutExecutableCheckers.RUnlock()
	if !ok {
		return process.ErrBuiltinFunctionNotExecutable
	}

	// check if the function is executable
	return executableChecker.CheckIsExecutable(
		tx.GetSndAddr(),
		tx.GetValue(),
		tx.GetRcvAddr(),
		gasProvided,
		arguments,
	)
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

func (sc *scProcessor) createCompleteEventLogIfNoMoreAction(
	tx data.TransactionHandler,
	txHash []byte,
	results []data.TransactionHandler,
) *vmcommon.LogEntry {
	sndShardID := sc.shardCoordinator.ComputeId(tx.GetSndAddr())
	dstShardID := sc.shardCoordinator.ComputeId(tx.GetRcvAddr())
	isCrossShardTxWithExecAtSender := sc.shardCoordinator.SelfId() == sndShardID && sndShardID != dstShardID
	if isCrossShardTxWithExecAtSender && !sc.isInformativeTxHandler(tx) {
		return nil
	}

	for _, scr := range results {
		sndShardID = sc.shardCoordinator.ComputeId(scr.GetSndAddr())
		dstShardID = sc.shardCoordinator.ComputeId(scr.GetRcvAddr())
		isCrossShard := sndShardID != dstShardID
		if isCrossShard && !sc.isInformativeTxHandler(scr) {
			return nil
		}
	}

	return createCompleteEventLog(tx, txHash)
}

func createCompleteEventLog(tx data.TransactionHandler, txHash []byte) *vmcommon.LogEntry {
	prevTxHash := txHash
	originalSCR, ok := tx.(*smartContractResult.SmartContractResult)
	if ok {
		prevTxHash = originalSCR.PrevTxHash
	}

	newLog := &vmcommon.LogEntry{
		Identifier: []byte(core.CompletedTxEventIdentifier),
		Address:    tx.GetRcvAddr(),
		Topics:     [][]byte{prevTxHash},
	}

	return newLog
}

func isReturnOKTxHandler(
	resultTx data.TransactionHandler,
) bool {
	return bytes.HasPrefix(resultTx.GetData(), []byte(returnOkData))
}

// this function should only be called for logging reasons, since it does not perform sanity checks
func (sc *scProcessor) computeTxHashUnsafe(tx data.TransactionHandler) []byte {
	txHash, _ := core.CalculateHash(sc.marshalizer, sc.hasher, tx)

	return txHash
}

// IsPayable returns if address is payable, smart contract ca set to false
func (sc *scProcessor) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	return sc.blockChainHook.IsPayable(sndAddress, recvAddress)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	return sc == nil
}
