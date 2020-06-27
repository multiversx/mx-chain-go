package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.SmartContractResultProcessor = (*scProcessor)(nil)
var _ process.SmartContractProcessor = (*scProcessor)(nil)

var log = logger.GetOrCreate("process/smartcontract")

const executeDurationDebugThreshold = float64(0.030)
const executeDurationWarnThreshold = float64(0.100)

var zero = big.NewInt(0)

type scProcessor struct {
	accounts         state.AccountsAdapter
	tempAccounts     process.TemporaryAccountsHandler
	pubkeyConv       core.PubkeyConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	vmContainer      process.VirtualMachinesContainer
	argsParser       process.ArgumentsParser
	builtInFunctions process.BuiltInFunctionContainer

	scrForwarder  process.IntermediateTransactionHandler
	txFeeHandler  process.TransactionFeeHandler
	economicsFee  process.FeeHandler
	txTypeHandler process.TxTypeHandler
	gasHandler    process.GasHandler

	txLogsProcessor process.TransactionLogProcessor
}

// ArgsNewSmartContractProcessor defines the arguments needed for new smart contract processor
type ArgsNewSmartContractProcessor struct {
	VmContainer      process.VirtualMachinesContainer
	ArgsParser       process.ArgumentsParser
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	AccountsDB       state.AccountsAdapter
	TempAccounts     process.TemporaryAccountsHandler
	PubkeyConv       core.PubkeyConverter
	Coordinator      sharding.Coordinator
	ScrForwarder     process.IntermediateTransactionHandler
	TxFeeHandler     process.TransactionFeeHandler
	EconomicsFee     process.FeeHandler
	TxTypeHandler    process.TxTypeHandler
	GasHandler       process.GasHandler
	BuiltInFunctions process.BuiltInFunctionContainer
	TxLogsProcessor  process.TransactionLogProcessor
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
	if check.IfNil(args.TempAccounts) {
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
	if check.IfNil(args.BuiltInFunctions) {
		return nil, process.ErrNilBuiltInFunction
	}
	if check.IfNil(args.TxLogsProcessor) {
		return nil, process.ErrNilTxLogsProcessor
	}

	sc := &scProcessor{
		vmContainer:      args.VmContainer,
		argsParser:       args.ArgsParser,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		accounts:         args.AccountsDB,
		tempAccounts:     args.TempAccounts,
		pubkeyConv:       args.PubkeyConv,
		shardCoordinator: args.Coordinator,
		scrForwarder:     args.ScrForwarder,
		txFeeHandler:     args.TxFeeHandler,
		economicsFee:     args.EconomicsFee,
		txTypeHandler:    args.TxTypeHandler,
		gasHandler:       args.GasHandler,
		builtInFunctions: args.BuiltInFunctions,
		txLogsProcessor:  args.TxLogsProcessor,
	}

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
) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	sw := sc.monitorBeginExecuteSmartContractTransaction()
	err := sc.doExecuteSmartContractTransaction(tx, acntSnd, acntDst)
	sc.monitorEndExecuteSmartContractTransaction(sw, tx, err)
	return err
}

func (sc *scProcessor) monitorBeginExecuteSmartContractTransaction() *core.StopWatch {
	sw := core.NewStopWatch()
	sw.Start("execute")
	return sw
}

func (sc *scProcessor) monitorEndExecuteSmartContractTransaction(sw *core.StopWatch, tx data.TransactionHandler, err error) {
	sw.Stop("execute")
	duration := sw.GetMeasurement("execute")

	logFunc := log.Trace
	if duration > executeDurationDebugThreshold {
		logFunc = log.Debug
	}
	if duration > executeDurationWarnThreshold {
		logFunc = log.Warn
	}

	logFunc("scProcessor.ExecuteSmartContractTransaction()", "sc", tx.GetRcvAddr(), "data", string(tx.GetData()), "duration", duration, "err", err)
}

func (sc *scProcessor) doExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		log.Debug("process sc payment error", "error", err.Error())
		return err
	}

	var txHash []byte
	txHash, err = core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return err
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		log.Debug("saveAccounts error", "error", err)
		return err
	}

	returnMessage := ""

	var vmOutput *vmcommon.VMOutput
	var executedBuiltIn bool
	snapshot := sc.accounts.JournalLen()
	defer func() {
		if err != nil && !executedBuiltIn {
			if vmOutput != nil {
				returnMessage = vmOutput.ReturnMessage
			}

			errNotCritical := sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot)
			if errNotCritical != nil {
				log.Debug("error while processing error in smart contract processor")
			}
		}
	}()

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		returnMessage = "cannot prepare smart contract call"
		log.Debug("prepare smart contract call error", "error", err.Error())
		return nil
	}

	var vmInput *vmcommon.ContractCallInput
	vmInput, err = sc.createVMCallInput(tx)
	if err != nil {
		returnMessage = "cannot create VMInput, check the transaction data field"
		log.Debug("create vm call input error", "error", err.Error())
		return nil
	}

	executedBuiltIn, err = sc.resolveBuiltInFunctions(txHash, tx, acntSnd, acntDst, vmInput)
	if err != nil {
		returnMessage = "cannot resolve build in function"
		log.Debug("processed built in functions error", "error", err.Error())
		return err
	}
	if executedBuiltIn {
		return nil
	}

	if check.IfNil(acntDst) {
		return process.ErrNilSCDestAccount
	}

	var vm vmcommon.VMExecutionHandler
	vm, err = findVMByTransaction(sc.vmContainer, tx)
	if err != nil {
		returnMessage = "cannot get vm from address"
		log.Debug("get vm from address error", "error", err.Error())
		return nil
	}

	vmOutput, err = vm.RunSmartContractCall(vmInput)
	if err != nil {
		log.Debug("run smart contract call error", "error", err.Error())
		return nil
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		returnMessage = "cannot save accounts"
		log.Debug("saveAccounts error", "error", err)
		return nil
	}

	var consumedFee *big.Int
	var results []data.TransactionHandler
	results, consumedFee, err = sc.processVMOutput(vmOutput, txHash, tx, acntSnd, vmInput.CallType)
	if err != nil {
		log.Trace("process vm output returned with problem ", "err", err.Error())
		return nil
	}

	finalResults := sc.deleteSCRsWithValueZeroGoingToMeta(results)
	err = sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Debug("AddIntermediateTransactions error", "error", err.Error())
		return nil
	}

	ignorableError := sc.txLogsProcessor.SaveLog(txHash, tx, vmOutput.Logs)
	if ignorableError != nil {
		log.Debug("txLogsProcessor.SaveLog() error", "error", ignorableError.Error())
	}

	newDeveloperReward := core.GetPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())

	acntDst, err = sc.reloadLocalAccount(acntDst)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return nil
	}

	acntDst.AddToDeveloperReward(newDeveloperReward)
	sc.txFeeHandler.ProcessTransactionFee(consumedFee, newDeveloperReward, txHash)

	err = sc.accounts.SaveAccount(acntDst)
	if err != nil {
		log.Debug("error saving account")
	}

	return nil
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

	// return error here only if acntSnd is not nil - so this is sender shard
	vmOutput, err := builtIn.ProcessBuiltinFunction(acntSnd, acntDst, vmInput)
	if err != nil {
		if !check.IfNil(acntSnd) {
			log.Trace("built in function error at sender", "err", err, "function", vmInput.Function)
			return true, err
		}

		vmOutput = &vmcommon.VMOutput{ReturnCode: vmcommon.UserError, ReturnMessage: err.Error()}
	}

	scrResults := make([]data.TransactionHandler, 0, len(vmOutput.OutputAccounts)+1)

	outputAccounts := sortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		storageUpdates := getSortedStorageUpdates(outAcc)
		scTx := sc.createSmartContractResult(outAcc, tx, txHash, storageUpdates)
		scrResults = append(scrResults, scTx)
	}

	refund := big.NewInt(0)
	if vmOutput.GasRefund != nil {
		refund = vmOutput.GasRefund
	}
	scrForSender, consumedFee := sc.createSCRForSender(
		refund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		vmOutput.ReturnMessage,
		tx,
		txHash,
		acntSnd,
		vmcommon.DirectCall,
	)
	if !check.IfNil(acntSnd) {
		err = acntSnd.AddToBalance(scrForSender.Value)
		if err != nil {
			return true, err
		}
	}

	scrResults = append(scrResults, scrForSender)
	finalResults := sc.deleteSCRsWithValueZeroGoingToMeta(scrResults)
	err = sc.scrForwarder.AddIntermediateTransactions(finalResults)
	if err != nil {
		log.Debug("AddIntermediateTransactions error", "error", err.Error())
		return true, err
	}

	if check.IfNil(acntSnd) {
		// it was already consumed in sender shard
		consumedFee = big.NewInt(0)
	}

	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)
	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)

	return true, sc.saveAccounts(acntSnd, acntDst)
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
	}

	consumedFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GetGasLimit()), big.NewInt(0).SetUint64(tx.GetGasPrice()))
	scrIfError, err := sc.createSCRsWhenError(txHash, tx, returnCode, returnMessage)
	if err != nil {
		return err
	}

	if !check.IfNil(acntSnd) {
		err = acntSnd.AddToBalance(tx.GetValue())
		if err != nil {
			return err
		}

		err = sc.accounts.SaveAccount(acntSnd)
		if err != nil {
			log.Debug("error saving account")
		}
	} else {
		moveBalanceCost := sc.economicsFee.ComputeFee(tx)
		consumedFee.Sub(consumedFee, moveBalanceCost)
	}

	err = sc.scrForwarder.AddIntermediateTransactions(scrIfError)
	if err != nil {
		return err
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)

	return nil
}

func (sc *scProcessor) prepareSmartContractCall(tx data.TransactionHandler, acntSnd state.UserAccountHandler) error {
	nonce := tx.GetNonce()
	if acntSnd != nil && !acntSnd.IsInterfaceNil() {
		nonce = acntSnd.GetNonce()
	}

	txValue := big.NewInt(0).Set(tx.GetValue())
	sc.tempAccounts.AddTempAccount(tx.GetSndAddr(), txValue, nonce)

	return nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	err := sc.checkTxValidity(tx)
	if err != nil {
		log.Debug("invalid transaction", "error", err.Error())
		return err
	}

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		log.Debug("wrong transaction", "error", process.ErrWrongTransaction.Error())
		return process.ErrWrongTransaction
	}

	txHash, err := core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return err
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	err = sc.saveAccounts(acntSnd, nil)
	if err != nil {
		log.Debug("saveAccounts error", "error", err)
		return err
	}

	var vmOutput *vmcommon.VMOutput
	snapshot := sc.accounts.JournalLen()
	defer func() {
		if err != nil {
			returnMessage := ""
			if vmOutput != nil {
				returnMessage = vmOutput.ReturnMessage
			}

			errNotCritical := sc.ProcessIfError(acntSnd, txHash, tx, err.Error(), []byte(returnMessage), snapshot)
			if errNotCritical != nil {
				log.Debug("error while processing error in smart contract processor")
			}
		}
	}()

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		log.Debug("Transaction error", "error", err.Error())
		return nil
	}

	vmInput, vmType, err := sc.createVMDeployInput(tx)
	if err != nil {
		log.Debug("Transaction error", "error", err.Error())
		return nil
	}

	vm, err := sc.vmContainer.Get(vmType)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return nil
	}

	vmOutput, err = vm.RunSmartContractCreate(vmInput)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return nil
	}

	err = sc.accounts.SaveAccount(acntSnd)
	if err != nil {
		log.Debug("Save account error", "error", err.Error())
		return nil
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, txHash, tx, acntSnd, vmInput.CallType)
	if err != nil {
		log.Trace("Processing error", "error", err.Error())
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("AddIntermediate Transaction error", "error", err.Error())
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee, big.NewInt(0), txHash)
	sc.printScDeployed(vmOutput, tx)

	return nil
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

func (sc *scProcessor) computeGasRemainingFromVMOutput(vmOutput *vmcommon.VMOutput) uint64 {
	gasRemaining := vmOutput.GasRemaining
	for _, outAcc := range vmOutput.OutputAccounts {
		if outAcc.GasLimit == 0 {
			continue
		}

		if gasRemaining < outAcc.GasLimit {
			log.Error("gasLimit mismatch in vmOutput - too much consumed gas - more then tx.GasLimit")
			return 0
		}

		gasRemaining -= outAcc.GasLimit
	}

	return gasRemaining
}

func (sc *scProcessor) processVMOutput(
	vmOutput *vmcommon.VMOutput,
	txHash []byte,
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
	callType vmcommon.CallType,
) ([]data.TransactionHandler, *big.Int, error) {
	if vmOutput == nil {
		return nil, nil, process.ErrNilVMOutput
	}
	if check.IfNil(tx) {
		return nil, nil, process.ErrNilTransaction
	}

	gasRemainedForSender := sc.computeGasRemainingFromVMOutput(vmOutput)
	scrForSender, consumedFee := sc.createSCRForSender(
		vmOutput.GasRefund,
		gasRemainedForSender,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		vmOutput.ReturnMessage,
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

		if callType == vmcommon.AsynchronousCall {
			return []data.TransactionHandler{scrForSender}, consumedFee, nil
		}

		return nil, nil, fmt.Errorf(vmOutput.ReturnCode.String())
	}

	outPutAccounts := sortVMOutputInsideData(vmOutput)

	scrTxs, err := sc.processSCOutputAccounts(outPutAccounts, tx, txHash)
	if err != nil {
		return nil, nil, err
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return nil, nil, err
	}

	totalGasConsumed := tx.GetGasLimit() - gasRemainedForSender
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

func sortVMOutputInsideData(vmOutput *vmcommon.VMOutput) []*vmcommon.OutputAccount {
	sort.Slice(vmOutput.DeletedAccounts, func(i, j int) bool {
		return bytes.Compare(vmOutput.DeletedAccounts[i], vmOutput.DeletedAccounts[j]) < 0
	})
	sort.Slice(vmOutput.TouchedAccounts, func(i, j int) bool {
		return bytes.Compare(vmOutput.TouchedAccounts[i], vmOutput.TouchedAccounts[j]) < 0
	})

	outPutAccounts := make([]*vmcommon.OutputAccount, len(vmOutput.OutputAccounts))
	i := 0
	for _, outAcc := range vmOutput.OutputAccounts {
		outPutAccounts[i] = outAcc
		i++
	}

	sort.Slice(outPutAccounts, func(i, j int) bool {
		return bytes.Compare(outPutAccounts[i].Address, outPutAccounts[j].Address) < 0
	})

	return outPutAccounts
}

func getSortedStorageUpdates(account *vmcommon.OutputAccount) []*vmcommon.StorageUpdate {
	storageUpdates := make([]*vmcommon.StorageUpdate, len(account.StorageUpdates))
	i := 0
	for _, update := range account.StorageUpdates {
		storageUpdates[i] = update
		i++
	}

	sort.Slice(storageUpdates, func(i, j int) bool {
		return bytes.Compare(storageUpdates[i].Offset, storageUpdates[j].Offset) < 0
	})

	return storageUpdates
}

func (sc *scProcessor) createSCRsWhenError(
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
) ([]data.TransactionHandler, error) {
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
		Code:          nil,
		Data:          []byte("@" + hex.EncodeToString([]byte(returnCode)) + "@" + hex.EncodeToString(txHash)),
		PrevTxHash:    txHash,
		ReturnMessage: returnMessage,
	}
	setOriginalTxHash(scr, txHash, tx)

	resultedScrs := []data.TransactionHandler{scr}

	return resultedScrs, nil
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
	result.Nonce = outAcc.Nonce
	result.RcvAddr = outAcc.Address
	result.SndAddr = tx.GetRcvAddr()
	result.Code = outAcc.Code
	result.Data = outAcc.Data
	if bytes.Equal(result.GetRcvAddr(), factory.StakingSCAddress) {
		//TODO: write directly from staking smart contract to the validator trie - it gets complicated in reverts
		// storage update for staking contract is used in stakingToPeer
		result.Data = append(result.Data, sc.argsParser.CreateDataFromStorageUpdate(storageUpdates)...)
	}
	result.GasLimit = outAcc.GasLimit
	result.GasPrice = tx.GetGasPrice()
	result.PrevTxHash = txHash
	result.CallType = outAcc.CallType
	setOriginalTxHash(result, txHash, tx)

	return result
}

// createSCRForSender(vmOutput, tx, txHash, acntSnd)
// give back the user the unused gas money
func (sc *scProcessor) createSCRForSender(
	gasRefund *big.Int,
	gasRemaining uint64,
	returnCode vmcommon.ReturnCode,
	returnData [][]byte,
	returnMessage string,
	tx data.TransactionHandler,
	txHash []byte,
	acntSnd state.UserAccountHandler,
	callType vmcommon.CallType,
) (*smartContractResult.SmartContractResult, *big.Int) {
	storageFreeRefund := big.NewInt(0).Mul(gasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))

	consumedFee := big.NewInt(0)
	consumedFee.Mul(big.NewInt(0).SetUint64(tx.GetGasPrice()), big.NewInt(0).SetUint64(tx.GetGasLimit()))

	refundErd := big.NewInt(0)
	refundErd.Mul(big.NewInt(0).SetUint64(gasRemaining), big.NewInt(0).SetUint64(tx.GetGasPrice()))
	consumedFee.Sub(consumedFee, refundErd)

	rcvAddress := tx.GetSndAddr()
	if callType == vmcommon.AsynchronousCallBack {
		rcvAddress = tx.GetRcvAddr()
	}

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = big.NewInt(0).Add(refundErd, storageFreeRefund)
	scTx.RcvAddr = rcvAddress
	scTx.SndAddr = tx.GetRcvAddr()
	scTx.Nonce = tx.GetNonce() + 1
	scTx.PrevTxHash = txHash
	scTx.GasLimit = gasRemaining
	scTx.GasPrice = tx.GetGasPrice()
	scTx.ReturnMessage = []byte(returnMessage)
	setOriginalTxHash(scTx, txHash, tx)

	if callType == vmcommon.AsynchronousCall {
		scTx.CallType = vmcommon.AsynchronousCallBack
		scTx.Data = []byte("@" + hex.EncodeToString(big.NewInt(int64(returnCode)).Bytes()))
	} else {
		scTx.Data = []byte("@" + hex.EncodeToString([]byte(returnCode.String())))
	}

	for _, retData := range returnData {
		scTx.Data = append(scTx.Data, []byte("@"+hex.EncodeToString(retData))...)
	}

	if check.IfNil(acntSnd) {
		// cross shard move balance fee was already consumed at sender shard
		moveBalanceCost := sc.economicsFee.ComputeFee(tx)
		consumedFee.Sub(consumedFee, moveBalanceCost)
		return scTx, consumedFee
	}

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

		storageUpdates := getSortedStorageUpdates(outAcc)
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

		err = sc.updateSmartContractCode(acc, outAcc, tx)
		if err != nil {
			return nil, err
		}

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

func (sc *scProcessor) updateSmartContractCode(
	scAccount state.UserAccountHandler,
	outputAccount *vmcommon.OutputAccount,
	tx data.TransactionHandler,
) error {
	if len(outputAccount.Code) == 0 {
		return nil
	}

	codeMetadata := vmcommon.CodeMetadataFromBytes(scAccount.GetCodeMetadata())

	isDeployment := len(scAccount.GetCode()) == 0
	if isDeployment {
		scAccount.SetOwnerAddress(tx.GetSndAddr())
		scAccount.SetCodeMetadata(outputAccount.CodeMetadata)
		scAccount.SetCode(outputAccount.Code)
		log.Trace("updateSmartContractCode(): created", "address", sc.pubkeyConv.Encode(outputAccount.Address))
		return nil
	}

	isSenderOwner := bytes.Equal(scAccount.GetOwnerAddress(), tx.GetSndAddr())
	isUpgrade := !isDeployment && isSenderOwner && codeMetadata.Upgradeable
	if isUpgrade {
		scAccount.SetCodeMetadata(outputAccount.CodeMetadata)
		scAccount.SetCode(outputAccount.Code)
		log.Trace("updateSmartContractCode(): upgraded", "address", sc.pubkeyConv.Encode(outputAccount.Address))
		return nil
	}

	log.Trace("updateSmartContractCode() nothing changed", "address", outputAccount.Address)
	// TODO: change to return some error when IELE is updated. Currently IELE sends the code in output account even for normal SC RUN
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
func (sc *scProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	if scr == nil {
		return process.ErrNilSmartContractResult
	}

	log.Trace("scProcessor.ProcessSmartContractResult()", "sender", scr.GetSndAddr(), "receiver", scr.GetRcvAddr())

	var err error
	txHash, err := core.CalculateHash(sc.marshalizer, sc.hasher, scr)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return err
	}

	snapshot := sc.accounts.JournalLen()
	defer func() {
		if err != nil {
			errNotCritical := sc.ProcessIfError(nil, txHash, scr, err.Error(), scr.ReturnMessage, snapshot)
			if errNotCritical != nil {
				log.Debug("error while processing error in smart contract processor")
			}
		}
	}()

	dstAcc, err := sc.getAccountFromAddress(scr.RcvAddr)
	if err != nil {
		return nil
	}
	if check.IfNil(dstAcc) {
		err = process.ErrNilSCDestAccount
		return nil
	}

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
		return nil
	case process.SCDeployment:
		err = process.ErrSCDeployFromSCRIsNotPermitted
		return nil
	case process.SCInvoking:
		err = sc.ExecuteSmartContractTransaction(scr, nil, dstAcc)
		return nil
	case process.BuiltInFunctionCall:
		err = sc.ExecuteSmartContractTransaction(scr, nil, dstAcc)
		return nil
	}

	err = process.ErrWrongTransaction
	return nil
}

func (sc *scProcessor) processSimpleSCR(
	scResult *smartContractResult.SmartContractResult,
	dstAcc state.UserAccountHandler,
) error {
	outputAccount := &vmcommon.OutputAccount{
		Code:         scResult.Code,
		CodeMetadata: scResult.CodeMetadata,
		Address:      scResult.RcvAddr,
	}

	err := sc.updateSmartContractCode(dstAcc, outputAccount, scResult)
	if err != nil {
		return err
	}

	if scResult.Value != nil {
		err = dstAcc.AddToBalance(scResult.Value)
		if err != nil {
			return err
		}
	}

	return sc.accounts.SaveAccount(dstAcc)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	return sc == nil
}
