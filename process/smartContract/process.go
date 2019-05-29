package smartContract

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-vm-common"
)

type scProcessor struct {
	vm vmcommon.VMExecutionHandler
}

// NewSmartContractProcessor create a smart contract processor creates and interprets VM data
func NewSmartContractProcessor(vm vmcommon.VMExecutionHandler) (*scProcessor, error) {
	if vm == nil {
		return nil, process.ErrNoVM
	}

	return &scProcessor{vm: vm}, nil
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
	if sc.vm == nil {
		return process.ErrNoVM
	}

	vmInput, err := sc.createVMCallInput(tx, acntSrc, acntDst)
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

// DeploySmartContract runs the VM to verify register the smart contract and saves the data into the state
func (sc *scProcessor) DeploySmartContract(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler) error {
	if sc.vm == nil {
		return process.ErrNoVM
	}

	vmInput, err := sc.createVMDeployInput(tx, acntSrc, acntDst)
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

func (sc *scProcessor) createVMCallInput(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler) (*vmcommon.ContractCallInput, error) {
	return &vmcommon.ContractCallInput{}, nil
}

func (sc *scProcessor) createVMDeployInput(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler) (*vmcommon.ContractCreateInput, error) {
	return &vmcommon.ContractCreateInput{}, nil
}

func (sc *scProcessor) processVMOutput(vmOutput *vmcommon.VMOutput, acntSrc, acntDst state.AccountHandler) error {
	return nil
}
