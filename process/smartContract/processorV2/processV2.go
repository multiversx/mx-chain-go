package processorV2

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
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/multiversx/mx-chain-vm-go/vmhost/contexts"
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

	generalSCRIdentifier = "writeLog"
	signalError          = "signalError"
	completedTxEvent     = "completedTxEvent"
	returnOkData         = "@6f6b"
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
	arwenChangeLocker  common.Locker

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

type sameShardExecutionDataAfterBuiltIn struct {
	isSCCallSelfShard    bool
	callType             vmData.CallType
	parsedTransfer       *vmcommon.ParsedESDTTransfers
	scExecuteOutTransfer *vmcommon.OutputTransfer
}

type outputResultsToBeMerged struct {
	tmpCreatedAsyncCallback bool
	createdAsyncCallback    bool
	newSCRTxs               []data.TransactionHandler
	scrResults              []data.TransactionHandler
	newVMOutput             *vmcommon.VMOutput
	vmOutput                *vmcommon.VMOutput
}

type internalIndexedScr struct {
	result data.TransactionHandler
	index  uint32
}

// NewSmartContractProcessorV2 creates a smart contract processor that creates and interprets VM data
func NewSmartContractProcessorV2(args scrCommon.ArgsNewSmartContractProcessor) (*scProcessor, error) {
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
		badTxForwarder:      args.BadTxForwarder,
		builtInFunctions:    args.BuiltInFunctions,
		isGenesisProcessing: args.IsGenesisProcessing,
		arwenChangeLocker:   args.WasmVMChangeLocker,
		vmOutputCacher:      args.VMOutputCacher,
		enableEpochsHandler: args.EnableEpochsHandler,
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
		return vmcommon.Ok, process.ErrNilTransaction
	}

	sw := core.NewStopWatch()
	sw.Start("execute")

	failureContext := NewFailureContext()
	returnCode, err := sc.doExecuteSmartContractTransaction(tx, acntSnd, acntDst, failureContext)
	if failureContext.processFail {
		failureProcessingError := sc.processIfErrorWithAddedLogs(
			acntSnd,
			tx,
			failureContext)
		if failureProcessingError != nil {
			err = failureProcessingError
		}
	}

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
	failureContext *failureContext,
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

	failureContext.makeSnaphot(sc.accounts.JournalLen())
	failureContext.setTxHash(txHash)

	var vmInput *vmcommon.ContractCallInput
	vmInput, err = sc.createVMCallInput(tx, txHash, builtInFuncCall)
	if err != nil {
		returnMessage := "cannot create VMInput, check the transaction data field"
		log.Debug("create vm call input error", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(returnMessage))
		return vmcommon.UserError, vmInput, txHash, nil
	}

	err = sc.checkUpgradePermission(acntDst, vmInput)
	if err != nil {
		log.Debug("checkUpgradePermission", "error", err.Error())
		failureContext.setMessagesFromError(err)
		return vmcommon.UserError, vmInput, txHash, nil
	}

	return vmcommon.Ok, vmInput, txHash, nil
}

func (sc *scProcessor) doExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	failureContext *failureContext,
) (vmcommon.ReturnCode, error) {
	returnCode, vmInput, txHash, err := sc.prepareExecution(tx, acntSnd, acntDst, false, failureContext)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}

	failureContext.makeSnaphot(sc.accounts.JournalLen())
	failureContext.setTxHash(txHash)
	failureContext.setGasLocked(vmInput.GasLocked)

	vmOutput, errReturnCode, err := sc.executeSmartContractCallAndCheckGas(vmInput, tx, txHash, acntSnd, acntDst, failureContext)
	if errReturnCode != vmcommon.Ok || err != nil {
		return errReturnCode, err
	}

	var results []data.TransactionHandler
	results, err = sc.processVMOutput(&vmInput.VMInput, vmOutput, txHash, tx, acntSnd, failureContext)
	if err != nil {
		log.Trace("process vm output returned with problem ", "err", err.Error())
		failureContext.setMessages(err.Error(), []byte(vmOutput.ReturnMessage))
		return vmcommon.ExecutionFailed, nil
	}

	return sc.finishSCExecution(results, txHash, tx, vmOutput, 0)
}

func (sc *scProcessor) executeSmartContractCallAndCheckGas(
	vmInput *vmcommon.ContractCallInput,
	tx data.TransactionHandler,
	txHash []byte,
	acntSnd, acntDst state.UserAccountHandler,
	failureContext *failureContext,
) (*vmcommon.VMOutput, vmcommon.ReturnCode, error) {
	vmOutput, err := sc.executeSmartContractCall(vmInput, tx, vmInput.CurrentTxHash, acntSnd, acntDst, failureContext)
	if err != nil {
		return vmOutput, vmcommon.Ok, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput, vmcommon.UserError, nil
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		failureContext.setMessages(err.Error(), []byte("gas consumed exceeded"))
		return vmOutput, vmcommon.ExecutionFailed, nil
	}

	return vmOutput, vmcommon.Ok, nil
}

func (sc *scProcessor) executeSmartContractCall(
	vmInput *vmcommon.ContractCallInput,
	tx data.TransactionHandler,
	_ []byte,
	acntSnd, acntDst state.UserAccountHandler,
	failureContext *failureContext,
) (*vmcommon.VMOutput, error) {
	if check.IfNil(acntDst) {
		return nil, process.ErrNilSCDestAccount
	}

	sc.arwenChangeLocker.RLock()

	userErrorVmOutput := &vmcommon.VMOutput{
		ReturnCode: vmcommon.UserError,
	}
	vmExec, _, err := scrCommon.FindVMByScAddress(sc.vmContainer, vmInput.RecipientAddr)
	if err != nil {
		sc.arwenChangeLocker.RUnlock()
		returnMessage := "cannot get vm from address"
		log.Trace("get vm from address error", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(returnMessage))
		return userErrorVmOutput, nil
	}

	sc.blockChainHook.ResetCounters()
	defer sc.printBlockchainHookCounters(tx)

	var vmOutput *vmcommon.VMOutput
	vmOutput, err = vmExec.RunSmartContractCall(vmInput)

	sc.arwenChangeLocker.RUnlock()
	if err != nil {
		log.Debug("run smart contract call error", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(""))
		return userErrorVmOutput, nil
	}
	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Debug("run smart contract call error", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(""))
		return userErrorVmOutput, nil
	}

	if sc.isMultiLevelAsync(vmInput.CallType, vmOutput) {
		return nil, vmhost.ErrAsyncNoMultiLevel
	}

	vmOutput.GasRemaining += vmInput.GasLocked

	if vmOutput.ReturnCode != vmcommon.Ok {
		failureContext.setMessages(vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage))
		failureContext.setLogs(vmOutput.Logs)
		return userErrorVmOutput, nil
	}
	acntSnd, err = sc.reloadLocalAccount(acntSnd) // nolint
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return nil, err
	}

	return vmOutput, nil
}

func (sc *scProcessor) isMultiLevelAsync(callType vmData.CallType, vmOutput *vmcommon.VMOutput) bool {
	if callType != vmData.AsynchronousCall && callType != vmData.AsynchronousCallBack {
		return false
	}
	for _, outputAcc := range vmOutput.OutputAccounts {
		for _, outTransfer := range outputAcc.OutputTransfers {
			if outTransfer.CallType == vmData.AsynchronousCall {
				return true
			}
		}
	}
	return false
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

	err := sc.updateDeveloperRewards(tx, vmOutput, builtInFuncGasUsed)
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

func (sc *scProcessor) updateDeveloperRewards(
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
			err = sc.addToDevRewards(outAcc.Address, outAcc.GasUsed, tx)
			if err != nil {
				return err
			}
		}
	}

	moveBalanceGasLimit := sc.economicsFee.ComputeGasLimit(tx)
	if !isSmartContractResult(tx) {
		usedGasByMainSC, err = core.SafeSubUint64(usedGasByMainSC, moveBalanceGasLimit)
		if err != nil {
			return err
		}
	}

	err = sc.addToDevRewards(tx.GetRcvAddr(), usedGasByMainSC, tx)
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) addToDevRewards(address []byte, gasUsed uint64, tx data.TransactionHandler) error {
	if core.IsEmptyAddress(address) || !core.IsSmartContractAddress(address) {
		return nil
	}

	consumedFee := sc.economicsFee.ComputeFeeForProcessing(tx, gasUsed)
	devRwd := core.GetIntTrimmedPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())
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
	totalDevRwd := core.GetIntTrimmedPercentageOfValue(totalFeeMinusBuiltIn, sc.economicsFee.DeveloperPercentage())

	if !isSmartContractResult(tx) && senderInSelfShard {
		totalFee.Add(totalFee, sc.economicsFee.ComputeMoveBalanceFee(tx))
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
	_ state.UserAccountHandler,
	tx data.TransactionHandler,
	_ []byte,
) error {
	if _, ok := tx.(*transaction.Transaction); ok {
		err := sc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx})
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

	if isCrossShard {
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

	failureContext := NewFailureContext()
	retCode, err := sc.doExecuteBuiltInFunctionWithoutFailureProcessing(tx, acntSnd, acntDst, failureContext)
	if failureContext.processFail {
		failureProcessingError := sc.processIfErrorWithAddedLogs(
			acntSnd,
			tx,
			failureContext)
		if failureProcessingError != nil {
			err = failureProcessingError
		}
	}
	return retCode, err
}

func (sc *scProcessor) doExecuteBuiltInFunctionWithoutFailureProcessing(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	failureContext *failureContext,
) (vmcommon.ReturnCode, error) {
	returnCode, vmInput, txHash, err := sc.prepareExecution(tx, acntSnd, acntDst, true, failureContext)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}

	failureContext.makeSnaphot(sc.accounts.JournalLen())
	failureContext.setTxHash(txHash)
	failureContext.setGasLocked(vmInput.GasLocked)

	var vmOutput *vmcommon.VMOutput
	vmOutput, err = sc.resolveBuiltInFunctions(vmInput)
	if err != nil {
		log.Debug("processed built in functions error", "error", err.Error())
		return vmcommon.Ok, err
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return vmcommon.Ok, err
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
			err := sc.resolveFailedTransaction(acntSnd, tx, txHash)
			failureContext.setMessage(vmOutput.ReturnMessage)
			failureContext.setGasLocked(0)
			return vmcommon.UserError, err

		}
		failureContext.setMessages(vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage))
		return vmcommon.UserError, nil
	}

	if vmInput.CallType == vmData.AsynchronousCallBack {
		// in case of asynchronous callback - the process of built in function is a must
		failureContext.makeSnaphot(sc.accounts.JournalLen())
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		failureContext.setMessages(err.Error(), []byte("gas consumed exceeded"))
		return vmcommon.ExecutionFailed, nil
	}

	createdAsyncCallback, scrResults, err := sc.processSCOutputAccounts(&vmInput.VMInput, vmOutput, tx, txHash)
	if err != nil {
		failureContext.setMessagesFromError(err)
		return vmcommon.ExecutionFailed, nil
	}

	executionDataAfterBuiltIn, err := sc.isSameShardSCExecutionAfterBuiltInFunc(tx, vmInput, vmOutput)
	if err != nil {
		failureContext.setMessagesFromError(err)
		return vmcommon.ExecutionFailed, nil
	}

	newVMInput := vmInput
	newVMOutput := vmOutput

	var errReturnCode vmcommon.ReturnCode
	if executionDataAfterBuiltIn.isSCCallSelfShard {
		var newDestSC state.UserAccountHandler
		newVMInput, newDestSC, err = sc.prepareExecutionAfterBuiltInFunc(
			&inputDataAfterBuiltInCall{
				callType:             executionDataAfterBuiltIn.callType,
				vmInput:              vmInput,
				vmOutput:             vmOutput,
				parsedTransfer:       executionDataAfterBuiltIn.parsedTransfer,
				scExecuteOutTransfer: executionDataAfterBuiltIn.scExecuteOutTransfer,
				tx:                   tx,
				acntSnd:              acntSnd,
				failureContext:       failureContext,
			})
		if err != nil {
			failureContext.setMessagesFromError(err)
			return vmcommon.ExecutionFailed, nil
		}

		newVMOutput, errReturnCode, err = sc.executeSmartContractCallAndCheckGas(newVMInput, tx, newVMInput.CurrentTxHash, acntSnd, newDestSC, failureContext)
		if err != nil {
			failureContext.setMessagesFromError(err)
			return vmcommon.ExecutionFailed, nil
		}
		if errReturnCode != vmcommon.Ok {
			return errReturnCode, nil // process if error already happened inside executeSmartContractCallAndCheckGas
		}

		tmpCreatedAsyncCallback, newSCRTxs, err := sc.processSCOutputAccounts(&vmInput.VMInput, newVMOutput, tx, txHash)
		if err != nil {
			failureContext.setMessagesFromError(err)
			return vmcommon.ExecutionFailed, nil
		}

		createdAsyncCallback, scrResults = mergeOutputResultsWithBuiltinResults(
			&outputResultsToBeMerged{
				tmpCreatedAsyncCallback: tmpCreatedAsyncCallback,
				createdAsyncCallback:    createdAsyncCallback,
				newSCRTxs:               newSCRTxs,
				scrResults:              scrResults,
				newVMOutput:             newVMOutput,
				vmOutput:                vmOutput,
			})
	}

	isSCCallCrossShard := !executionDataAfterBuiltIn.isSCCallSelfShard && txTypeOnDst == process.SCInvoking
	if !isSCCallCrossShard /* isSCCallSelfShard || txTypeOnDst != process.SCInvoking */ {
		scrResults, errReturnCode, err = sc.completeOutputProcessingAndCreateCallback(
			&outputDataFromCall{
				vmInput:              &vmInput.VMInput,
				vmOutput:             newVMOutput,
				txHash:               txHash,
				tx:                   tx,
				scrTxs:               scrResults,
				acntSnd:              acntSnd,
				createdAsyncCallback: createdAsyncCallback,
				failureContext:       failureContext,
			})
		if errReturnCode != vmcommon.Ok || err != nil {
			return errReturnCode, nil
		}
	}

	err = sc.checkSCRSizeInvariant(scrResults)
	if err != nil {
		failureContext.setMessagesFromError(err)
		return vmcommon.UserError, nil
	}

	return sc.finishSCExecution(scrResults, txHash, tx, newVMOutput, builtInFuncGasUsed)
}

func mergeOutputResultsWithBuiltinResults(results *outputResultsToBeMerged) (bool, []data.TransactionHandler) {
	results.createdAsyncCallback = results.createdAsyncCallback || results.tmpCreatedAsyncCallback

	if len(results.newSCRTxs) > 0 {
		results.scrResults = append(results.scrResults, results.newSCRTxs...)
	}

	mergeVMOutputLogs(results.newVMOutput, results.vmOutput)
	return results.createdAsyncCallback, results.scrResults
}

type inputDataForCallbackSCR struct {
	vmInput       *vmcommon.VMInput
	vmOutput      *vmcommon.VMOutput
	tx            data.TransactionHandler
	txHash        []byte
	scrForSender  *smartContractResult.SmartContractResult
	scrForRelayer *smartContractResult.SmartContractResult
	scrResults    []data.TransactionHandler
}

func (sc *scProcessor) createAsyncCallBackSCR(inputForCallback *inputDataForCallbackSCR) ([]data.TransactionHandler, error) {
	if inputForCallback.vmInput.CallType == vmData.AsynchronousCall {
		asyncCallBackSCR, err := sc.createAsyncCallBackSCRFromVMOutput(
			inputForCallback.vmOutput,
			inputForCallback.tx,
			inputForCallback.txHash,
			inputForCallback.vmInput.AsyncArguments)

		if err != nil {
			return nil, err
		}

		inputForCallback.scrResults = append(inputForCallback.scrResults, asyncCallBackSCR)
	} else {
		inputForCallback.scrResults = append(inputForCallback.scrResults, inputForCallback.scrForSender)
	}

	if !check.IfNil(inputForCallback.scrForRelayer) {
		inputForCallback.scrResults = append(inputForCallback.scrResults, inputForCallback.scrForRelayer)
	}

	return inputForCallback.scrResults, nil
}

func (sc *scProcessor) extractAsyncCallParamsFromTxData(data string) (*vmcommon.AsyncArguments, []byte, error) {
	function, args, err := sc.argsParser.ParseCallData(data)
	dataAsString := function
	if err != nil {
		log.Trace("scProcessor.createSCRsWhenError()", "error parsing args", data)
		return nil, nil, err
	}

	if len(args) < 2 {
		log.Trace("scProcessor.createSCRsWhenError()", "no async params found", data)
		return nil, nil, err
	}

	callIDIndex := len(args) - 3
	callerCallIDIndex := len(args) - 2
	asyncArgs := &vmcommon.AsyncArguments{
		CallID:       args[callIDIndex],
		CallerCallID: args[callerCallIDIndex],
	}

	for index, arg := range args {
		if index != callIDIndex && index != callerCallIDIndex {
			dataAsString += "@" + hex.EncodeToString(arg)
		}
	}

	return asyncArgs, []byte(dataAsString), nil
}

func (sc *scProcessor) reAppendAsyncParamsToTxCallbackData(data string, isCrossShardESDTCall bool, asyncArgs *vmcommon.AsyncArguments) (string, error) {
	newAsyncParams := contexts.CreateCallbackAsyncParams(hooks.NewVMCryptoHook(), asyncArgs)
	var newArgs [][]byte
	if isCrossShardESDTCall {
		function, args, err := sc.argsParser.ParseCallData(data)
		if err != nil {
			log.Trace("scProcessor.createSCRsWhenError()", "error parsing args", data)
			return "", err
		}

		if len(args) < 2 {
			log.Trace("scProcessor.createSCRsWhenError()", "no async params found", data)
			return "", err
		}
		data = function
		newArgs = append(args, newAsyncParams...)
	} else {
		newArgs = newAsyncParams
	}

	for _, arg := range newArgs {
		data += "@" + hex.EncodeToString(arg)
	}

	return data, nil
}

func mergeVMOutputLogs(newVMOutput *vmcommon.VMOutput, vmOutput *vmcommon.VMOutput) {
	if len(vmOutput.Logs) == 0 {
		return
	}

	if newVMOutput.Logs == nil {
		newVMOutput.Logs = make([]*vmcommon.LogEntry, 0, len(vmOutput.Logs))
	}

	newVMOutput.Logs = append(vmOutput.Logs, newVMOutput.Logs...)
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

		return vmOutput, nil
	}

	return vmOutput, nil
}

func (sc *scProcessor) isSameShardSCExecutionAfterBuiltInFunc(
	tx data.TransactionHandler,
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
) (*sameShardExecutionDataAfterBuiltIn, error) {
	noExecutionDTO := &sameShardExecutionDataAfterBuiltIn{
		isSCCallSelfShard:    false,
		callType:             0,
		parsedTransfer:       nil,
		scExecuteOutTransfer: nil,
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return noExecutionDTO, nil
	}

	parsedTransfer, err := sc.esdtTransferParser.ParseESDTTransfers(vmInput.CallerAddr, vmInput.RecipientAddr, vmInput.Function, vmInput.Arguments)
	if err != nil {
		return noExecutionDTO, nil
	}
	if !core.IsSmartContractAddress(parsedTransfer.RcvAddr) {
		return noExecutionDTO, nil
	}

	if sc.shardCoordinator.ComputeId(parsedTransfer.RcvAddr) != sc.shardCoordinator.SelfId() {
		return noExecutionDTO, nil
	}

	callType := determineCallType(tx)
	if callType == vmData.AsynchronousCallBack {
		return &sameShardExecutionDataAfterBuiltIn{
			isSCCallSelfShard:    true,
			callType:             callType,
			parsedTransfer:       parsedTransfer,
			scExecuteOutTransfer: nil,
		}, nil
	}

	if len(parsedTransfer.CallFunction) == 0 {
		return noExecutionDTO, nil
	}

	outAcc, ok := vmOutput.OutputAccounts[string(parsedTransfer.RcvAddr)]
	if !ok {
		return noExecutionDTO, nil
	}
	if len(outAcc.OutputTransfers) != 1 {
		return noExecutionDTO, nil
	}

	scExecuteOutTransfer := outAcc.OutputTransfers[0]
	return &sameShardExecutionDataAfterBuiltIn{
		isSCCallSelfShard:    true,
		callType:             callType,
		parsedTransfer:       parsedTransfer,
		scExecuteOutTransfer: &scExecuteOutTransfer,
	}, nil
}

type inputDataAfterBuiltInCall struct {
	callType             vmData.CallType
	vmInput              *vmcommon.ContractCallInput
	vmOutput             *vmcommon.VMOutput
	parsedTransfer       *vmcommon.ParsedESDTTransfers
	scExecuteOutTransfer *vmcommon.OutputTransfer
	tx                   data.TransactionHandler
	acntSnd              state.UserAccountHandler
	failureContext       *failureContext
}

func (sc *scProcessor) prepareExecutionAfterBuiltInFunc(
	in *inputDataAfterBuiltInCall) (*vmcommon.ContractCallInput, state.UserAccountHandler, error) {

	var newVMInput *vmcommon.ContractCallInput
	if in.callType == vmData.AsynchronousCallBack {
		newVMInput = sc.createVMInputWithAsyncCallBackAfterBuiltIn(in.vmInput, in.vmOutput, in.parsedTransfer)
	} else {
		newVMInput = &vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallerAddr:           in.vmInput.CallerAddr,
				AsyncArguments:       in.vmInput.AsyncArguments,
				Arguments:            in.parsedTransfer.CallArgs,
				CallValue:            big.NewInt(0),
				CallType:             in.callType,
				GasPrice:             in.vmInput.GasPrice,
				GasProvided:          in.scExecuteOutTransfer.GasLimit,
				GasLocked:            in.vmInput.GasLocked,
				OriginalTxHash:       in.vmInput.OriginalTxHash,
				CurrentTxHash:        in.vmInput.CurrentTxHash,
				ReturnCallAfterError: in.vmInput.ReturnCallAfterError,
			},
			RecipientAddr:     in.parsedTransfer.RcvAddr,
			Function:          in.parsedTransfer.CallFunction,
			AllowInitFunction: false,
		}
		newVMInput.ESDTTransfers = in.parsedTransfer.ESDTTransfers
	}

	newDestSC, err := sc.getAccountFromAddress(in.vmInput.RecipientAddr)
	if err != nil {
		in.failureContext.setMessages(err.Error(), []byte(""))
		return newVMInput, newDestSC, nil
	}
	err = sc.checkUpgradePermission(newDestSC, newVMInput)
	if err != nil {
		log.Debug("checkUpgradePermission", "error", err.Error())
		in.failureContext.setMessages(err.Error(), []byte(""))
		return newVMInput, newDestSC, nil
	}

	return newVMInput, newDestSC, nil
}

func (sc *scProcessor) createVMInputWithAsyncCallBackAfterBuiltIn(
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
	parsedTransfer *vmcommon.ParsedESDTTransfers,
) *vmcommon.ContractCallInput {
	arguments := [][]byte{contexts.ReturnCodeToBytes(vmOutput.ReturnCode)}
	gasLimit := vmOutput.GasRemaining

	outAcc, ok := vmOutput.OutputAccounts[string(vmInput.RecipientAddr)]
	if ok && len(outAcc.OutputTransfers) == 1 {
		arguments = [][]byte{}

		gasLimit = outAcc.OutputTransfers[0].GasLimit

		args, err := sc.argsParser.ParseArguments(string(outAcc.OutputTransfers[0].Data))
		log.LogIfError(err, "function", "createVMInputWithAsyncCallBackAfterBuiltIn.ParseArguments")
		arguments = append(arguments, args...)
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
	return sc.processIfErrorWithAddedLogs(
		acntSnd, tx,
		&failureContext{
			txHash:        txHash,
			errorMessage:  returnCode,
			returnMessage: returnMessage,
			snapshot:      snapshot,
			gasLocked:     gasLocked,
			logs:          nil,
		})
}

func (sc *scProcessor) processIfErrorWithAddedLogs(acntSnd state.UserAccountHandler,
	tx data.TransactionHandler,
	failureContext *failureContext,
) error {
	sc.vmOutputCacher.Put(failureContext.txHash, &vmcommon.VMOutput{
		ReturnCode:    vmcommon.SimulateFailed,
		ReturnMessage: string(failureContext.returnMessage),
	}, 0)

	err := sc.accounts.RevertToSnapshot(failureContext.snapshot)
	if err != nil {
		if !core.IsClosingError(err) {
			log.Warn("revert to snapshot", "error", err.Error())
		}

		return err
	}

	if len(failureContext.returnMessage) == 0 {
		failureContext.returnMessage = []byte(failureContext.errorMessage)
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return err
	}

	sc.setEmptyRoothashOnErrorIfSaveKeyValue(tx, acntSnd)

	scrIfError, consumedFee := sc.createSCRsWhenError(
		acntSnd,
		failureContext.txHash,
		tx,
		failureContext.errorMessage,
		failureContext.returnMessage,
		failureContext.gasLocked)
	err = sc.addBackTxValues(acntSnd, scrIfError, tx)
	if err != nil {
		return err
	}

	userErrorLog := createNewLogFromSCRIfError(scrIfError)

	isRecvSelfShard := sc.shardCoordinator.SelfId() == sc.shardCoordinator.ComputeId(scrIfError.RcvAddr)
	if !isRecvSelfShard && !sc.isInformativeTxHandler(scrIfError) {
		err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrIfError})
		if err != nil {
			return err
		}
	}

	relayerLog, err := sc.processForRelayerWhenError(tx, failureContext.txHash, failureContext.returnMessage)
	if err != nil {
		return err
	}

	processIfErrorLogs := make([]*vmcommon.LogEntry, 0)
	processIfErrorLogs = append(processIfErrorLogs, userErrorLog)
	if relayerLog != nil {
		processIfErrorLogs = append(processIfErrorLogs, relayerLog)
	}
	if len(failureContext.logs) > 0 {
		processIfErrorLogs = append(processIfErrorLogs, failureContext.logs...)
	}

	logsTxHash := sc.getOriginalTxHashIfIntraShardRelayedSCR(tx, failureContext.txHash)
	ignorableError := sc.txLogsProcessor.SaveLog(logsTxHash, tx, processIfErrorLogs)
	if ignorableError != nil {
		log.Debug("scProcessor.ProcessIfError() txLogsProcessor.SaveLog()", "error", ignorableError.Error())
	}

	txType, _ := sc.txTypeHandler.ComputeTransactionType(tx)
	isCrossShardMoveBalance := txType == process.MoveBalance && check.IfNil(acntSnd)
	if isCrossShardMoveBalance {
		// move balance was already consumed in sender shard
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), failureContext.txHash)

	err = sc.blockChainHook.SaveNFTMetaDataToSystemAccount(tx)
	if err != nil {
		return err
	}

	return nil
}

// TODO delete this function and actually resolve the issue in the built in function of saveKeyValue
func (sc *scProcessor) setEmptyRoothashOnErrorIfSaveKeyValue(tx data.TransactionHandler, account state.UserAccountHandler) {
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

	if scrForRelayer.Value.Cmp(zero) > 0 {
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
		Identifier: []byte(generalSCRIdentifier),
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
		Identifier: []byte(signalError),
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

	isOriginalTxAsyncCallBack :=
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

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(tx data.TransactionHandler, acntSnd state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	err := sc.checkTxValidity(tx)
	if err != nil {
		log.Debug("invalid transaction", "error", err.Error())
		return 0, err
	}

	sw := core.NewStopWatch()
	sw.Start("deploy")

	failureContext := NewFailureContext()
	returnCode, err := sc.doDeploySmartContract(tx, acntSnd, failureContext)
	if failureContext.processFail {
		failureProcessingError := sc.processIfErrorWithAddedLogs(
			acntSnd,
			tx,
			failureContext)
		if failureProcessingError != nil {
			err = failureProcessingError
		}
	}

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
	failureContext *failureContext,
) (vmcommon.ReturnCode, error) {
	sc.blockChainHook.ResetCounters()
	defer sc.printBlockchainHookCounters(tx)

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		log.Debug("wrong transaction - not empty address", "error", process.ErrWrongTransaction.Error())
		return vmcommon.Ok, process.ErrWrongTransaction
	}

	txHash, err := core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return vmcommon.Ok, err
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return vmcommon.Ok, err
	}

	err = sc.saveAccounts(acntSnd, nil)
	if err != nil {
		log.Debug("saveAccounts error", "error", err)
		return vmcommon.Ok, err
	}

	var vmOutput *vmcommon.VMOutput

	failureContext.makeSnaphot(sc.accounts.JournalLen())
	failureContext.setTxHash(txHash)

	vmInput, vmType, err := sc.createVMDeployInput(tx)
	if err != nil {
		log.Trace("Transaction data invalid", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(""))
		return vmcommon.UserError, nil
	}

	failureContext.setGasLocked(vmInput.GasLocked)

	sc.arwenChangeLocker.RLock()
	vmExec, err := sc.vmContainer.Get(vmType)
	if err != nil {
		sc.arwenChangeLocker.RUnlock()
		log.Trace("VM not found", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(""))
		return vmcommon.UserError, nil
	}

	vmOutput, err = vmExec.RunSmartContractCreate(vmInput)
	sc.arwenChangeLocker.RUnlock()
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(""))
		return vmcommon.UserError, nil
	}

	if vmOutput == nil {
		err = process.ErrNilVMOutput
		log.Trace("run smart contract create", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(""))
		return vmcommon.UserError, nil
	}
	vmOutput.GasRemaining += vmInput.GasLocked
	if vmOutput.ReturnCode != vmcommon.Ok {
		failureContext.setMessages(vmOutput.ReturnCode.String(), []byte(vmOutput.ReturnMessage))
		failureContext.setLogs(vmOutput.Logs)
		return vmcommon.UserError, nil
	}

	err = sc.gasConsumedChecks(tx, vmInput.GasProvided, vmInput.GasLocked, vmOutput)
	if err != nil {
		log.Error("gasConsumedChecks with problem ", "err", err.Error(), "txHash", txHash)
		failureContext.setMessages(err.Error(), []byte("gas consumed exceeded"))
		return vmcommon.ExecutionFailed, nil
	}

	results, err := sc.processVMOutput(&vmInput.VMInput, vmOutput, txHash, tx, acntSnd, failureContext)
	if err != nil {
		log.Trace("Processing error", "error", err.Error())
		failureContext.setMessages(err.Error(), []byte(vmOutput.ReturnMessage))
		return vmcommon.ExecutionFailed, nil
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd) // nolint
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return 0, err
	}
	finalResults, logsFromSCRs := sc.cleanInformativeOnlySCRs(results)

	vmOutput.Logs = append(vmOutput.Logs, logsFromSCRs...)
	err = sc.updateDeveloperRewards(tx, vmOutput, 0)
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

		scGenerated = append(scGenerated, sc.pubkeyConv.SilentEncode(addr, log))
	}

	log.Debug("SmartContract deployed",
		"owner", sc.pubkeyConv.SilentEncode(tx.GetSndAddr(), log),
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
	acntSnd state.UserAccountHandler,
	failureContext *failureContext,
) ([]data.TransactionHandler, error) {

	createdAsyncCallback, scrTxs, err := sc.processSCOutputAccounts(vmInput, vmOutput, tx, txHash)
	if err != nil {
		return nil, err
	}

	scrTxs, _, err = sc.completeOutputProcessingAndCreateCallback(
		&outputDataFromCall{
			vmInput:              vmInput,
			vmOutput:             vmOutput,
			txHash:               txHash,
			tx:                   tx,
			scrTxs:               scrTxs,
			acntSnd:              acntSnd,
			createdAsyncCallback: createdAsyncCallback,
			failureContext:       failureContext,
		})
	if err != nil {
		return nil, err
	}

	errCheck := sc.checkSCRSizeInvariant(scrTxs)
	if errCheck != nil {
		return nil, err
	}

	return scrTxs, nil
}

type outputDataFromCall struct {
	vmInput              *vmcommon.VMInput
	vmOutput             *vmcommon.VMOutput
	txHash               []byte
	tx                   data.TransactionHandler
	scrTxs               []data.TransactionHandler
	acntSnd              state.UserAccountHandler
	createdAsyncCallback bool
	failureContext       *failureContext
}

func (sc *scProcessor) completeOutputProcessingAndCreateCallback(
	outData *outputDataFromCall,
) ([]data.TransactionHandler, vmcommon.ReturnCode, error) {

	sc.penalizeUserIfNeeded(outData.tx, outData.txHash, outData.vmInput.CallType, outData.vmInput.GasProvided, outData.vmOutput)

	scrTxs, scrForSender, scrForRelayer, err := sc.createSCRForSenderAndRelayerAndRefundGas(outData)
	if err != nil {
		return scrTxs, vmcommon.ExecutionFailed, err
	}

	if !outData.createdAsyncCallback {
		scrTxs, err = sc.createAsyncCallBackSCR(
			&inputDataForCallbackSCR{
				vmInput:       outData.vmInput,
				vmOutput:      outData.vmOutput,
				tx:            outData.tx,
				txHash:        outData.txHash,
				scrForSender:  scrForSender,
				scrForRelayer: scrForRelayer,
				scrResults:    scrTxs,
			})
		if err != nil {
			return scrTxs, 0, err
		}
	}

	err = sc.deleteAccounts(outData.vmOutput.DeletedAccounts)
	if err != nil {
		return nil, 0, err
	}

	return scrTxs, 0, nil
}

func (sc *scProcessor) createSCRForSenderAndRelayerAndRefundGas(outData *outputDataFromCall) ([]data.TransactionHandler, *smartContractResult.SmartContractResult, *smartContractResult.SmartContractResult, error) {
	scrForSender, scrForRelayer := sc.createSCRForSenderAndRelayer(
		outData.vmOutput,
		outData.tx,
		outData.txHash,
		outData.vmInput.CallType,
	)

	var err error
	if !check.IfNil(scrForRelayer) {
		outData.scrTxs = append(outData.scrTxs, scrForRelayer)
		err = sc.addGasRefundIfInShard(scrForRelayer.RcvAddr, scrForRelayer.Value)
		if err != nil {
			return outData.scrTxs, scrForSender, scrForRelayer, err
		}
	}

	if !outData.createdAsyncCallback {
		err = sc.addGasRefundIfInShard(scrForSender.RcvAddr, scrForSender.Value)
		if err != nil {
			return outData.scrTxs, scrForSender, scrForRelayer, err
		}
	}

	return outData.scrTxs, scrForSender, scrForRelayer, nil
}

func (sc *scProcessor) checkSCRSizeInvariant(scrTxs []data.TransactionHandler) error {
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

	if core.IsSmartContractAddress(address) {
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

	vmOutput.ReturnMessage += "@"
	sc.gasHandler.SetGasPenalized(vmOutput.GasRemaining, txHash)

	vmOutput.ReturnMessage += fmt.Sprintf("%s for processing: gas provided = %d, gas used = %d",
		TooMuchGasProvidedMessage, gasProvidedForProcessing, gasUsed)
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

	txData := tx.GetData()
	var asyncArgs *vmcommon.AsyncArguments
	if callType == vmData.AsynchronousCall {
		var err error
		asyncArgs, txData, err = sc.extractAsyncCallParamsFromTxData(string(txData))
		if err != nil {
			return nil, nil
		}
	}

	accumulatedSCRData := ""
	esdtReturnData, isCrossShardESDTCall := sc.isCrossShardESDTTransfer(tx.GetSndAddr(), tx.GetRcvAddr(), txData)
	if callType != vmData.AsynchronousCallBack && isCrossShardESDTCall {
		accumulatedSCRData += esdtReturnData
	}

	consumedFee := sc.economicsFee.ComputeTxFee(tx)

	if callType == vmData.AsynchronousCall {
		scr.CallType = vmData.AsynchronousCallBack
		scr.GasPrice = tx.GetGasPrice()
		if tx.GetGasLimit() >= gasLocked {
			scr.GasLimit = gasLocked
			consumedFee = sc.economicsFee.ComputeFeeForProcessing(tx, tx.GetGasLimit()-gasLocked)
		}

		accumulatedSCRData += "@" + core.ConvertToEvenHex(int(vmcommon.UserError))
		accumulatedSCRData += "@" + hex.EncodeToString(returnMessage)

		var err error
		accumulatedSCRData, err = sc.reAppendAsyncParamsToTxCallbackData(accumulatedSCRData, isCrossShardESDTCall, asyncArgs)
		if err != nil {
			return nil, nil
		}
	} else {
		accumulatedSCRData += "@" + hex.EncodeToString([]byte(returnCode))
		if check.IfNil(acntSnd) {
			moveBalanceCost := sc.economicsFee.ComputeMoveBalanceFee(tx)
			consumedFee.Sub(consumedFee, moveBalanceCost)
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

	if vmOutput.ReturnCode != vmcommon.Ok {
		encodedReturnMessage := "@" + hex.EncodeToString([]byte(vmOutput.ReturnMessage))
		result.Data = append(result.Data, encodedReturnMessage...)
	}

	addReturnDataToSCR(vmOutput, result)
}

func (sc *scProcessor) createAsyncCallBackSCRFromVMOutput(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte,
	asyncParams *vmcommon.AsyncArguments,
) (*smartContractResult.SmartContractResult, error) {
	origScr, ok := (tx).(*smartContractResult.SmartContractResult)
	if !ok {
		return nil, process.ErrWrongTransactionType
	}
	scr := &smartContractResult.SmartContractResult{
		Value:          big.NewInt(0),
		RcvAddr:        tx.GetSndAddr(),
		SndAddr:        tx.GetRcvAddr(),
		PrevTxHash:     txHash,
		GasPrice:       tx.GetGasPrice(),
		ReturnMessage:  []byte(vmOutput.ReturnMessage),
		OriginalSender: origScr.GetOriginalSender(),
	}
	setOriginalTxHash(scr, txHash, tx)
	relayedTx, isRelayed := isRelayedTx(tx)
	if isRelayed {
		scr.RelayedValue = big.NewInt(0)
		scr.RelayerAddr = relayedTx.RelayerAddr
	}

	sc.addVMOutputResultsToSCR(vmOutput, scr)

	var err error
	scr.Data, err = contexts.AppendAsyncArgumentsToCallbackCallData(
		hooks.NewVMCryptoHook(),
		scr.Data,
		asyncParams,
		sc.argsParser.ParseArguments)

	if err != nil {
		return nil, err
	}

	return scr, nil
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
	scrIndex uint32,
) (bool, []internalIndexedScr) {
	if callType == vmData.AsynchronousCall && bytes.Equal(outAcc.Address, tx.GetSndAddr()) {
		result := createBaseSCR(outAcc, tx, txHash, 0)
		sc.addVMOutputResultsToSCR(vmOutput, result)
		return true, []internalIndexedScr{{result, scrIndex}}
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
	transferNonce := uint64(index)
	result := createBaseSCR(outAcc, tx, txHash, transferNonce)

	if outputTransfer.Value != nil {
		result.Value.Set(outputTransfer.Value)
	}

	result.GasLimit = outputTransfer.GasLimit
	result.CallType = outputTransfer.CallType
	setOriginalTxHash(result, txHash, tx)
	result.OriginalSender = GetOriginalSenderForTx(tx)
	return result
}

// GetOriginalSenderForTx obtains the original's sender address depending on transaction's type
func GetOriginalSenderForTx(tx data.TransactionHandler) []byte {
	origScr, ok := (tx).(*smartContractResult.SmartContractResult)
	if ok {
		return origScr.GetOriginalSender()
	} else {
		return tx.GetSndAddr()
	}
}

func (sc *scProcessor) createSmartContractResults(
	vmInput *vmcommon.VMInput,
	vmOutput *vmcommon.VMOutput,
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) (bool, []internalIndexedScr) {

	nextAvailableScrIndex := vmOutput.GetNextAvailableOutputTransferIndex()

	result := sc.createSCRFromStakingSC(outAcc, tx, txHash)
	if !check.IfNil(result) {
		return false, []internalIndexedScr{{result, nextAvailableScrIndex}}
	}

	lenOutTransfers := len(outAcc.OutputTransfers)
	if lenOutTransfers == 0 {
		return sc.createSCRIfNoOutputTransfer(vmOutput, vmInput.CallType, outAcc, tx, txHash, nextAvailableScrIndex)
	}

	createdAsyncCallBack := false
	indexedSCResults := make([]internalIndexedScr, 0, len(outAcc.OutputTransfers))
	for i, outputTransfer := range outAcc.OutputTransfers {
		result = sc.preprocessOutTransferToSCR(i, outputTransfer, outAcc, tx, txHash)

		isCrossShard := sc.shardCoordinator.ComputeId(outAcc.Address) != sc.shardCoordinator.SelfId()

		useSenderAddressFromOutTransfer :=
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
		} else {
			result.Data, _ = contexts.AppendTransferAsyncDataToCallData(
				outputTransfer.Data,
				outputTransfer.AsyncData,
				sc.argsParser.ParseArguments,
			)
		}

		if result.CallType == vmData.AsynchronousCall {
			if isCrossShard {
				result.GasLimit += outputTransfer.GasLocked
				lastArgAsGasLocked := "@" + hex.EncodeToString(big.NewInt(0).SetUint64(outputTransfer.GasLocked).Bytes())
				result.Data = append(result.Data, []byte(lastArgAsGasLocked)...)
			}
		}

		indexedSCResults = append(indexedSCResults, internalIndexedScr{result, outputTransfer.Index})
	}

	return createdAsyncCallBack, indexedSCResults
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
	isAsyncTransferBackToSender := vmInput.CallType == vmData.AsynchronousCall &&
		bytes.Equal(outAcc.Address, tx.GetSndAddr())
	if !isAsyncTransferBackToSender {
		return false
	}

	if !isCrossShard {
		return false
	}

	if !sc.isTransferWithNoAdditionalData(outputTransfer.SenderAddress, outAcc.Address, outputTransfer.Data) {
		return false
	}

	addReturnDataToSCR(vmOutput, result)
	result.CallType = vmData.AsynchronousCallBack
	result.GasLimit, _ = core.SafeAddUint64(result.GasLimit, vmOutput.GasRemaining)

	var err error
	asyncParams := contexts.CreateCallbackAsyncParams(hooks.NewVMCryptoHook(), vmInput.AsyncArguments)
	dataBuilder, err := sc.prependAsyncParamsToData(asyncParams, outputTransfer.Data, int(vmOutput.ReturnCode))
	if err != nil {
		log.Debug("processed built in functions error (async params extraction)", "error", err.Error())
		return false
	}

	result.Data = dataBuilder.ToBytes()

	return true
}

func (sc *scProcessor) prependAsyncParamsToData(asyncParams [][]byte, data []byte, returnCode int) (*txDataBuilder.TxDataBuilder, error) {
	var args [][]byte
	var err error

	callData := txDataBuilder.NewBuilder()

	// These nested conditions ensure that prependAsyncParamsToData() can handle
	// data strings with or without a function as the first token in the string.
	// The string "ESDTTransfer@...@..." requires ParseCallData().
	// The string "@...@..." requires ParseArguments() instead.
	function, args, err := sc.argsParser.ParseCallData(string(data))
	if err == nil {
		callData.Func(function)
	} else {
		args, err = sc.argsParser.ParseArguments(string(data))
		if err != nil {
			return nil, err
		}
	}

	for _, arg := range args {
		callData.Bytes(arg)
	}

	callData.Int(returnCode)

	for _, asyncParam := range asyncParams {
		callData.Bytes(asyncParam)
	}

	return callData, nil
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
			ReturnMessage:  []byte("gas refund for relayer"),
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
	if callType == vmData.AsynchronousCall {
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
	tx data.TransactionHandler,
	txHash []byte,
) (bool, []data.TransactionHandler, error) {
	return NewVMOutputAccountsProcessor(sc, vmInput, vmOutput, tx, txHash).
		Run()
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
		currentEpoch := sc.enableEpochsHandler.GetCurrentEpoch()
		if sc.shardCoordinator.SelfId() == core.MetachainShardId && !sc.enableEpochsHandler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(currentEpoch) {
			returnCode, err = sc.ExecuteSmartContractTransaction(scr, sndAcc, dstAcc)
			return returnCode, err
		}
		returnCode, err = sc.ExecuteBuiltInFunction(scr, sndAcc, dstAcc)
		return returnCode, err
	}

	err = process.ErrWrongTransaction
	return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot, gasLocked)
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

func (sc *scProcessor) getGasLockedFromSCR(scr *smartContractResult.SmartContractResult) uint64 {
	if scr.CallType != vmData.AsynchronousCall {
		return 0
	}

	_, arguments, err := sc.argsParser.ParseCallData(string(scr.Data))
	if err != nil {
		return 0
	}
	_, gasLocked := getAsyncCallGasLockFromTxData(scr.CallType, arguments)
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

	isTransferValueOnlySCR := len(scResult.Data) == 0 && scResult.Value.Cmp(zero) > 0
	shouldGenerateCompleteEvent := isReturnOKTxHandler(scResult) || isTransferValueOnlySCR
	if shouldGenerateCompleteEvent {
		completedTxLog := createCompleteEventLog(scResult, txHash)
		ignorableError := sc.txLogsProcessor.SaveLog(txHash, scResult, []*vmcommon.LogEntry{completedTxLog})
		if ignorableError != nil {
			log.Debug("scProcessor.finishSCExecution txLogsProcessor.SaveLog()", "error", ignorableError.Error())
		}
	}

	return nil
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
		Identifier: []byte(completedTxEvent),
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

// GetTxLogsProcessor is a getter for txLogsProcessor
func (sc *scProcessor) GetTxLogsProcessor() process.TransactionLogProcessor {
	return sc.txLogsProcessor
}

// GetScrForwarder is a getter for scrForwarder
func (sc *scProcessor) GetScrForwarder() process.IntermediateTransactionHandler {
	return sc.scrForwarder
}

func (sc *scProcessor) printBlockchainHookCounters(tx data.TransactionHandler) {
	if logCounters.GetLevel() > logger.LogTrace {
		return
	}

	receiver, _ := sc.pubkeyConv.Encode(tx.GetRcvAddr())
	sender, _ := sc.pubkeyConv.Encode(tx.GetSndAddr())

	logCounters.Trace("blockchain hook counters",
		"counters", sc.getBlockchainHookCountersString(),
		"tx hash", sc.computeTxHashUnsafe(tx),
		"receiver", receiver,
		"sender", sender,
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
