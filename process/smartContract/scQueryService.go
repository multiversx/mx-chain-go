package smartContract

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/pkg/errors"
)

var _ process.SCQueryService = (*SCQueryService)(nil)

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer  process.VirtualMachinesContainer
	economicsFee process.FeeHandler
	mutRunSc     sync.Mutex
}

// NewSCQueryService returns a new instance of SCQueryService
func NewSCQueryService(
	vmContainer process.VirtualMachinesContainer,
	economicsFee process.FeeHandler,
) (*SCQueryService, error) {
	if check.IfNil(vmContainer) {
		return nil, process.ErrNoVM
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	return &SCQueryService{
		vmContainer:  vmContainer,
		economicsFee: economicsFee,
	}, nil
}

// ExecuteQuery returns the VMOutput resulted upon running the function on the smart contract
func (service *SCQueryService) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	err := checkQueryFields(query)
	if err != nil {
		return nil, err
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	return service.executeScCall(query, 0, nil, core.ScQueryDefaultCallValue)
}

// ExecuteQueryWithCallerAndValue returns the VMOutput resulted upon running the function on the smart contract
// and includes the caller address and the value
func (service *SCQueryService) ExecuteQueryWithCallerAndValue(query *process.SCQuery, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error) {
	err := checkQueryFields(query)
	if err != nil {
		return nil, err
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	return service.executeScCall(query, 0, callerAddr, callValue)
}

func checkQueryFields(query *process.SCQuery) error {
	if query.ScAddress == nil {
		return process.ErrNilScAddress
	}
	if len(query.FuncName) == 0 {
		return process.ErrEmptyFunctionName
	}

	return nil
}

func (service *SCQueryService) executeScCall(query *process.SCQuery, gasPrice uint64, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error) {
	vm, err := findVMByScAddress(service.vmContainer, query.ScAddress)
	if err != nil {
		return nil, err
	}

	vmInput := service.createVMCallInput(query, gasPrice)
	if callerAddr != nil &&
		len(callerAddr) > 0 &&
		bytes.Compare(callerAddr, []byte(core.ScQueryDefaultCallerAddress)) != 0 {
		vmInput.CallerAddr = callerAddr
	}
	if callValue != nil && callValue.Cmp(core.ScQueryDefaultCallValue) != 0 {
		vmInput.CallValue = callValue
	}
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
		GasProvided: service.economicsFee.MaxGasLimitPerBlock(0),
		Arguments:   query.Arguments,
		CallType:    vmcommon.DirectCall,
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
		return errors.New(fmt.Sprintf("error running vm func: code: %d, %s (%s)",
			vmOutput.ReturnCode,
			vmOutput.ReturnCode,
			vmOutput.ReturnMessage))
	}

	return nil
}

// ComputeScCallGasLimit will estimate how many gas a transaction will consume
func (service *SCQueryService) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	argumentParser := parsers.NewCallArgsParser()

	function, arguments, err := argumentParser.ParseData(string(tx.Data))
	if err != nil {
		return 0, err
	}

	query := &process.SCQuery{
		ScAddress: tx.RcvAddr,
		FuncName:  function,
		Arguments: arguments,
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	vmOutput, err := service.executeScCall(query, 1, nil, core.ScQueryDefaultCallValue)
	if err != nil {
		return 0, err
	}

	gasConsumed := service.economicsFee.MaxGasLimitPerBlock(0) - vmOutput.GasRemaining

	return gasConsumed, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (service *SCQueryService) IsInterfaceNil() bool {
	return service == nil
}
