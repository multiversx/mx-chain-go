package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/feeTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-vm-common"
)

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

	mutSCState   sync.Mutex
	mapExecState map[uint64]scExecutionState

	scrForwarder process.IntermediateTransactionHandler
	txFeeHandler process.UnsignedTxHandler
}

var log = logger.DefaultLogger()

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
	txFeeHandler process.UnsignedTxHandler,
) (*scProcessor, error) {
	if vmContainer == nil || vmContainer.IsInterfaceNil() {
		return nil, process.ErrNoVM
	}
	if argsParser == nil || argsParser.IsInterfaceNil() {
		return nil, process.ErrNilArgumentParser
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if accountsDB == nil || accountsDB.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if tempAccounts == nil || tempAccounts.IsInterfaceNil() {
		return nil, process.ErrNilTemporaryAccountsHandler
	}
	if adrConv == nil || adrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if coordinator == nil || coordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if scrForwarder == nil || scrForwarder.IsInterfaceNil() {
		return nil, process.ErrNilIntermediateTransactionHandler
	}
	if txFeeHandler == nil {
		return nil, process.ErrNilUnsignedTxHandler
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
		mapExecState:     make(map[uint64]scExecutionState)}, nil
}

// ComputeTransactionType calculates the type of the transaction
func (sc *scProcessor) ComputeTransactionType(tx *transaction.Transaction) (process.TransactionType, error) {
	err := sc.checkTxValidity(tx)
	if err != nil {
		return 0, err
	}

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.Data) > 0 {
			return process.SCDeployment, nil
		}
		return 0, process.ErrWrongTransaction
	}

	acntDst, err := sc.getAccountFromAddress(tx.RcvAddr)
	if err != nil {
		return 0, err
	}

	if acntDst == nil || acntDst.IsInterfaceNil() {
		return process.MoveBalance, nil
	}

	if !acntDst.IsInterfaceNil() && len(acntDst.GetCode()) > 0 {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

func (sc *scProcessor) checkTxValidity(tx *transaction.Transaction) error {
	if tx == nil || tx.IsInterfaceNil() {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := sc.adrConv.AddressLen() != len(tx.RcvAddr)
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

func (sc *scProcessor) isDestAddressEmpty(tx *transaction.Transaction) bool {
	isEmptyAddress := bytes.Equal(tx.RcvAddr, make([]byte, sc.adrConv.AddressLen()))
	return isEmptyAddress
}

// ExecuteSmartContractTransaction processes the transaction, call the VM and processes the SC call output
func (sc *scProcessor) ExecuteSmartContractTransaction(
	tx *transaction.Transaction,
	acntSnd, acntDst state.AccountHandler,
	round uint64,
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

	err := sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		return err
	}

	vmInput, err := sc.createVMCallInput(tx)
	if err != nil {
		return err
	}

	vm, err := sc.getVMFromTransaction(tx)
	if err != nil {
		return err
	}

	vmOutput, err := vm.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	// VM is formally verified and the output is correct
	crossTxs, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd, round)
	if err != nil {
		return err
	}

	err = sc.scrForwarder.AddIntermediateTransactions(crossTxs)
	if err != nil {
		return err
	}

	sc.txFeeHandler.AddProcessedUTx(consumedFee)

	return nil
}

func (sc *scProcessor) prepareSmartContractCall(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	err := sc.argsParser.ParseData(tx.Data)
	if err != nil {
		return err
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	nonce := tx.Nonce
	if acntSnd != nil && !acntSnd.IsInterfaceNil() {
		nonce = acntSnd.GetNonce()
	}
	txValue := big.NewInt(0).Set(tx.Value)
	sc.tempAccounts.AddTempAccount(tx.SndAddr, txValue, nonce)

	return nil
}

func (sc *scProcessor) getVMFromTransaction(tx *transaction.Transaction) (vmcommon.VMExecutionHandler, error) {
	//TODO add processing here - like calculating what kind of VM does this contract call needs
	vm, err := sc.vmContainer.Get([]byte(factory.IELEVirtualMachine))
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(
	tx *transaction.Transaction,
	acntSnd state.AccountHandler,
	round uint64,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	err := sc.checkTxValidity(tx)
	if err != nil {
		return err
	}

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		return process.ErrWrongTransaction
	}

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		return err
	}

	vmInput, err := sc.createVMDeployInput(tx)
	if err != nil {
		return err
	}

	vm, err := sc.getVMFromTransaction(tx)
	if err != nil {
		return err
	}

	// TODO: Smart contract address calculation
	vmOutput, err := vm.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}

	// VM is formally verified, the output is correct
	crossTxs, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd, round)
	if err != nil {
		return err
	}

	err = sc.scrForwarder.AddIntermediateTransactions(crossTxs)
	if err != nil {
		return err
	}

	sc.txFeeHandler.AddProcessedUTx(consumedFee)

	return nil
}

func (sc *scProcessor) createVMCallInput(tx *transaction.Transaction) (*vmcommon.ContractCallInput, error) {
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

	vmCallInput.RecipientAddr = tx.RcvAddr

	return vmCallInput, nil
}

func (sc *scProcessor) createVMDeployInput(tx *transaction.Transaction) (*vmcommon.ContractCreateInput, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, err
	}

	vmCreateInput := &vmcommon.ContractCreateInput{}
	hexCode, err := sc.argsParser.GetCode()
	if err != nil {
		return nil, err
	}

	vmCreateInput.ContractCode, err = hex.DecodeString(string(hexCode))
	if err != nil {
		return nil, err
	}

	vmCreateInput.VMInput = *vmInput

	return vmCreateInput, nil
}

func (sc *scProcessor) createVMInput(tx *transaction.Transaction) (*vmcommon.VMInput, error) {
	var err error
	vmInput := &vmcommon.VMInput{}

	vmInput.CallerAddr = tx.SndAddr
	vmInput.Arguments, err = sc.argsParser.GetArguments()
	if err != nil {
		return nil, err
	}
	vmInput.CallValue = tx.Value
	vmInput.GasPrice = big.NewInt(int64(tx.GasPrice))
	vmInput.GasProvided = big.NewInt(int64(tx.GasLimit))

	//TODO: change this when we know for what they are used.
	scCallHeader := &vmcommon.SCCallHeader{}
	scCallHeader.GasLimit = big.NewInt(0)
	scCallHeader.Number = big.NewInt(0)
	scCallHeader.Timestamp = big.NewInt(0)
	scCallHeader.Beneficiary = big.NewInt(0)

	vmInput.Header = scCallHeader

	return vmInput, nil
}

// taking money from sender, as VM might not have access to him because of state sharding
func (sc *scProcessor) processSCPayment(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		// transaction was already done at sender shard
		return nil
	}

	err := acntSnd.SetNonceWithJournal(acntSnd.GetNonce() + 1)
	if err != nil {
		return err
	}

	cost := big.NewInt(0)
	cost = cost.Mul(big.NewInt(0).SetUint64(tx.GasPrice), big.NewInt(0).SetUint64(tx.GasLimit))
	cost = cost.Add(cost, tx.Value)

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

func (sc *scProcessor) processVMOutput(
	vmOutput *vmcommon.VMOutput,
	tx *transaction.Transaction,
	acntSnd state.AccountHandler,
	round uint64,
) ([]data.TransactionHandler, *feeTx.FeeTx, error) {
	if vmOutput == nil {
		return nil, nil, process.ErrNilVMOutput
	}
	if tx == nil {
		return nil, nil, process.ErrNilTransaction
	}

	txBytes, err := sc.marshalizer.Marshal(tx)
	if err != nil {
		return nil, nil, err
	}
	txHash := sc.hasher.Compute(string(txBytes))

	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Info(fmt.Sprintf(
			"error processing tx %s in VM: return code: %s",
			hex.EncodeToString(txHash),
			vmOutput.ReturnCode),
		)
	}

	err = sc.saveSCOutputToCurrentState(vmOutput, round, txHash)
	if err != nil {
		return nil, nil, err
	}

	crossOutAccs, err := sc.processSCOutputAccounts(vmOutput.OutputAccounts)
	if err != nil {
		return nil, nil, err
	}

	crossTxs, err := sc.createCrossShardTransactions(crossOutAccs, tx, txHash)
	if err != nil {
		return nil, nil, err
	}

	acntSnd, err = sc.reloadLocalSndAccount(acntSnd)
	if err != nil {
		return nil, nil, err
	}

	totalGasRefund := big.NewInt(0)
	totalGasRefund = totalGasRefund.Add(vmOutput.GasRefund, vmOutput.GasRemaining)
	scrIfCrossShard, consumedFee, err := sc.refundGasToSender(totalGasRefund, tx, txHash, acntSnd)
	if err != nil {
		return nil, nil, err
	}

	if scrIfCrossShard != nil {
		crossTxs = append(crossTxs, scrIfCrossShard)
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, nil, err
	}

	err = sc.processTouchedAccounts(vmOutput.TouchedAccounts)
	if err != nil {
		return nil, nil, err
	}

	currFeeTx := &feeTx.FeeTx{
		Nonce: tx.Nonce,
		Value: consumedFee,
	}

	return crossTxs, currFeeTx, nil
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
	scAddress []byte,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	crossSc := &smartContractResult.SmartContractResult{}

	crossSc.Value = outAcc.Balance
	crossSc.Nonce = outAcc.Nonce.Uint64()
	crossSc.RcvAddr = outAcc.Address
	crossSc.SndAddr = scAddress
	crossSc.Code = outAcc.Code
	crossSc.Data = sc.argsParser.CreateDataFromStorageUpdate(outAcc.StorageUpdates)
	crossSc.TxHash = txHash

	return crossSc
}

func (sc *scProcessor) createCrossShardTransactions(
	crossOutAccs []*vmcommon.OutputAccount,
	tx *transaction.Transaction,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	crossSCTxs := make([]data.TransactionHandler, 0)

	for i := 0; i < len(crossOutAccs); i++ {
		scTx := sc.createSmartContractResult(crossOutAccs[i], tx.RcvAddr, txHash)
		crossSCTxs = append(crossSCTxs, scTx)
	}

	return crossSCTxs, nil
}

// give back the user the unused gas money
func (sc *scProcessor) refundGasToSender(
	gasRefund *big.Int,
	tx *transaction.Transaction,
	txHash []byte,
	acntSnd state.AccountHandler,
) (*smartContractResult.SmartContractResult, *big.Int, error) {
	consumedFee := big.NewInt(0)
	consumedFee = consumedFee.Mul(big.NewInt(0).SetUint64(tx.GasPrice), big.NewInt(0).SetUint64(tx.GasLimit))
	if gasRefund == nil || gasRefund.Cmp(big.NewInt(0)) <= 0 {
		return nil, consumedFee, nil
	}

	refundErd := big.NewInt(0)
	refundErd = refundErd.Mul(gasRefund, big.NewInt(int64(tx.GasPrice)))
	consumedFee = consumedFee.Sub(consumedFee, refundErd)

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = refundErd
	scTx.RcvAddr = tx.SndAddr
	scTx.SndAddr = tx.RcvAddr
	scTx.Nonce = tx.Nonce + 1
	scTx.TxHash = txHash

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
func (sc *scProcessor) processSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount) ([]*vmcommon.OutputAccount, error) {
	crossOutAccs := make([]*vmcommon.OutputAccount, 0)
	for i := 0; i < len(outputAccounts); i++ {
		outAcc := outputAccounts[i]
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return nil, err
		}

		fakeAcc := sc.tempAccounts.TempAccount(outAcc.Address)

		if acc == nil || acc.IsInterfaceNil() {
			crossOutAccs = append(crossOutAccs, outAcc)
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
				return nil, err
			}
		}

		if len(outAcc.Code) > 0 {
			err = sc.accounts.PutCode(acc, outAcc.Code)
			if err != nil {
				return nil, err
			}

			//TODO remove this when receipts are implemented
			log.Info(fmt.Sprintf("*** Generated/called SC account: %s ***", hex.EncodeToString(outAcc.Address)))
		}

		if outAcc.Nonce == nil || outAcc.Nonce.Cmp(big.NewInt(int64(acc.GetNonce()))) < 0 {
			return nil, process.ErrWrongNonceInVMOutput
		}

		err = acc.SetNonceWithJournal(outAcc.Nonce.Uint64())
		if err != nil {
			return nil, err
		}

		if outAcc.Balance == nil {
			return nil, process.ErrNilBalanceFromSC
		}

		stAcc, ok := acc.(*state.Account)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		// if fake account, than VM only has transaction value as balance, so anything remaining is a plus
		if fakeAcc != nil && !fakeAcc.IsInterfaceNil() {
			outAcc.Balance = outAcc.Balance.Add(outAcc.Balance, stAcc.Balance)
		}

		realBalanceChange := big.NewInt(0).Sub(outAcc.Balance, stAcc.Balance)

		// update the values according to SC output
		err = stAcc.SetBalanceWithJournal(outAcc.Balance)
		if err != nil {
			return nil, err
		}

		zero := big.NewInt(0)
		if realBalanceChange.Cmp(zero) != 0 {
			outAcc.Balance = realBalanceChange
			crossOutAccs = append(crossOutAccs, outAcc)
		}
	}

	return crossOutAccs, nil
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

func (sc *scProcessor) processTouchedAccounts(touchedAccounts [][]byte) error {
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

// GetAllSmartContractCallRootHash returns the roothash of the state of the SC executions for defined round
func (sc *scProcessor) GetAllSmartContractCallRootHash(round uint64) []byte {
	return []byte("roothash")
}

// saves VM output into state
func (sc *scProcessor) saveSCOutputToCurrentState(output *vmcommon.VMOutput, round uint64, txHash []byte) error {
	var err error

	sc.mutSCState.Lock()
	defer sc.mutSCState.Unlock()

	if _, ok := sc.mapExecState[round]; !ok {
		sc.mapExecState[round] = scExecutionState{
			allLogs:       make(map[string][]*vmcommon.LogEntry),
			allReturnData: make(map[string][]*big.Int),
			returnCodes:   make(map[string]vmcommon.ReturnCode)}
	}

	tmpCurrScState := sc.mapExecState[round]
	defer func() {
		if err != nil {
			sc.mapExecState[round] = tmpCurrScState
		}
	}()

	err = sc.saveReturnData(output.ReturnData, round, txHash)
	if err != nil {
		return err
	}

	err = sc.saveReturnCode(output.ReturnCode, round, txHash)
	if err != nil {
		return err
	}

	err = sc.saveLogsIntoState(output.Logs, round, txHash)
	if err != nil {
		return err
	}

	return nil
}

// saves return data into account state
func (sc *scProcessor) saveReturnData(returnData []*big.Int, round uint64, txHash []byte) error {
	sc.mapExecState[round].allReturnData[string(txHash)] = returnData
	return nil
}

// saves smart contract return code into account state
func (sc *scProcessor) saveReturnCode(returnCode vmcommon.ReturnCode, round uint64, txHash []byte) error {
	sc.mapExecState[round].returnCodes[string(txHash)] = returnCode
	return nil
}

// save vm output logs into accounts
func (sc *scProcessor) saveLogsIntoState(logs []*vmcommon.LogEntry, round uint64, txHash []byte) error {
	sc.mapExecState[round].allLogs[string(txHash)] = logs
	return nil
}

// ProcessSmartContractResult updates the account state from the smart contract result
func (sc *scProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	if scr == nil {
		return process.ErrNilSmartContractResult
	}

	accHandler, err := sc.getAccountFromAddress(scr.RcvAddr)
	if err != nil {
		return err
	}
	if accHandler == nil || accHandler.IsInterfaceNil() {
		return process.ErrNilSCDestAccount
	}

	stAcc, ok := accHandler.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	storageUpdates, err := sc.argsParser.GetStorageUpdates(scr.Data)
	for i := 0; i < len(storageUpdates); i++ {
		stAcc.DataTrieTracker().SaveKeyValue(storageUpdates[i].Offset, storageUpdates[i].Data)
	}

	if len(scr.Data) > 0 {
		//SC with data variables
		err := sc.accounts.SaveDataTrie(stAcc)
		if err != nil {
			return err
		}
	}

	if len(scr.Code) > 0 {
		err = sc.accounts.PutCode(stAcc, scr.Code)
		if err != nil {
			return err
		}
	}

	if scr.Value == nil {
		return process.ErrNilBalanceFromSC
	}

	operation := big.NewInt(0)
	operation = operation.Add(scr.Value, stAcc.Balance)
	err = stAcc.SetBalanceWithJournal(operation)
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
