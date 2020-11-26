package process

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgsNewSystemVM {
	gasMap := make(map[string]map[string]uint64)
	gasMap[core.ElrondAPICost] = make(map[string]uint64)
	gasMap[core.ElrondAPICost][core.AsyncCallStepField] = 1000
	gasMap[core.ElrondAPICost][core.AsyncCallbackGasLockField] = 3000
	args := ArgsNewSystemVM{
		SystemEI:        &mock.SystemEIStub{},
		SystemContracts: &mock.SystemSCContainerStub{},
		VmType:          factory.SystemVirtualMachine,
		GasSchedule:     mock.NewGasScheduleNotifierMock(gasMap),
	}
	return args
}

func TestNewSystemVM_NilSystemEI(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.SystemEI = nil
	sVM, err := NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewSystemVM_NilContainer(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.SystemContracts = nil
	sVM, err := NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilSystemContractsContainer, err)
}

func TestNewSystemVM_NilVMType(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.VmType = nil
	sVM, err := NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilVMType, err)
}

func TestNewSystemVM_NilGasSchedule(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.GasSchedule = nil
	sVM, err := NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilGasSchedule, err)
}

func TestNewSystemVM_NoApiCost(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	gasMap := make(map[string]map[string]uint64)
	args.GasSchedule = mock.NewGasScheduleNotifierMock(gasMap)
	sVM, err := NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilGasSchedule, err)

	gasMap[core.ElrondAPICost] = make(map[string]uint64)
	args.GasSchedule = mock.NewGasScheduleNotifierMock(gasMap)
	sVM, err = NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilGasSchedule, err)
}

func TestNewSystemVM_Ok(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	sVM, err := NewSystemVM(args)

	assert.Nil(t, err)
	assert.NotNil(t, sVM)
}

func TestSystemVM_RunSmartContractCreate(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	sVM, _ := NewSystemVM(args)

	vmOutput, err := sVM.RunSmartContractCreate(nil)
	assert.Nil(t, vmOutput)
	assert.Equal(t, vm.ErrInputArgsIsNil, err)

	vmOutput, err = sVM.RunSmartContractCreate(&vmcommon.ContractCreateInput{})
	assert.Nil(t, vmOutput)
	assert.Equal(t, vm.ErrInputCallerAddrIsNil, err)
}

func TestSystemVM_RunSmartContractCallWrongSmartContract(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	sVM, _ := NewSystemVM(args)

	vmOutput, err := sVM.RunSmartContractCall(&vmcommon.ContractCallInput{RecipientAddr: []byte("tralala")})
	assert.Nil(t, vmOutput)
	assert.Equal(t, vm.ErrUnknownSystemSmartContract, err)
}

func TestSystemVM_RunSmartContractCall(t *testing.T) {
	t.Parallel()

	sc := &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
		return vmcommon.Ok
	}}
	scAddress := []byte("tralala")

	container := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, e error) {
		if bytes.Equal(scAddress, key) {
			return sc, nil
		}
		return nil, vm.ErrUnknownSystemSmartContract
	}}

	args := createMockArguments()
	args.SystemContracts = container
	sVM, _ := NewSystemVM(args)

	vmOutput, err := sVM.RunSmartContractCall(&vmcommon.ContractCallInput{RecipientAddr: scAddress})
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
}
