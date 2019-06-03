package smartContract

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-vm-common"
)

type scExecutionState struct {
	mVMOutput map[string]*vmcommon.VMOutput
	rootHash  []byte
}

type scProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	vm               vmcommon.VMExecutionHandler
	argsParser       process.ArgumentsParser

	mutSCState    sync.Mutex
	currExecState scExecutionState
}

var log = logger.DefaultLogger()

// NewSmartContractProcessor create a smart contract processor creates and interprets VM data
func NewSmartContractProcessor(
	vm vmcommon.VMExecutionHandler,
	argsParser process.ArgumentsParser,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountsDB state.AccountsAdapter,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
) (*scProcessor, error) {
	if vm == nil {
		return nil, process.ErrNoVM
	}
	if argsParser == nil {
		return nil, process.ErrNilArgumentParser
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if accountsDB == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if adrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if coordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	return &scProcessor{
		vm:               vm,
		argsParser:       argsParser,
		hasher:           hasher,
		marshalizer:      marshalizer,
		accounts:         accountsDB,
		adrConv:          adrConv,
		shardCoordinator: coordinator}, nil
}

// ComputeTransactionType calculates the type of the transaction
func (sc *scProcessor) ComputeTransactionType(
	tx *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
) (process.TransactionType, error) {
	//TODO: add all kind of tests here
	if tx == nil {
		return 0, process.ErrNilTransaction
	}

	if len(tx.RcvAddr) == 0 {
		if len(tx.Data) > 0 {
			return process.SCDeployment, nil
		}
		return 0, process.ErrWrongTransaction
	}

	if acntDst == nil {
		return process.MoveBalance, nil
	}

	if !acntDst.IsInterfaceNil() && len(acntDst.GetCode()) > 0 {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

// ExecuteSmartContractTransaction processes the transaction, call the VM and processes the SC call output
func (sc *scProcessor) ExecuteSmartContractTransaction(
	tx *transaction.Transaction,
	acntSnd, acntDst state.AccountHandler,
) error {
	if sc.vm == nil {
		return process.ErrNoVM
	}
	if tx == nil {
		return process.ErrNilTransaction
	}
	if acntDst == nil {
		return process.ErrNilSCDestAccount
	}
	if acntDst.IsInterfaceNil() || acntDst.GetCode() == nil {
		return process.ErrNilSCDestAccount
	}

	err := sc.argsParser.ParseData(tx.Data)
	if err != nil {
		return err
	}

	vmInput, err := sc.createVMCallInput(tx)
	if err != nil {
		return err
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	err = sc.moveBalances(acntSnd, acntDst, tx.Value)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			errToLog := sc.moveBalances(acntDst, acntSnd, tx.Value)
			if errToLog != nil {
				log.Debug(errToLog.Error())
			}
		}
	}()

	vmOutput, err := sc.vm.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	err = sc.processVMOutput(vmOutput, tx, acntSnd, acntDst)
	if err != nil {
		return err
	}

	return nil
}

// DeploySmartContract runs the VM to verify register the smart contract and saves the data into the state
func (sc *scProcessor) DeploySmartContract(tx *transaction.Transaction, acntSnd, acntDst state.AccountHandler) error {
	if sc.vm == nil {
		return process.ErrNoVM
	}
	if len(tx.RcvAddr) != 0 {
		return process.ErrWrongTransaction
	}

	err := sc.argsParser.ParseData(tx.Data)
	if err != nil {
		return err
	}

	vmInput, err := sc.createVMDeployInput(tx)
	if err != nil {
		return err
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	vmOutput, err := sc.vm.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}

	err = sc.processVMOutput(vmOutput, tx, acntSnd, acntDst)
	if err != nil {
		return err
	}

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

	return &vmcommon.ContractCallInput{}, nil
}

func (sc *scProcessor) createVMDeployInput(tx *transaction.Transaction) (*vmcommon.ContractCreateInput, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, err
	}

	vmCreateInput := &vmcommon.ContractCreateInput{}
	vmCreateInput.ContractCode, err = sc.argsParser.GetCode()
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

	return vmInput, nil
}

func (sc *scProcessor) processSCPayment(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	//TODO: add gas economics here - currently sc call is free
	return nil
}

func (sc *scProcessor) moveBalances(acntSnd, acntDst state.AccountHandler, value *big.Int) error {
	operation1 := big.NewInt(0)
	operation2 := big.NewInt(0)

	// is sender address in node shard
	if acntSnd != nil {
		stAcc, ok := acntSnd.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		err := stAcc.SetBalanceWithJournal(operation1.Sub(stAcc.Balance, value))
		if err != nil {
			return err
		}
	}

	// is receiver address in node shard
	if acntDst != nil {
		stAcc, ok := acntDst.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		err := stAcc.SetBalanceWithJournal(operation2.Add(stAcc.Balance, value))
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *scProcessor) increaseNonce(acntSrc state.AccountHandler) error {
	return acntSrc.SetNonceWithJournal(acntSrc.GetNonce() + 1)
}

func (sc *scProcessor) processVMOutput(vmOutput *vmcommon.VMOutput, tx *transaction.Transaction, acntSnd, acntDst state.AccountHandler) error {
	if vmOutput == nil {
		return process.ErrNilVMOutput
	}
	if tx == nil {
		return process.ErrNilTransaction
	}

	err := sc.saveSCOutputToCurrentState(vmOutput)
	if err != nil {
		return err
	}

	err = sc.refundGasToSender(vmOutput.GasRefund, tx, acntSnd)
	if err != nil {
		return err
	}

	err = sc.processSCOutputAccounts(vmOutput.OutputAccounts, acntDst)
	if err != nil {
		return err
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return err
	}

	err = sc.processTouchedAccounts(vmOutput.TouchedAccounts)
	if err != nil {
		return err
	}

	err = sc.saveLogsIntoState(vmOutput.Logs)
	if err != nil {
		return err
	}

	return nil
}

// give back the user the unused gas money
func (sc *scProcessor) refundGasToSender(gasRefund *big.Int, tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	txPaidFee := big.NewInt(0)
	txPaidFee = txPaidFee.Mul(big.NewInt(int64(tx.GasPrice)), big.NewInt(int64(tx.GasLimit)))
	if gasRefund == nil || gasRefund.Cmp(txPaidFee) <= 0 {
		return nil
	}

	if acntSnd == nil || !acntSnd.IsInterfaceNil() {
		//TODO: sharded smart contract processing
		//TODO: create cross shard transaction here...
		return nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		log.Debug(process.ErrWrongTransaction.Error())
		return nil
	}

	refundErd := big.NewInt(0)
	refundErd = refundErd.Mul(gasRefund, big.NewInt(int64(tx.GasLimit)))

	operation := big.NewInt(0)
	err := stAcc.SetBalanceWithJournal(operation.Add(stAcc.Balance, refundErd))
	if err != nil {
		return err
	}

	return nil
}

// save account changes in state from vmOutput
func (sc *scProcessor) processSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount, acntDst state.AccountHandler) error {
	for i := 0; i < len(outputAccounts); i++ {
		outAcc := outputAccounts[i]
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return err
		}

		if acc == nil || !acc.IsInterfaceNil() {
			//TODO: sharded smart contract processing
			//TODO: create cross shard transaction here...
			continue
		}

		for j := 0; j < len(outAcc.StorageUpdates); j++ {
			storeUpdate := outAcc.StorageUpdates[j]
			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		if len(outAcc.Code) > 0 {
			acc.SetCode(outAcc.Code)
		}

		if outAcc.Nonce == nil || outAcc.Nonce.Cmp(big.NewInt(int64(acc.GetNonce()))) < 0 {
			return process.ErrWrongNonceInVMOutput
		}

		if outAcc.Balance == nil {
			continue
		}

		stAcc, ok := acc.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		difference := big.NewInt(0)
		difference = difference.Sub(stAcc.Balance, outAcc.Balance)
		err = sc.moveBalances(acntDst, stAcc, difference)
		if err != nil {
			return err
		}
	}

	return nil
}

// delete accounts
func (sc *scProcessor) deleteAccounts(deletedAccounts [][]byte) error {
	for _, value := range deletedAccounts {
		acc, err := sc.getAccountFromAddress(value)
		if err != nil {
			return err
		}

		if acc == nil || !acc.IsInterfaceNil() {
			//TODO: sharded Smart Contract processing
			continue
		}

		//TODO: protect this
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

	acnt, err := sc.accounts.GetExistingAccount(adrSrc)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// saves VM output into state
func (sc *scProcessor) saveSCOutputToCurrentState(output *vmcommon.VMOutput) error {
	var err error

	sc.mutSCState.Lock()
	defer sc.mutSCState.Unlock()

	tmpCurrScState := sc.currExecState
	defer func() {
		if err != nil {
			sc.currExecState = tmpCurrScState
		}
	}()

	err = sc.saveReturnData(output.ReturnData)
	if err != nil {
		return err
	}

	err = sc.saveReturnCode(output.ReturnCode)
	if err != nil {
		return err
	}

	err = sc.saveLogsIntoState(output.Logs)
	if err != nil {
		return err
	}

	return nil
}

// saves return data into account state
func (sc *scProcessor) saveReturnData(returnData []*big.Int) error {
	//TODO: implement
	return nil
}

// saves smart contract return code into account state
func (sc *scProcessor) saveReturnCode(returnCode vmcommon.ReturnCode) error {
	//TODO: implement
	return nil
}

// save vm output logs into accounts
func (sc *scProcessor) saveLogsIntoState(logs []*vmcommon.LogEntry) error {
	//TODO: implement
	return nil
}
