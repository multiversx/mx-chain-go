package smartContract

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
)

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer   process.VirtualMachinesContainer
	txTypeHandler process.TxTypeHandler
	economicsFee  process.FeeHandler
	mutRunSc      sync.Mutex
}

// NewSCQueryService returns a new instance of SCQueryService
func NewSCQueryService(
	vmContainer process.VirtualMachinesContainer,
	txTypeHandler process.TxTypeHandler,
	economicsFee process.FeeHandler,
) (*SCQueryService, error) {
	if check.IfNil(vmContainer) {
		return nil, process.ErrNoVM
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	return &SCQueryService{
		vmContainer:   vmContainer,
		txTypeHandler: txTypeHandler,
		economicsFee:  economicsFee,
	}, nil
}

func (service *SCQueryService) getVMFromAddress(scAddress []byte) (vmcommon.VMExecutionHandler, error) {
	vmType := core.GetVMType(scAddress)
	vm, err := service.vmContainer.Get(vmType)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// ExecuteQuery returns the VMOutput resulted upon running the function on the smart contract
func (service *SCQueryService) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if query.ScAddress == nil {
		return nil, process.ErrNilScAddress
	}
	if len(query.FuncName) == 0 {
		return nil, process.ErrEmptyFunctionName
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	return service.executeScCall(query, 0)
}

func (service *SCQueryService) executeScCall(query *process.SCQuery, gasPrice uint64) (*vmcommon.VMOutput, error) {
	vm, err := service.getVMFromAddress(query.ScAddress)
	if err != nil {
		return nil, err
	}

	vmInput := service.createVMCallInput(query, gasPrice)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}

	err = service.checkVMOutput(vmOutput)
	if err != nil {
		return nil, err
	}

	return vmOutput, nil
}

func (service *SCQueryService) createVMCallInput(query *process.SCQuery, gasPrice uint64) *vmcommon.ContractCallInput {
	vmInput := vmcommon.VMInput{
		CallerAddr:  query.ScAddress,
		CallValue:   big.NewInt(0),
		GasPrice:    gasPrice,
		GasProvided: service.economicsFee.MaxGasLimitPerBlock(),
		Arguments:   query.Arguments,
	}

	vmContractCallInput := &vmcommon.ContractCallInput{
		RecipientAddr: query.ScAddress,
		Function:      query.FuncName,
		VMInput:       vmInput,
	}

	return vmContractCallInput
}

func (service *SCQueryService) checkVMOutput(vmOutput *vmcommon.VMOutput) error {
	if vmOutput.ReturnCode != vmcommon.Ok {
		return errors.New(fmt.Sprintf("error running vm func: code: %d, %s", vmOutput.ReturnCode, vmOutput.ReturnCode))
	}

	return nil
}

// ComputeTransactionCost will estimate how many gas a transaction will consume
func (service *SCQueryService) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	txType, err := service.txTypeHandler.ComputeTransactionType(tx)
	if err != nil {
		return nil, err
	}

	tx.GasPrice = 1

	switch txType {
	case process.MoveBalance:
		return service.estimateMoveBalance(tx)
	case process.SCDeployment:
		return service.estimateSCDeployment(tx)
	case process.SCInvoking:
		return service.estimateSCInvoking(tx)
	default:
		return nil, process.ErrWrongTransaction
	}
}

func (service *SCQueryService) estimateMoveBalance(tx *transaction.Transaction) (*big.Int, error) {
	cost := service.economicsFee.ComputeFee(tx)
	return cost, nil
}

func (service *SCQueryService) estimateSCDeployment(tx *transaction.Transaction) (*big.Int, error) {
	cost := service.economicsFee.ComputeFee(tx)
	return cost, nil
}

func (service *SCQueryService) estimateSCInvoking(tx *transaction.Transaction) (*big.Int, error) {
	argumentParser, _ := vmcommon.NewAtArgumentParser()

	err := argumentParser.ParseData(string(tx.Data))
	if err != nil {
		return nil, err
	}

	function, err := argumentParser.GetFunction()
	if err != nil {
		return nil, err
	}

	arguments, err := argumentParser.GetArguments()
	if err != nil {
		return nil, err
	}

	query := &process.SCQuery{
		ScAddress: tx.RcvAddr,
		FuncName:  function,
		Arguments: arguments,
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	vmOutput, err := service.executeScCall(query, 1)
	if err != nil {
		return nil, err
	}

	gasConsumed := service.economicsFee.MaxGasLimitPerBlock() - vmOutput.GasRemaining

	return big.NewInt(0).SetUint64(gasConsumed), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (service *SCQueryService) IsInterfaceNil() bool {
	return service == nil
}
