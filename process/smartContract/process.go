package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-vm-common"
)

// claimDeveloperRewardsFunctionName is a constant which defines the name for the claim developer rewards function
const claimDeveloperRewardsFunctionName = "ClaimDeveloperRewards"

// changeOwnerAddressFunctionName is a constant which defines the name for the change owner address function
const changeOwnerAddressFunctionName = "ChangeOwnerAddress"

// builtInFunctionBaseCostMultiplier is a constant which defines the multiplier to calculate the transactions gas cost
// when using built-in protocol functions
const builtInFunctionBaseCostMultiplier = 2

var log = logger.GetOrCreate("process/smartcontract")

type scProcessor struct {
	accounts         state.AccountsAdapter
	tempAccounts     process.TemporaryAccountsHandler
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	vmContainer      process.VirtualMachinesContainer
	argsParser       process.ArgumentsParser
	isCallBack       bool
	builtInFunctions map[string]process.BuiltinFunction

	scrForwarder  process.IntermediateTransactionHandler
	txFeeHandler  process.TransactionFeeHandler
	economicsFee  process.FeeHandler
	txTypeHandler process.TxTypeHandler
	gasHandler    process.GasHandler
}

// NewSmartContractProcessor create a smart contract processor creates and interprets VM data
func NewSmartContractProcessor(
	vmContainer process.VirtualMachinesContainer,
	argsParser process.ArgumentsParser,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountsDB state.AccountsAdapter,
	tempAccounts process.TemporaryAccountsHandler,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
	scrForwarder process.IntermediateTransactionHandler,
	txFeeHandler process.TransactionFeeHandler,
	economicsFee process.FeeHandler,
	txTypeHandler process.TxTypeHandler,
	gasHandler process.GasHandler,
) (*scProcessor, error) {

	if check.IfNil(vmContainer) {
		return nil, process.ErrNoVM
	}
	if check.IfNil(argsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(accountsDB) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(tempAccounts) {
		return nil, process.ErrNilTemporaryAccountsHandler
	}
	if check.IfNil(adrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(scrForwarder) {
		return nil, process.ErrNilIntermediateTransactionHandler
	}
	if check.IfNil(txFeeHandler) {
		return nil, process.ErrNilUnsignedTxHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}

	sc := &scProcessor{
		vmContainer:      vmContainer,
		argsParser:       argsParser,
		hasher:           hasher,
		marshalizer:      marshalizer,
		accounts:         accountsDB,
		tempAccounts:     tempAccounts,
		adrConv:          adrConv,
		shardCoordinator: coordinator,
		scrForwarder:     scrForwarder,
		txFeeHandler:     txFeeHandler,
		economicsFee:     economicsFee,
		txTypeHandler:    txTypeHandler,
		gasHandler:       gasHandler,
	}

	err := sc.createBuiltInFunctions()
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func (sc *scProcessor) createBuiltInFunctions() error {
	sc.builtInFunctions = make(map[string]process.BuiltinFunction)

	sc.builtInFunctions[claimDeveloperRewardsFunctionName] = &claimDeveloperRewards{}
	sc.builtInFunctions[changeOwnerAddressFunctionName] = &changeOwnerAddress{}

	return nil
}

func (sc *scProcessor) checkTxValidity(tx data.TransactionHandler) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := sc.adrConv.AddressLen() != len(tx.GetRecvAddress())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

func (sc *scProcessor) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRecvAddress(), make([]byte, sc.adrConv.AddressLen()))
	return isEmptyAddress
}

// ExecuteSmartContractTransaction processes the transaction, call the VM and processes the SC call output
func (sc *scProcessor) ExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}
	if check.IfNil(acntDst) {
		return process.ErrNilSCDestAccount
	}

	err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		log.Debug("process sc payment error", "error", err.Error())
		return err
	}

	defer func() {
		if err != nil {
			err = sc.processIfError(acntSnd, tx, err.Error())
			if err != nil {
				log.Debug("error while processing error in smart contract processor")
			}

			err = sc.saveAccounts(acntSnd, nil)
			if err != nil {
				log.Debug("error saving account")
			}
		}
	}()

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		log.Debug("prepare smart contract call error", "error", err.Error())
		return nil
	}

	vmInput, err := sc.createVMCallInput(tx)
	if err != nil {
		log.Debug("create vm call input error", "error", err.Error())
		return nil
	}

	executed, err := sc.resolveBuiltInFunctions(tx, acntSnd, acntDst, vmInput)
	if err != nil {
		log.Debug("processed built in functions error", "error", err.Error())
		return nil
	}
	if executed {
		return nil
	}

	vm, err := sc.getVMFromRecvAddress(tx)
	if err != nil {
		log.Debug("get vm from address error", "error", err.Error())
		return nil
	}

	vmOutput, err := vm.RunSmartContractCall(vmInput)
	if err != nil {
		log.Debug("run smart contract call error", "error", err.Error())
		return nil
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		return err
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd)
	if err != nil {
		log.Trace("process vm output error", "error", err.Error())
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("AddIntermediateTransactions error", "error", err.Error())
		return nil
	}

	newDeveloperReward := core.GetPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())
	feeForValidators := big.NewInt(0).Sub(consumedFee, newDeveloperReward)

	acntDst, err = sc.reloadLocalAccount(acntDst)
	if err != nil {
		log.Debug("reloadLocalAccount error", "error", err.Error())
		return nil
	}

	acntDst.SetDeveloperReward(big.NewInt(0).Add(acntDst.GetDeveloperReward(), newDeveloperReward))

	sc.txFeeHandler.ProcessTransactionFee(feeForValidators)

	err = sc.saveAccounts(nil, acntDst)
	if err != nil {
		log.Debug("error saving account")
	}

	return nil
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
	tx data.TransactionHandler,
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (bool, error) {

	builtIn, ok := sc.builtInFunctions[vmInput.Function]
	if !ok {
		return false, nil
	}
	if check.IfNil(builtIn) {
		return true, process.ErrNilBuiltInFunction
	}

	valueToSend, err := builtIn.ProcessBuiltinFunction(tx, acntSnd, acntDst, vmInput)
	if err != nil {
		return true, err
	}

	err = sc.saveAccounts(acntSnd, acntDst)
	if err != nil {
		return true, err
	}

	txHash, err := sc.computeTransactionHash(tx)
	if err != nil {
		return true, err
	}

	gasConsumed := builtInFunctionBaseCostMultiplier * sc.economicsFee.ComputeGasLimit(tx)
	if tx.GetGasLimit() < gasConsumed {
		return true, process.ErrNotEnoughGas
	}

	acntSnd, err = sc.reloadLocalAccount(acntSnd)
	if err != nil {
		return true, err
	}

	gasRemaining := tx.GetGasLimit() - gasConsumed
	scrRefund, consumedFee, err := sc.createSCRForSender(
		big.NewInt(0),
		gasRemaining,
		vmcommon.Ok,
		make([][]byte, 0),
		tx,
		txHash,
		acntSnd,
	)
	if err != nil {
		return true, err
	}

	scrRefund.Value.Add(scrRefund.Value, valueToSend)
	err = sc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrRefund})
	if err != nil {
		log.Debug("AddIntermediateTransactions error", "error", err.Error())
		return true, err
	}

	sc.gasHandler.SetGasRefunded(gasRemaining, txHash)
	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	return true, nil
}

func (sc *scProcessor) processIfError(
	acntSnd state.UserAccountHandler,
	tx data.TransactionHandler,
	returnCode string,
) error {
	consumedFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GetGasLimit()), big.NewInt(0).SetUint64(tx.GetGasPrice()))
	scrIfError, err := sc.createSCRsWhenError(tx, returnCode)
	if err != nil {
		return err
	}

	if !check.IfNil(acntSnd) {
		acntSnd.SetBalance(big.NewInt(0).Add(acntSnd.GetBalance(), tx.GetValue()))
	}

	err = sc.scrForwarder.AddIntermediateTransactions(scrIfError)
	if err != nil {
		return err
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	return nil
}

func (sc *scProcessor) prepareSmartContractCall(tx data.TransactionHandler, acntSnd state.UserAccountHandler) error {
	sc.isCallBack = false
	dataToParse := tx.GetData()

	scr, ok := tx.(*smartContractResult.SmartContractResult)
	isSCRResultFromCrossShardCall := ok && len(scr.Data) > 0 && scr.Data[0] == '@'
	if isSCRResultFromCrossShardCall {
		dataToParse = append([]byte("callBack"), tx.GetData()...)
		sc.isCallBack = true
	}

	err := sc.argsParser.ParseData(string(dataToParse))
	if err != nil {
		return err
	}

	nonce := tx.GetNonce()
	if acntSnd != nil && !acntSnd.IsInterfaceNil() {
		nonce = acntSnd.GetNonce()
	}

	txValue := big.NewInt(0).Set(tx.GetValue())
	sc.tempAccounts.AddTempAccount(tx.GetSndAddress(), txValue, nonce)

	return nil
}

func (sc *scProcessor) getVMTypeFromArguments(vmType []byte) ([]byte, error) {
	// first parsed argument after the code in case of vmDeploy is the actual vmType
	vmAppendedType := make([]byte, core.VMTypeLen)
	vmArgLen := len(vmType)
	if vmArgLen > core.VMTypeLen {
		return nil, process.ErrVMTypeLengthInvalid
	}

	copy(vmAppendedType[core.VMTypeLen-vmArgLen:], vmType)
	return vmAppendedType, nil
}

func (sc *scProcessor) getVMFromRecvAddress(tx data.TransactionHandler) (vmcommon.VMExecutionHandler, error) {
	vmType := core.GetVMType(tx.GetRecvAddress())
	vm, err := sc.vmContainer.Get(vmType)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	err := sc.checkTxValidity(tx)
	if err != nil {
		log.Debug("Transaction invalid", "error", err.Error())
		return err
	}

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		log.Debug("Transaction wrong", "error", process.ErrWrongTransaction.Error())
		return process.ErrWrongTransaction
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = sc.processIfError(acntSnd, tx, err.Error())
			if err != nil {
				log.Debug("error while processing error in smart contract processor")
			}

			err = sc.saveAccounts(acntSnd, nil)
			if err != nil {
				log.Debug("error saving account")
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

	vmOutput, err := vm.RunSmartContractCreate(vmInput)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return nil
	}

	err = sc.saveAccounts(acntSnd, nil)
	if err != nil {
		return err
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd)
	if err != nil {
		log.Debug("Processing error", "error", err.Error())
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("AddIntermediate Transaction error", "error", err.Error())
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	log.Trace("SmartContract deployed")
	return nil
}

func (sc *scProcessor) createVMCallInput(tx data.TransactionHandler) (*vmcommon.ContractCallInput, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, err
	}

	vmCallInput := &vmcommon.ContractCallInput{}
	vmCallInput.VMInput = *vmInput
	vmCallInput.Function, err = sc.argsParser.GetFunction()
	if err != nil {
		return nil, err
	}

	vmCallInput.RecipientAddr = tx.GetRecvAddress()

	return vmCallInput, nil
}

func (sc *scProcessor) createVMDeployInput(
	tx data.TransactionHandler,
) (*vmcommon.ContractCreateInput, []byte, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, nil, err
	}

	if len(vmInput.Arguments) < 1 {
		return nil, nil, process.ErrNotEnoughArgumentsToDeploy
	}

	vmType, err := sc.getVMTypeFromArguments(vmInput.Arguments[0])
	if err != nil {
		return nil, nil, err
	}
	// delete the first argument as it is the vmType
	vmInput.Arguments = vmInput.Arguments[1:]

	vmCreateInput := &vmcommon.ContractCreateInput{}
	hexCode, err := sc.argsParser.GetCode()
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.ContractCode, err = hex.DecodeString(string(hexCode))
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.VMInput = *vmInput

	return vmCreateInput, vmType, nil
}

func (sc *scProcessor) createVMInput(tx data.TransactionHandler) (*vmcommon.VMInput, error) {
	var err error
	vmInput := &vmcommon.VMInput{}

	vmInput.CallerAddr = tx.GetSndAddress()
	vmInput.Arguments, err = sc.argsParser.GetArguments()
	if err != nil {
		return nil, err
	}

	vmInput.CallValue = tx.GetValue()
	vmInput.GasPrice = tx.GetGasPrice()
	moveBalanceGasConsume := sc.economicsFee.ComputeGasLimit(tx)

	if tx.GetGasLimit() < moveBalanceGasConsume {
		return nil, process.ErrNotEnoughGas
	}

	vmInput.GasProvided = tx.GetGasLimit() - moveBalanceGasConsume

	return vmInput, nil
}

// taking money from sender, as VM might not have access to him because of state sharding
func (sc *scProcessor) processSCPayment(tx data.TransactionHandler, acntSnd state.UserAccountHandler) error {
	if check.IfNil(acntSnd) {
		// transaction was already processed at sender shard
		return nil
	}

	acntSnd.SetNonce(acntSnd.GetNonce() + 1)

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

	if acntSnd.GetBalance().Cmp(cost) < 0 {
		return process.ErrInsufficientFunds
	}

	acntSnd.SetBalance(big.NewInt(0).Add(acntSnd.GetBalance(), big.NewInt(0).Neg(cost)))

	return nil
}

func (sc *scProcessor) computeTransactionHash(tx data.TransactionHandler) ([]byte, error) {
	scr, ok := tx.(*smartContractResult.SmartContractResult)
	if ok {
		return scr.TxHash, nil
	}

	return core.CalculateHash(sc.marshalizer, sc.hasher, tx)
}

func (sc *scProcessor) processVMOutput(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	acntSnd state.UserAccountHandler,
) ([]data.TransactionHandler, *big.Int, error) {
	if vmOutput == nil {
		return nil, nil, process.ErrNilVMOutput
	}
	if check.IfNil(tx) {
		return nil, nil, process.ErrNilTransaction
	}

	txHash, err := sc.computeTransactionHash(tx)
	if err != nil {
		return nil, nil, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Trace("smart contract processing returned with error",
			"hash", txHash,
			"return code", vmOutput.ReturnCode.String(),
		)

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

	totalGasConsumed := tx.GetGasLimit() - vmOutput.GasRemaining
	log.Trace("total gas consumed", "value", totalGasConsumed, "hash", txHash)

	if vmOutput.GasRefund.Cmp(big.NewInt(0)) > 0 {
		log.Trace("total gas refunded", "value", vmOutput.GasRefund.String(), "hash", txHash)
	}

	scrRefund, consumedFee, err := sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		tx,
		txHash,
		acntSnd,
	)
	if err != nil {
		return nil, nil, err
	}

	scrTxs = append(scrTxs, scrRefund)

	err = sc.saveAccounts(acntSnd, nil)
	if err != nil {
		return nil, nil, err
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, nil, err
	}

	err = sc.processTouchedAccounts(vmOutput.TouchedAccounts)
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
	tx data.TransactionHandler,
	returnCode string,
) ([]data.TransactionHandler, error) {
	txHash, err := core.CalculateHash(sc.marshalizer, sc.hasher, tx)
	if err != nil {
		return nil, err
	}

	rcvAddress := tx.GetSndAddress()
	if sc.isCallBack {
		rcvAddress = tx.GetRecvAddress()
	}

	scr := &smartContractResult.SmartContractResult{
		Nonce:   tx.GetNonce(),
		Value:   tx.GetValue(),
		RcvAddr: rcvAddress,
		SndAddr: tx.GetRecvAddress(),
		Code:    nil,
		Data:    []byte("@" + hex.EncodeToString([]byte(returnCode)) + "@" + hex.EncodeToString(txHash)),
		TxHash:  txHash,
	}

	resultedScrs := []data.TransactionHandler{scr}

	return resultedScrs, nil
}

// reloadLocalAccount will reload from current account state the sender account
// this requirement is needed because in the case of refunding the exact account that was previously
// modified in saveSCOutputToCurrentState, the modifications done there should be visible here
func (sc *scProcessor) reloadLocalAccount(acntSnd state.UserAccountHandler) (state.UserAccountHandler, error) {
	if check.IfNil(acntSnd) {
		return acntSnd, nil
	}

	isAccountFromCurrentShard := acntSnd.AddressContainer() != nil
	if !isAccountFromCurrentShard {
		return acntSnd, nil
	}

	return sc.getAccountFromAddress(acntSnd.AddressContainer().Bytes())
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
	result.SndAddr = tx.GetRecvAddress()
	result.Code = outAcc.Code
	result.Data = append(outAcc.Data, sc.argsParser.CreateDataFromStorageUpdate(storageUpdates)...)
	result.GasLimit = outAcc.GasLimit
	result.GasPrice = tx.GetGasPrice()
	result.TxHash = txHash

	return result
}

// createSCRForSender(vmOutput, tx, txHash, acntSnd)
// give back the user the unused gas money
func (sc *scProcessor) createSCRForSender(
	gasRefund *big.Int,
	gasRemaining uint64,
	returnCode vmcommon.ReturnCode,
	returnData [][]byte,
	tx data.TransactionHandler,
	txHash []byte,
	acntSnd state.UserAccountHandler,
) (*smartContractResult.SmartContractResult, *big.Int, error) {
	storageFreeRefund := big.NewInt(0).Mul(gasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))

	consumedFee := big.NewInt(0)
	consumedFee = consumedFee.Mul(big.NewInt(0).SetUint64(tx.GetGasPrice()), big.NewInt(0).SetUint64(tx.GetGasLimit()))

	refundErd := big.NewInt(0)
	refundErd = refundErd.Mul(big.NewInt(0).SetUint64(gasRemaining), big.NewInt(0).SetUint64(tx.GetGasPrice()))
	consumedFee = consumedFee.Sub(consumedFee, refundErd)

	rcvAddress := tx.GetSndAddress()
	if sc.isCallBack {
		rcvAddress = tx.GetRecvAddress()
	}

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = big.NewInt(0).Add(refundErd, storageFreeRefund)
	scTx.RcvAddr = rcvAddress
	scTx.SndAddr = tx.GetRecvAddress()
	scTx.Nonce = tx.GetNonce() + 1
	scTx.TxHash = txHash
	scTx.GasLimit = gasRemaining
	scTx.GasPrice = tx.GetGasPrice()

	scTx.Data = []byte("@" + hex.EncodeToString([]byte(returnCode.String())))
	for _, retData := range returnData {
		scTx.Data = append(scTx.Data, []byte("@"+hex.EncodeToString(retData))...)
	}

	if check.IfNil(acntSnd) {
		return scTx, consumedFee, nil
	}

	acntSnd.SetBalance(big.NewInt(0).Add(acntSnd.GetBalance(), scTx.Value))

	return scTx, consumedFee, nil
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(
	outputAccounts []*vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	scResults := make([]data.TransactionHandler, 0, len(outputAccounts))

	sumOfAllDiff := big.NewInt(0)
	sumOfAllDiff = sumOfAllDiff.Sub(sumOfAllDiff, tx.GetValue())

	zero := big.NewInt(0)
	for i := 0; i < len(outputAccounts); i++ {
		outAcc := outputAccounts[i]
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return nil, err
		}

		storageUpdates := getSortedStorageUpdates(outAcc)
		scTx := sc.createSmartContractResult(outAcc, tx, txHash, storageUpdates)
		scResults = append(scResults, scTx)

		if check.IfNil(acc) {
			if outAcc.BalanceDelta != nil {
				sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
			}
			continue
		}

		for j := 0; j < len(storageUpdates); j++ {
			storeUpdate := storageUpdates[j]
			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)

			log.Trace("storeUpdate", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
		}

		//if len(outAcc.StorageUpdates) > 0 {
		//	//SC with data variables
		//	err = sc.accounts.SaveDataTrie(acc)
		//	if err != nil {
		//		return nil, err
		//	}
		//}

		err = sc.updateSmartContractCode(acc, outAcc, tx)
		if err != nil {
			return nil, err
		}

		// change nonce only if there is a change
		if outAcc.Nonce != acc.GetNonce() && outAcc.Nonce != 0 {
			if outAcc.Nonce < acc.GetNonce() {
				return nil, process.ErrWrongNonceInVMOutput
			}

			acc.SetNonce(outAcc.Nonce)
		}

		err = sc.accounts.SaveAccount(acc)
		if err != nil {
			return nil, err
		}

		// if no change then continue
		if outAcc.BalanceDelta == nil || outAcc.BalanceDelta.Cmp(zero) == 0 {
			continue
		}

		sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)

		newBalance := big.NewInt(0).Add(acc.GetBalance(), outAcc.BalanceDelta)
		if newBalance.Cmp(big.NewInt(0)) < 0 {
			return nil, process.ErrInsufficientFunds
		}

		acc.SetBalance(newBalance)

		err = sc.saveAccounts(acc, nil)
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
	account state.UserAccountHandler,
	outAcc *vmcommon.OutputAccount,
	tx data.TransactionHandler,
) error {
	if len(outAcc.Code) == 0 {
		return nil
	}

	isDeployment := len(account.GetCode()) == 0
	if isDeployment {
		account.SetOwnerAddress(tx.GetSndAddress())
		account.SetCode(outAcc.Code)

		log.Trace("created SC address", "address", hex.EncodeToString(outAcc.Address))
		return nil
	}

	// TODO: implement upgradable flag for smart contracts
	isUpgradeEnabled := bytes.Equal(account.GetOwnerAddress(), tx.GetSndAddress())
	if isUpgradeEnabled {
		account.SetCode(outAcc.Code)

		log.Trace("created SC address", "address", hex.EncodeToString(outAcc.Address))
		return nil
	}

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

		if acc == nil || acc.IsInterfaceNil() {
			//TODO: sharded Smart Contract processing
			continue
		}

		err = sc.accounts.RemoveAccount(acc.AddressContainer())
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *scProcessor) processTouchedAccounts(_ [][]byte) error {
	//TODO: implement
	return nil
}

func (sc *scProcessor) getAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	adrSrc, err := sc.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := sc.shardCoordinator.SelfId()
	shardForSrc := sc.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := sc.accounts.LoadAccount(adrSrc)
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

	var err error
	defer func() {
		if err != nil {
			err = sc.processIfError(nil, scr, err.Error())
			if err != nil {
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

	process.DisplayProcessTxDetails("ProcessSmartContractResult: receiver account details", dstAcc, scr)

	txType, err := sc.txTypeHandler.ComputeTransactionType(scr)
	if err != nil {
		return nil
	}

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
	}

	err = process.ErrWrongTransaction
	return nil
}

func (sc *scProcessor) processSimpleSCR(
	scr *smartContractResult.SmartContractResult,
	dstAcc state.UserAccountHandler,
) error {
	if len(scr.Data) > 0 {
		storageUpdates, err := sc.argsParser.GetStorageUpdates(string(scr.Data))
		if err != nil {
			log.Debug("storage updates could not be parsed")
		}

		for i := 0; i < len(storageUpdates); i++ {
			dstAcc.DataTrieTracker().SaveKeyValue(storageUpdates[i].Offset, storageUpdates[i].Data)
		}

		////SC with data variables
		//err = sc.accounts.SaveDataTrie(dstAcc)
		//if err != nil {
		//	return err
		//}
	}

	err := sc.updateSmartContractCode(dstAcc, &vmcommon.OutputAccount{Code: scr.Code, Address: scr.RcvAddr}, scr)
	if err != nil {
		return err
	}

	if scr.Value == nil {
		return nil
	}

	newBalance := big.NewInt(0).Add(dstAcc.GetBalance(), scr.Value)
	if newBalance.Cmp(big.NewInt(0)) < 0 {
		return process.ErrInsufficientFunds
	}

	dstAcc.SetBalance(newBalance)
	err = sc.saveAccounts(dstAcc, nil)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	return sc == nil
}
