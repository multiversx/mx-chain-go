package scrCommon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmhost "github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/stretchr/testify/require"
)

func Test_parseVMTypeFromContractAddress(t *testing.T) {
	fakeVMType, _ := hex.DecodeString("beaf")
	scAddress, _ := hex.DecodeString("0000000000000000beaf00000000000022cd8429ce92f8973bba2a9fb51e0eb3a1")
	vmType, err := vmcommon.ParseVMTypeFromContractAddress(scAddress)
	require.Nil(t, err)
	require.Equal(t, fakeVMType, vmType)
}

func Test_parseVMTypeFromContractAddress_WhenAddressTooShort(t *testing.T) {
	scAddress, _ := hex.DecodeString("0003a1")
	vmType, err := vmcommon.ParseVMTypeFromContractAddress(scAddress)
	require.Error(t, err, vmcommon.ErrInvalidVMType)
	require.Nil(t, vmType)
}

type runtimeErrorVM struct {
	mock.VMExecutionHandlerStub
	runtimeError error
}

func (vm *runtimeErrorVM) GetRuntimeErrors() error { return vm.runtimeError }

func TestLazyDataTrieLoad_VMRuntimeErrorExtraction(t *testing.T) {
	cause := core.NewGetNodeFromDBErrWithKey([]byte("root"), errors.New("missing"), "test")
	loadErr := &state.DataTrieLoadError{Cause: cause}
	unrelated := errors.New("contract error")
	for _, tc := range []struct {
		name                   string
		runtime, errorReturned error
		expected               error
	}{
		{name: "success"},
		{name: "ordinary-execution-failure", runtime: unrelated},
		{name: "legacy-descendant-failure", runtime: cause},
		{name: "direct", errorReturned: loadErr, expected: loadErr},
		{name: "validation-error-does-not-reuse-previous-runtime", runtime: loadErr, errorReturned: unrelated},
		{name: "runtime", runtime: loadErr, expected: loadErr},
		{name: "wrapped", runtime: fmt.Errorf("context: %w", loadErr), expected: loadErr},
		{name: "first-in-aggregate", runtime: vmhost.WrapError(loadErr).WrapWithError(unrelated), expected: loadErr},
		{name: "last-in-aggregate", runtime: vmhost.WrapError(unrelated).WrapWithError(loadErr), expected: loadErr},
		{name: "nested-aggregate", runtime: vmhost.WrapError(unrelated).WrapWithError(fmt.Errorf("nested: %w", vmhost.WrapError(unrelated).WrapWithError(loadErr))), expected: loadErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vm := &runtimeErrorVM{runtimeError: tc.runtime}
			actual := GetDataTrieLoadError(vm, tc.errorReturned)
			require.Equal(t, tc.expected, actual)
			if actual != nil {
				require.Same(t, cause, core.UnwrapGetNodeFromDBErr(actual))
			}
		})
	}
	require.Nil(t, GetDataTrieLoadError(&mock.VMExecutionHandlerStub{}, nil))
}

func TestLazyDataTrieLoad_HealthyRuntimeCheckDoesNotAllocate(t *testing.T) {
	executor := &runtimeErrorVM{}
	allocations := testing.AllocsPerRun(100, func() {
		if err := GetDataTrieLoadError(executor, nil); err != nil {
			panic(err)
		}
	})
	require.Zero(t, allocations)
}
