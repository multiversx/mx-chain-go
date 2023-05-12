package smartContract

import (
	"fmt"
	"sort"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type sovereignSCProcessor struct {
	blockChainHook   process.BlockChainHookHandler
	shardCoordinator sharding.Coordinator
	pubkeyConv       core.PubkeyConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	txTypeHandler    process.TxTypeHandler
	accounts         state.AccountsAdapter
}

func (sc *sovereignSCProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
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

	if check.IfNil(dstAcc) {
		err = process.ErrNilSCDestAccount
		return returnCode, err
	}

	//snapshot := sc.accounts.JournalLen()
	process.DisplayProcessTxDetails(
		"ProcessSmartContractResult: receiver account details",
		dstAcc,
		scr,
		txHash,
		sc.pubkeyConv,
	)

	txType, _ := sc.txTypeHandler.ComputeTransactionType(scr)
	switch txType {
	case process.BuiltInFunctionCall: //todo: also check multiesdttransfer
		returnCode, err = sc.ExecuteBuiltInFunction(scr, nil, dstAcc) //nil sender account
		return returnCode, err
	default:
		err = process.ErrWrongTransaction
	}

	return returnCode, err
	//return returnCode, sc.ProcessIfError(sndAcc, txHash, scr, err.Error(), scr.ReturnMessage, snapshot, gasLocked)
}

func (sc *sovereignSCProcessor) getAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	addrShard := sc.shardCoordinator.ComputeId(address)
	if addrShard != core.SovereignChainShardId {
		return nil, nil
	}

	account, err := sc.accounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAccount, nil
}

func (sc *sovereignSCProcessor) IsInterfaceNil() bool {
	return sc == nil
}

// ExecuteBuiltInFunction  processes the transaction, executes the built in function call and subsequent results
func (sc *sovereignSCProcessor) ExecuteBuiltInFunction(
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

func (sc *sovereignSCProcessor) doExecuteBuiltInFunction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	sc.blockChainHook.ResetCounters()
	defer sc.printBlockchainHookCounters(tx)

	/*
		returnCode, vmInput, txHash, err := sc.prepareExecution(tx, acntSnd, acntDst, true)
		if err != nil || returnCode != vmcommon.Ok {
			return returnCode, err
		}

		snapshot := sc.accounts.JournalLen()
		if !sc.enableEpochsHandler.IsBuiltInFunctionsFlagEnabled() {
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
			tmpCreatedAsyncCallback, scTxs := sc.createSmartContractResults(vmOutput, vmInput.CallType, outAcc, tx, txHash)
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

			if sc.enableEpochsHandler.IsRepairCallbackFlagEnabled() {
				sc.penalizeUserIfNeeded(tx, txHash, newVMInput.CallType, newVMInput.GasProvided, newVMOutput)
			}

			outPutAccounts := process.SortVMOutputInsideData(newVMOutput)
			var newSCRTxs []data.TransactionHandler
			tmpCreatedAsyncCallback := false
			tmpCreatedAsyncCallback, newSCRTxs, err = sc.processSCOutputAccounts(newVMOutput, vmInput.CallType, outPutAccounts, tx, txHash)
			if err != nil {
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
			if sc.enableEpochsHandler.IsRepairCallbackFlagEnabled() {
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
	*/
	return 0, nil
}

// todo: reuse this from scProc
func (sc *sovereignSCProcessor) printBlockchainHookCounters(tx data.TransactionHandler) {
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

func (sc *sovereignSCProcessor) getBlockchainHookCountersString() string {
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

// this function should only be called for logging reasons, since it does not perform sanity checks
func (sc *sovereignSCProcessor) computeTxHashUnsafe(tx data.TransactionHandler) []byte {
	txHash, _ := core.CalculateHash(sc.marshalizer, sc.hasher, tx)

	return txHash
}
