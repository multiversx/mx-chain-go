package smartContract

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
)

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer      process.VirtualMachinesContainer
	gasLimitPerBlock uint64
	mutRunSc         sync.Mutex
}

// NewSCQueryService returns a new instance of SCQueryService
func NewSCQueryService(
	vmContainer process.VirtualMachinesContainer,
	gasLimitPerBlock uint64,
) (*SCQueryService, error) {

	if vmContainer == nil || vmContainer.IsInterfaceNil() {
		return nil, process.ErrNoVM
	}

	return &SCQueryService{
		vmContainer:      vmContainer,
		gasLimitPerBlock: gasLimitPerBlock,
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

	vm, err := service.getVMFromAddress(query.ScAddress)
	if err != nil {
		return nil, err
	}

	vmInput := service.createVMCallInput(query)
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

func (service *SCQueryService) createVMCallInput(query *process.SCQuery) *vmcommon.ContractCallInput {
	vmInput := vmcommon.VMInput{
		CallerAddr:  query.ScAddress,
		CallValue:   big.NewInt(0),
		GasPrice:    0,
		GasProvided: service.gasLimitPerBlock,
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

// IsInterfaceNil returns true if there is no value under the interface
func (service *SCQueryService) IsInterfaceNil() bool {
	if service == nil {
		return true
	}
	return false
}
