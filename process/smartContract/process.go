package smartContract

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-vm-common"
)

type scProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	vm               vmcommon.VMExecutionHandler
	argsParser       process.ArgumentsParser
}

// NewSmartContractProcessor create a smart contract processor creates and interprets VM data
func NewSmartContractProcessor(vm vmcommon.VMExecutionHandler, argsParser process.ArgumentsParser) (*scProcessor, error) {
	if vm == nil {
		return nil, process.ErrNoVM
	}
	if argsParser == nil {
		return nil, process.ErrNilArgumentParser
	}

	return &scProcessor{vm: vm, argsParser: argsParser}, nil
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
	acntSrc, acntDst state.AccountHandler,
) error {
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

	vmOutput, err := sc.vm.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	err = sc.processVMOutput(vmOutput, acntSrc, acntDst)
	if err != nil {
		return err
	}

	return nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler) error {
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

	vmOutput, err := sc.vm.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}

	err = sc.processVMOutput(vmOutput, acntSrc, acntDst)
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

	// TODO: decide who is the recipient address
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

func (sc *scProcessor) processVMOutput(vmOutput *vmcommon.VMOutput, acntSrc, acntDst state.AccountHandler) error {
	// TODO process output, pull request is comming.
	return nil
}
