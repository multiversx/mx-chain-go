package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

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

var log = logger.GetOrCreate("process/smartcontract")

type scExecutionState struct {
	allLogs       map[string][]*vmcommon.LogEntry
	allReturnData map[string][]*big.Int
	returnCodes   map[string]vmcommon.ReturnCode
	rootHash      []byte
}

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

	mutSCState   sync.Mutex
	mapExecState map[uint64]scExecutionState

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

	return &scProcessor{
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
		mapExecState:     make(map[uint64]scExecutionState)}, nil
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
	acntSnd, acntDst state.AccountHandler,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	if tx == nil || tx.IsInterfaceNil() {
		return process.ErrNilTransaction
	}
	if acntDst == nil || acntDst.IsInterfaceNil() {
		return process.ErrNilSCDestAccount
	}
	if acntDst.IsInterfaceNil() || acntDst.GetCode() == nil {
		return process.ErrNilSCDestAccount
	}

	err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = sc.processIfError(acntSnd, tx, err.Error())
			if err != nil {
				log.Debug("error while processing error in smart contract processor")
			}
		}
	}()

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		return nil
	}

	vmInput, err := sc.createVMCallInput(tx)
	if err != nil {
		return nil
	}

	vm, err := sc.getVMFromRecvAddress(tx)
	if err != nil {
		return nil
	}

	vmOutput, err := vm.RunSmartContractCall(vmInput)
	if err != nil {
		return nil
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd)
	if err != nil {
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	return nil
}

func (sc *scProcessor) processIfError(
	acntSnd state.AccountHandler,
	tx data.TransactionHandler,
	returnCode string,
) error {
	consumedFee := big.NewInt(0).SetUint64(tx.GetGasLimit() * tx.GetGasPrice())
	scrIfError, err := sc.createSCRsWhenError(tx, returnCode)
	if err != nil {
		return err
	}

	if !check.IfNil(acntSnd) {
		stAcc, ok := acntSnd.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		totalCost := big.NewInt(0)
		err = stAcc.SetBalanceWithJournal(totalCost.Add(stAcc.Balance, tx.GetValue()))
		if err != nil {
			return err
		}
	}

	err = sc.scrForwarder.AddIntermediateTransactions(scrIfError)
	if err != nil {
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	return nil
}

func (sc *scProcessor) prepareSmartContractCall(tx data.TransactionHandler, acntSnd state.AccountHandler) error {
	sc.isCallBack = false
	dataToParse := tx.GetData()

	scr, ok := tx.(*smartContractResult.SmartContractResult)
	isSCRResultFromCrossShardCall := ok && len(scr.Data) > 0 && scr.Data[0] == '@'
	if isSCRResultFromCrossShardCall {
		dataToParse = "callBack" + tx.GetData()
		sc.isCallBack = true
	}

	err := sc.argsParser.ParseData(dataToParse)
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
	acntSnd state.AccountHandler,
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

	results, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd)
	if err != nil {
		log.Debug("Processing error", "error", err.Error())
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("Processing error", "error", err.Error())
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
func (sc *scProcessor) processSCPayment(tx data.TransactionHandler, acntSnd state.AccountHandler) error {
	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		// transaction was already processed at sender shard
		return nil
	}

	err := acntSnd.SetNonceWithJournal(acntSnd.GetNonce() + 1)
	if err != nil {
		return err
	}

	err = sc.economicsFee.CheckValidityTxValues(tx)
	if err != nil {
		return err
	}

	cost := big.NewInt(0)
	cost = cost.Mul(big.NewInt(0).SetUint64(tx.GetGasPrice()), big.NewInt(0).SetUint64(tx.GetGasLimit()))
	cost = cost.Add(cost, tx.GetValue())

	if cost.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if stAcc.Balance.Cmp(cost) < 0 {
		return process.ErrInsufficientFunds
	}

	totalCost := big.NewInt(0)
	err = stAcc.SetBalanceWithJournal(totalCost.Sub(stAcc.Balance, cost))
	if err != nil {
		return err
	}

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
	acntSnd state.AccountHandler,
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
		log.Debug("smart contract processing returned with error",
			"hash", txHash,
			"return code", vmOutput.ReturnCode.String(),
		)

		return nil, nil, fmt.Errorf(vmOutput.ReturnCode.String())
	}

	err = sc.processSCOutputAccounts(vmOutput.OutputAccounts, tx)
	if err != nil {
		return nil, nil, err
	}

	scrTxs, err := sc.createSCRTransactions(vmOutput.OutputAccounts, tx, txHash)
	if err != nil {
		return nil, nil, err
	}

	acntSnd, err = sc.reloadLocalSndAccount(acntSnd)
	if err != nil {
		return nil, nil, err
	}

	totalGasConsumed := tx.GetGasLimit() - vmOutput.GasRemaining
	log.Debug("total gas consumed", "value", totalGasConsumed, "hash", txHash)

	if vmOutput.GasRefund.Uint64() > 0 {
		log.Debug("total gas refunded", "value", vmOutput.GasRefund.Uint64(), "hash", txHash)
	}

	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)
	scrRefund, consumedFee, err := sc.createSCRForSender(vmOutput, tx, txHash, acntSnd)
	if err != nil {
		return nil, nil, err
	}

	if scrRefund != nil {
		scrTxs = append(scrTxs, scrRefund)
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, nil, err
	}

	err = sc.processTouchedAccounts(vmOutput.TouchedAccounts)
	if err != nil {
		return nil, nil, err
	}

	//TODO: think about if merge is needed between the resulted SCRs and if whether it does not complicates stuff
	// when smart contract processes in the second shard

	return scrTxs, consumedFee, nil
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
		Data:    "@" + hex.EncodeToString([]byte(returnCode)) + "@" + hex.EncodeToString(txHash),
		TxHash:  txHash,
	}

	resultedScrs := make([]data.TransactionHandler, 0)
	resultedScrs = append(resultedScrs, scr)

	return resultedScrs, nil
}

// reloadLocalSndAccount will reload from current account state the sender account
// this requirement is needed because in the case of refunding the exact account that was previously
// modified in saveSCOutputToCurrentState, the modifications done there should be visible here
func (sc *scProcessor) reloadLocalSndAccount(acntSnd state.AccountHandler) (state.AccountHandler, error) {
	if acntSnd == nil || acntSnd.IsInterfaceNil() {
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
) *smartContractResult.SmartContractResult {
	result := &smartContractResult.SmartContractResult{}

	result.Value = outAcc.BalanceDelta
	result.Nonce = outAcc.Nonce
	result.RcvAddr = outAcc.Address
	result.SndAddr = tx.GetRecvAddress()
	result.Code = outAcc.Code
	result.Data = string(outAcc.Data) + sc.argsParser.CreateDataFromStorageUpdate(outAcc.StorageUpdates)
	result.GasLimit = outAcc.GasLimit
	result.GasPrice = tx.GetGasPrice()
	result.TxHash = txHash

	return result
}

func (sc *scProcessor) createSCRTransactions(
	outAccs []*vmcommon.OutputAccount,
	tx data.TransactionHandler,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	scResults := make([]data.TransactionHandler, 0)

	for i := 0; i < len(outAccs); i++ {
		scTx := sc.createSmartContractResult(outAccs[i], tx, txHash)
		scResults = append(scResults, scTx)
	}

	return scResults, nil
}

// createSCRForSender(vmOutput, tx, txHash, acntSnd)
// give back the user the unused gas money
func (sc *scProcessor) createSCRForSender(
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte,
	acntSnd state.AccountHandler,
) (*smartContractResult.SmartContractResult, *big.Int, error) {
	gasRefund := big.NewInt(0).Add(vmOutput.GasRefund, big.NewInt(0).SetUint64(vmOutput.GasRemaining))

	consumedFee := big.NewInt(0)
	consumedFee = consumedFee.Mul(big.NewInt(0).SetUint64(tx.GetGasPrice()), big.NewInt(0).SetUint64(tx.GetGasLimit()))

	refundErd := big.NewInt(0)
	refundErd = refundErd.Mul(gasRefund, big.NewInt(int64(tx.GetGasPrice())))
	consumedFee = consumedFee.Sub(consumedFee, refundErd)

	rcvAddress := tx.GetSndAddress()
	if sc.isCallBack {
		rcvAddress = tx.GetRecvAddress()
	}

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = refundErd
	scTx.RcvAddr = rcvAddress
	scTx.SndAddr = tx.GetRecvAddress()
	scTx.Nonce = tx.GetNonce() + 1
	scTx.TxHash = txHash
	scTx.GasLimit = vmOutput.GasRemaining
	scTx.GasPrice = tx.GetGasPrice()

	scTx.Data = "@" + hex.EncodeToString([]byte(vmOutput.ReturnCode.String()))
	for _, retData := range vmOutput.ReturnData {
		scTx.Data += "@" + hex.EncodeToString(retData)
	}

	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		return scTx, consumedFee, nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	newBalance := big.NewInt(0).Add(stAcc.Balance, refundErd)
	err := stAcc.SetBalanceWithJournal(newBalance)
	if err != nil {
		return nil, nil, err
	}

	return scTx, consumedFee, nil
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount, tx data.TransactionHandler) error {
	sumOfAllDiff := big.NewInt(0)
	sumOfAllDiff = sumOfAllDiff.Sub(sumOfAllDiff, tx.GetValue())

	zero := big.NewInt(0)
	for i := 0; i < len(outputAccounts); i++ {
		outAcc := outputAccounts[i]
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return err
		}

		if acc == nil || acc.IsInterfaceNil() {
			if outAcc.BalanceDelta != nil {
				sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
			}
			continue
		}

		for j := 0; j < len(outAcc.StorageUpdates); j++ {
			storeUpdate := outAcc.StorageUpdates[j]
			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		if len(outAcc.StorageUpdates) > 0 {
			//SC with data variables
			err := sc.accounts.SaveDataTrie(acc)
			if err != nil {
				return err
			}
		}

		// change code if there is a change
		if len(outAcc.Code) > 0 {
			err = sc.accounts.PutCode(acc, outAcc.Code)
			if err != nil {
				return err
			}

			log.Debug("created SC address", "address", hex.EncodeToString(outAcc.Address))
		}

		// change nonce only if there is a change
		if outAcc.Nonce != acc.GetNonce() && outAcc.Nonce != 0 {
			if outAcc.Nonce < acc.GetNonce() {
				return process.ErrWrongNonceInVMOutput
			}

			err = acc.SetNonceWithJournal(outAcc.Nonce)
			if err != nil {
				return err
			}
		}

		// if no change then continue
		if outAcc.BalanceDelta == nil || outAcc.BalanceDelta.Cmp(zero) == 0 {
			continue
		}

		stAcc, ok := acc.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)

		// update the values according to SC output
		updatedBalance := big.NewInt(0)
		updatedBalance = updatedBalance.Add(stAcc.Balance, outAcc.BalanceDelta)
		if updatedBalance.Cmp(big.NewInt(0)) < 0 {
			return process.ErrOverallBalanceChangeFromSC
		}

		err = stAcc.SetBalanceWithJournal(updatedBalance)
		if err != nil {
			return err
		}
	}

	if sumOfAllDiff.Cmp(zero) != 0 {
		return process.ErrOverallBalanceChangeFromSC
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

func (sc *scProcessor) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
	adrSrc, err := sc.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := sc.shardCoordinator.SelfId()
	shardForSrc := sc.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := sc.accounts.GetAccountWithJournal(adrSrc)
	if err != nil {
		return nil, err
	}

	return acnt, nil
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
	if dstAcc == nil || dstAcc.IsInterfaceNil() {
		err = process.ErrNilSCDestAccount
		return nil
	}

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
	dstAcc state.AccountHandler,
) error {
	stAcc, ok := dstAcc.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if len(scr.Data) > 0 {
		storageUpdates, err := sc.argsParser.GetStorageUpdates(scr.Data)
		if err != nil {
			log.Debug("storage updates could not be parsed")
		}

		for i := 0; i < len(storageUpdates); i++ {
			stAcc.DataTrieTracker().SaveKeyValue(storageUpdates[i].Offset, storageUpdates[i].Data)
		}

		//SC with data variables
		err = sc.accounts.SaveDataTrie(stAcc)
		if err != nil {
			return err
		}
	}

	if len(scr.Code) > 0 {
		err := sc.accounts.PutCode(stAcc, scr.Code)
		if err != nil {
			return err
		}
	}

	if scr.Value == nil {
		return nil
	}

	operation := big.NewInt(0)
	operation = operation.Add(scr.Value, stAcc.Balance)
	err := stAcc.SetBalanceWithJournal(operation)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	if sc == nil {
		return true
	}
	return false
}
