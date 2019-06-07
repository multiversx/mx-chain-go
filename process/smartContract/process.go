package smartContract

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
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
	fakeAccounts     process.FakeAccountsHandler
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	vm               vmcommon.VMExecutionHandler
	argsParser       process.ArgumentsParser

	mutSCState   sync.Mutex
	mapExecState map[uint32]scExecutionState
}

// NewSmartContractProcessor create a smart contract processor creates and interprets VM data
func NewSmartContractProcessor(
	vm vmcommon.VMExecutionHandler,
	argsParser process.ArgumentsParser,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountsDB state.AccountsAdapter,
	fakeAccounts process.FakeAccountsHandler,
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
	if fakeAccounts == nil {
		return nil, process.ErrNilFakeAccountsHandler
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
		fakeAccounts:     fakeAccounts,
		adrConv:          adrConv,
		shardCoordinator: coordinator,
		mapExecState:     make(map[uint32]scExecutionState)}, nil
}

// ComputeTransactionType calculates the type of the transaction
func (sc *scProcessor) ComputeTransactionType(
	tx *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
) (process.TransactionType, error) {
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
	round uint32,
) error {
	defer sc.fakeAccounts.CleanFakeAccounts()

	if tx == nil {
		return process.ErrNilTransaction
	}
	if acntDst == nil {
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

	vmOutput, err := sc.vm.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	// VM is formally verified and the output is correct
	err = sc.processVMOutput(vmOutput, tx, acntSnd, acntDst, round)
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) prepareSmartContractCall(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	err := sc.argsParser.ParseData(tx.Data)
	if err != nil {
		return err
	}

	totalBalance, err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	sc.fakeAccounts.CreateFakeAccounts(tx.SndAddr, totalBalance, tx.Nonce)

	return nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(tx *transaction.Transaction, acntSnd, acntDst state.AccountHandler, round uint32) error {
	defer sc.fakeAccounts.CleanFakeAccounts()

	if len(tx.RcvAddr) != 0 {
		return process.ErrWrongTransaction
	}

	err := sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		return err
	}

	vmInput, err := sc.createVMDeployInput(tx)
	if err != nil {
		return err
	}

	// TODO: Smart contract address calculation
	vmOutput, err := sc.vm.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}

	// VM is formally verified, the output is correct
	err = sc.processVMOutput(vmOutput, tx, acntSnd, acntDst, round)
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

	return vmCallInput, nil
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

// taking money from sender, as VM might not have access to him because of state sharding
func (sc *scProcessor) processSCPayment(tx *transaction.Transaction, acntSnd state.AccountHandler) (*big.Int, error) {
	operation := big.NewInt(0)
	operation = operation.Mul(big.NewInt(int64(tx.GasPrice)), big.NewInt(int64(tx.GasLimit)))
	operation = operation.Add(operation, tx.Value)

	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		// transaction was already done at sender shard
		return operation, nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	if stAcc.Balance.Cmp(operation) < 0 {
		return nil, process.ErrInsufficientFunds
	}

	totalCost := big.NewInt(0)
	err := stAcc.SetBalanceWithJournal(totalCost.Sub(stAcc.Balance, operation))
	if err != nil {
		return nil, err
	}

	err = stAcc.SetNonceWithJournal(stAcc.GetNonce() + 1)
	if err != nil {
		return nil, err
	}

	return operation, nil
}

func (sc *scProcessor) processVMOutput(vmOutput *vmcommon.VMOutput, tx *transaction.Transaction, acntSnd, acntDst state.AccountHandler, round uint32) error {
	if vmOutput == nil {
		return process.ErrNilVMOutput
	}
	if tx == nil {
		return process.ErrNilTransaction
	}

	txBytes, err := sc.marshalizer.Marshal(tx)
	if err != nil {
		return err
	}
	txHash := sc.hasher.Compute(string(txBytes))

	err = sc.saveSCOutputToCurrentState(vmOutput, round, txHash)
	if err != nil {
		return err
	}

	err = sc.refundGasToSender(vmOutput.GasRefund, tx, acntSnd)
	if err != nil {
		return err
	}

	err = sc.processSCOutputAccounts(vmOutput.OutputAccounts)
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

	return nil
}

// give back the user the unused gas money
func (sc *scProcessor) refundGasToSender(gasRefund *big.Int, tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	if gasRefund == nil || gasRefund.Cmp(big.NewInt(0)) <= 0 {
		return nil
	}

	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		//TODO: sharded smart contract processing
		//TODO: create cross shard transaction here...
		return nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	refundErd := big.NewInt(0)
	refundErd = refundErd.Mul(gasRefund, big.NewInt(int64(tx.GasPrice)))

	operation := big.NewInt(0)
	err := stAcc.SetBalanceWithJournal(operation.Add(stAcc.Balance, refundErd))
	if err != nil {
		return err
	}

	return nil
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount) error {
	for i := 0; i < len(outputAccounts); i++ {
		outAcc := outputAccounts[i]
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return err
		}

		fakeAcc := sc.fakeAccounts.GetFakeAccount(outAcc.Address)

		if acc == nil || acc.IsInterfaceNil() {
			//TODO: sharded smart contract processing
			//TODO: create cross shard transaction here...
			//if fakeAcc use the difference between outacc and fakeAcc for the new transaction
			continue
		}

		for j := 0; j < len(outAcc.StorageUpdates); j++ {
			storeUpdate := outAcc.StorageUpdates[j]
			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		if len(outAcc.Code) > 0 {
			err = acc.SetCodeWithJournal(outAcc.Code)
			if err != nil {
				return err
			}

			hash := sc.hasher.Compute(string(outAcc.Code))
			err = acc.SetCodeHashWithJournal(hash)
			if err != nil {
				return err
			}
		}

		if outAcc.Nonce == nil || outAcc.Nonce.Cmp(big.NewInt(int64(acc.GetNonce()))) < 0 {
			return process.ErrWrongNonceInVMOutput
		}

		err = acc.SetNonceWithJournal(outAcc.Nonce.Uint64())
		if err != nil {
			return err
		}

		if outAcc.Balance == nil {
			return process.ErrNilBalanceFromSC
		}

		stAcc, ok := acc.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		// if fake account, than VM so only transaction value plus fee as balance, so anything remaining is a plus
		if fakeAcc != nil && !fakeAcc.IsInterfaceNil() {
			outAcc.Balance = outAcc.Balance.Add(outAcc.Balance, stAcc.Balance)
		}

		// update the values according to SC output
		err = stAcc.SetBalanceWithJournal(outAcc.Balance)
		if err != nil {
			return err
		}
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

// GetAllSmartContractCallRoothash returns the roothash of the state of the SC executions for defined round
func (sc *scProcessor) GetAllSmartContractCallRootHash(round uint32) []byte {
	return []byte("roothash")
}

// saves VM output into state
func (sc *scProcessor) saveSCOutputToCurrentState(output *vmcommon.VMOutput, round uint32, txHash []byte) error {
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
func (sc *scProcessor) saveReturnData(returnData []*big.Int, round uint32, txHash []byte) error {
	sc.mapExecState[round].allReturnData[string(txHash)] = returnData
	return nil
}

// saves smart contract return code into account state
func (sc *scProcessor) saveReturnCode(returnCode vmcommon.ReturnCode, round uint32, txHash []byte) error {
	sc.mapExecState[round].returnCodes[string(txHash)] = returnCode
	return nil
}

// save vm output logs into accounts
func (sc *scProcessor) saveLogsIntoState(logs []*vmcommon.LogEntry, round uint32, txHash []byte) error {
	sc.mapExecState[round].allLogs[string(txHash)] = logs
	return nil
}
