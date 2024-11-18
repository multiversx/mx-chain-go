package systemSmartContracts

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
)

func TestNewSovereignVMContext(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createDefaultEeiArgs()
		vmCtx, err := NewVMContext(args)
		sovVM, err := NewOneShardSystemVMEEI(vmCtx)
		require.Nil(t, err)
		require.False(t, sovVM.IsInterfaceNil())
	})
	t.Run("nil vm context input, should return error", func(t *testing.T) {
		sovVM, err := NewOneShardSystemVMEEI(nil)
		require.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
		require.Nil(t, sovVM)
	})
}

func TestSovereignVMContext_SendGlobalSettingToAll(t *testing.T) {
	t.Parallel()

	expectedSender := []byte("sender")
	expectedInput := []byte("input")
	wasProcessBuiltInCalled := false

	args := createDefaultEeiArgs()
	args.BlockChainHook = &mock.BlockChainHookStub{
		ProcessBuiltInFunctionCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			require.Equal(t, core.SystemAccountAddress, input.RecipientAddr)
			require.Equal(t, expectedSender, input.CallerAddr)
			require.Equal(t, string(expectedInput), input.Function)
			require.Equal(t, big.NewInt(0), input.VMInput.CallValue)
			require.Zero(t, input.GasProvided)

			wasProcessBuiltInCalled = true
			return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
		},
		IsBuiltinFunctionNameCalled: func(functionName string) bool {
			return true
		},
	}
	vmCtx, err := NewVMContext(args)

	sovVM, _ := NewOneShardSystemVMEEI(vmCtx)
	err = sovVM.SendGlobalSettingToAll(expectedSender, expectedInput)
	require.Nil(t, err)
	require.True(t, wasProcessBuiltInCalled)
}
