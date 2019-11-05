package smartContract

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
)

var maxGasValue = big.NewInt(math.MaxInt64)

// scDataGetter can execute Get functions over SC to fetch stored values
type scDataGetter struct {
	vmContainer process.VirtualMachinesContainer
	mutRunSc    sync.Mutex
}

// NewSCDataGetter returns a new instance of scDataGetter
func NewSCDataGetter(
	vmContainer process.VirtualMachinesContainer,
) (*scDataGetter, error) {

	if vmContainer == nil || vmContainer.IsInterfaceNil() {
		return nil, process.ErrNoVM
	}

	return &scDataGetter{
		vmContainer: vmContainer,
	}, nil
}

func (scdg *scDataGetter) getVMFromAddress(scAddress []byte) (vmcommon.VMExecutionHandler, error) {
	vmType := hooks.VMTypeFromAddressBytes(scAddress)
	vm, err := scdg.vmContainer.Get(vmType)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// Get returns the VMOutput resulted upon invoking the function on the smart contract
func (scdg *scDataGetter) Get(scAddress []byte, funcName string, args ...[]byte) (interface{}, error) {
	if scAddress == nil {
		return nil, process.ErrNilScAddress
	}
	if len(funcName) == 0 {
		return nil, process.ErrEmptyFunctionName
	}

	scdg.mutRunSc.Lock()
	defer scdg.mutRunSc.Unlock()

	vm, err := scdg.getVMFromAddress(scAddress)
	if err != nil {
		return nil, err
	}

	vmInput := scdg.createVMCallInput(scAddress, funcName, args...)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}

	scdg.checkVMOutput(vmOutput)
	return vmOutput, nil
}

func (scdg *scDataGetter) createVMCallInput(
	scAddress []byte,
	funcName string,
	args ...[]byte,
) *vmcommon.ContractCallInput {

	argsInt := make([]*big.Int, 0)
	for _, arg := range args {
		argsInt = append(argsInt, big.NewInt(0).SetBytes(arg))
	}

	maxGasLimit := math.MaxInt64
	header := &vmcommon.SCCallHeader{
		GasLimit:    big.NewInt(int64(maxGasLimit)),
		Timestamp:   big.NewInt(0),
		Beneficiary: big.NewInt(0),
		Number:      big.NewInt(0),
	}

	vmInput := vmcommon.VMInput{
		CallerAddr:  scAddress,
		CallValue:   big.NewInt(0),
		GasPrice:    big.NewInt(0),
		GasProvided: maxGasValue,
		Arguments:   argsInt,
		Header:      header,
	}

	vmContractCallInput := &vmcommon.ContractCallInput{
		RecipientAddr: scAddress,
		Function:      funcName,
		VMInput:       vmInput,
	}

	return vmContractCallInput
}

func (scdg *scDataGetter) checkVMOutput(vmOutput *vmcommon.VMOutput) error {
	if vmOutput.ReturnCode != vmcommon.Ok {
		return errors.New(fmt.Sprintf("error running vm func: code: %d, %s", vmOutput.ReturnCode, vmOutput.ReturnCode))
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scdg *scDataGetter) IsInterfaceNil() bool {
	if scdg == nil {
		return true
	}
	return false
}
