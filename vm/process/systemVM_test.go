package process

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArguments() ArgsNewSystemVM {
	gasMap := make(map[string]map[string]uint64)
	gasMap[common.BaseOpsAPICost] = make(map[string]uint64)
	gasMap[common.BaseOpsAPICost][common.AsyncCallStepField] = 1000
	gasMap[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] = 3000
	args := ArgsNewSystemVM{
		SystemEI:        &mock.SystemEIStub{},
		SystemContracts: &mock.SystemSCContainerStub{},
		VmType:          factory.SystemVirtualMachine,
		GasSchedule:     testscommon.NewGasScheduleNotifierMock(gasMap),
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
	args.GasSchedule = testscommon.NewGasScheduleNotifierMock(gasMap)
	sVM, err := NewSystemVM(args)

	assert.Nil(t, sVM)
	assert.Equal(t, vm.ErrNilGasSchedule, err)

	gasMap[common.BaseOpsAPICost] = make(map[string]uint64)
	args.GasSchedule = testscommon.NewGasScheduleNotifierMock(gasMap)
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
	args.SystemEI = &mock.SystemEIStub{GetContractCalled: func(address []byte) (vm.SystemSmartContract, error) {
		return nil, vm.ErrUnknownSystemSmartContract
	}}
	sVM, _ := NewSystemVM(args)

	vmOutput, err := sVM.RunSmartContractCall(&vmcommon.ContractCallInput{RecipientAddr: []byte("dummyAddress")})
	assert.Nil(t, vmOutput)
	assert.Equal(t, vm.ErrUnknownSystemSmartContract, err)
}

func TestSystemVM_RunSmartContractCall(t *testing.T) {
	t.Parallel()

	sc := &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
		return vmcommon.Ok
	}}
	scAddress := []byte("dummyAddress")

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

func TestLazyDataTrieLoad_SystemVM(t *testing.T) {
	cause := core.NewGetNodeFromDBErrWithKey([]byte("root"), errors.New("missing"), "test")
	loadErr := &state.DataTrieLoadError{Cause: cause}
	for _, deploy := range []bool{false, true} {
		name := "call"
		if deploy {
			name = "create"
		}
		t.Run(name, func(t *testing.T) {
			var storageErr error
			faulty := true
			contract := &mock.SystemSCStub{ExecuteCalled: func(*vmcommon.ContractCallInput) vmcommon.ReturnCode {
				if faulty {
					storageErr = loadErr
				}
				return vmcommon.Ok
			}}
			args := createMockArguments()
			args.SystemEI = &mock.SystemEIStub{
				CleanCacheCalled:      func() { storageErr = nil },
				GetContractCalled:     func([]byte) (vm.SystemSmartContract, error) { return contract, nil },
				GetStorageErrorCalled: func() error { return storageErr },
			}
			args.SystemContracts = &mock.SystemSCContainerStub{GetCalled: func([]byte) (vm.SystemSmartContract, error) { return contract, nil }}
			system, err := NewSystemVM(args)
			require.NoError(t, err)
			run := func() (*vmcommon.VMOutput, error) {
				if deploy {
					return system.RunSmartContractCreate(&vmcommon.ContractCreateInput{VMInput: vmcommon.VMInput{CallerAddr: []byte("sc"), CallValue: big.NewInt(0)}})
				}
				return system.RunSmartContractCall(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}, RecipientAddr: []byte("sc"), Function: "read"})
			}
			output, err := run()
			require.Same(t, loadErr, err)
			require.Nil(t, output)
			require.Same(t, cause, core.UnwrapGetNodeFromDBErr(err))
			faulty = false
			output, err = run()
			require.NoError(t, err)
			require.Equal(t, vmcommon.Ok, output.ReturnCode)
		})
	}
}
