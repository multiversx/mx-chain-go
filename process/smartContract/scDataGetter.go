package smartContract

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
)

var maxGasValue = big.NewInt(math.MaxInt64)

// scDataGetter can execute Get functions over SC to fetch stored values
type scDataGetter struct {
	vm       vmcommon.VMExecutionHandler
	addrConv state.AddressConverter
	accounts state.AccountsAdapter
}

// NewSCDataGetter returns a new instance of scDataGetter
func NewSCDataGetter(
	vm vmcommon.VMExecutionHandler,
	addrConv state.AddressConverter,
	accountsDB state.AccountsAdapter,
) (*scDataGetter, error) {

	if vm == nil {
		return nil, process.ErrNoVM
	}
	if addrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if accountsDB == nil {
		return nil, process.ErrNilAccountsAdapter
	}

	return &scDataGetter{
		vm:       vm,
		addrConv: addrConv,
		accounts: accountsDB,
	}, nil
}

func (scdg *scDataGetter) Get(scAddress []byte, funcName string, args ...[]byte) ([][]byte, error) {
	if scAddress == nil {
		return nil, process.ErrNilScAddress
	}
	if len(funcName) == 0 {
		return nil, process.ErrEmptyFunctionName
	}

	vmInput, err := scdg.createVMCallInput(scAddress, funcName, args...)
	if err != nil {
		return nil, err
	}

	vmOutput, err := scdg.vm.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}

	// VM is formally verified and the output is correct
	return scdg.processVMOutput(vmOutput)
}

func (scdg *scDataGetter) createVMCallInput(scAddress []byte, funcName string, args ...[]byte) (*vmcommon.ContractCallInput, error) {
	argsInt := make([]*big.Int, 0)
	for _, arg := range args {
		argsInt = append(argsInt, big.NewInt(0).SetBytes(arg))
	}

	vmInput := vmcommon.VMInput{
		CallerAddr:  scAddress,
		CallValue:   big.NewInt(0),
		GasPrice:    big.NewInt(0),
		GasProvided: maxGasValue,
		Arguments:   argsInt,
	}
	vmContractCallInput := &vmcommon.ContractCallInput{
		RecipientAddr: scAddress,
		Function:      funcName,
		VMInput:       vmInput,
	}

	return vmContractCallInput, nil
}

func (scdg *scDataGetter) processVMOutput(vmOutput *vmcommon.VMOutput) ([][]byte, error) {
	returnedData := make([][]byte, 0)

	if vmOutput.ReturnCode != vmcommon.Ok {
		//TODO generate a stringified version of the error (after PR #7 will be merged in master)
		return nil, errors.New(fmt.Sprintf("error running vm func: code: %d", vmOutput.ReturnCode))
	}

	for _, val := range vmOutput.ReturnData {
		returnedData = append(returnedData, val.Bytes())
	}

	return returnedData, nil
}
