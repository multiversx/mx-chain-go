package smartContract

import (
	"encoding/hex"
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

// RunAndGetVMOutput returns the VMOutput resulted upon running the function on the smart contract
func (scdg *scDataGetter) RunAndGetVMOutput(scAddress []byte, funcName string, args ...[]byte) (interface{}, error) {
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

// TODO: Move to vm-common repository, output.go

// ReturnDataKind tells us how to interpret VMOutputs's return data
type ReturnDataKind int

const (
	// AsBigInt to interpret as big int
	AsBigInt ReturnDataKind = 1 << iota
	// AsBigIntString to interpret as big int string
	AsBigIntString
	// AsString to interpret as string
	AsString
	// AsHex to interpret as hex
	AsHex
)

// GetFirstReturnData returns the first ReturnData of VMOutput, interpreted as specified.
func GetFirstReturnData(vmOutput *vmcommon.VMOutput, asType ReturnDataKind) (interface{}, error) {
	if len(vmOutput.ReturnData) == 0 {
		return nil, fmt.Errorf("no return data")
	}

	returnData := vmOutput.ReturnData[0]
	returnDataAsBytes := returnData.Bytes()
	returnDataAsString := string(returnDataAsBytes)
	returnDataAsHex := hex.EncodeToString(returnDataAsBytes)

	if asType == AsBigInt {
		return returnData, nil
	}

	if asType == AsBigIntString {
		return returnData.String(), nil
	}

	if asType == AsString {
		return returnDataAsString, nil
	}

	if asType == AsHex {
		return returnDataAsHex, nil
	}

	return nil, fmt.Errorf("can't interpret return data")
}
