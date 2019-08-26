package smartContract

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
)

var maxGasValue = big.NewInt(math.MaxInt64)

// scDataGetter can execute Get functions over SC to fetch stored values
type scDataGetter struct {
	vm       vmcommon.VMExecutionHandler
	mutRunSc sync.Mutex
}

// NewSCDataGetter returns a new instance of scDataGetter
func NewSCDataGetter(
	vm vmcommon.VMExecutionHandler,
) (*scDataGetter, error) {

	if vm == nil {
		return nil, process.ErrNoVM
	}

	return &scDataGetter{
		vm: vm,
	}, nil
}

// Get returns the value as byte slice of the invoked func
func (scdg *scDataGetter) Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error) {
	if scAddress == nil {
		return nil, process.ErrNilScAddress
	}
	if len(funcName) == 0 {
		return nil, process.ErrEmptyFunctionName
	}

	scdg.mutRunSc.Lock()
	defer scdg.mutRunSc.Unlock()

	vmInput := scdg.createVMCallInput(scAddress, funcName, args...)
	vmOutput, err := scdg.vm.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}

	// VM is formally verified and the output is correct
	return scdg.checkVMOutput(vmOutput)
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

	header := &vmcommon.SCCallHeader{
		GasLimit:    big.NewInt(0),
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

func (scdg *scDataGetter) checkVMOutput(vmOutput *vmcommon.VMOutput) ([]byte, error) {
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, errors.New(fmt.Sprintf("error running vm func: code: %d, %s", vmOutput.ReturnCode, vmOutput.ReturnCode))
	}

	if len(vmOutput.ReturnData) > 0 {
		return vmOutput.ReturnData[0].Bytes(), nil
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scdg *scDataGetter) IsInterfaceNil() bool {
	if scdg == nil {
		return true
	}
	return false
}
