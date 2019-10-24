package smartContract

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
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
	vmContainer   process.VirtualMachinesContainer
	addrConverter state.AddressConverter
	mutRunSc      sync.Mutex
}

// NewSCDataGetter returns a new instance of scDataGetter
func NewSCDataGetter(
	addrConverter state.AddressConverter,
	vmContainer process.VirtualMachinesContainer,
) (*scDataGetter, error) {

	if vmContainer == nil || vmContainer.IsInterfaceNil() {
		return nil, process.ErrNoVM
	}
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}

	return &scDataGetter{
		vmContainer:   vmContainer,
		addrConverter: addrConverter,
	}, nil
}

// Get returns the value as byte slice of the invoked func
func (scdg *scDataGetter) Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error) {
	if scAddress == nil {
		return nil, process.ErrNilScAddress
	}
	if len(scAddress) < scdg.addrConverter.AddressLen() {
		return nil, process.ErrInvalidRcvAddr
	}
	if len(funcName) == 0 {
		return nil, process.ErrEmptyFunctionName
	}

	scdg.mutRunSc.Lock()
	defer scdg.mutRunSc.Unlock()

	vmType := scAddress[hooks.NumInitCharactersForScAddress-hooks.VMTypeLen : hooks.NumInitCharactersForScAddress]
	vm, err := scdg.vmContainer.Get(vmType)
	if err != nil {
		return nil, err
	}

	vmInput := scdg.createVMCallInput(scAddress, funcName, args...)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
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
