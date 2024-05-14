package systemSmartContracts

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignVMContext(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		sovVM, err := NewSystemVMEEI(&mock.SystemEIStub{})
		require.Nil(t, err)
		require.False(t, sovVM.IsInterfaceNil())
	})
	t.Run("nil vm context input, should return error", func(t *testing.T) {
		sovVM, err := NewSystemVMEEI(nil)
		require.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
		require.Nil(t, sovVM)
	})
}

func TestSovereignVMContext_SendGlobalSettingToAll(t *testing.T) {
	t.Parallel()

	expectedSender := []byte("sender")
	expectedInput := []byte("input")
	wasProcessBuiltInCalled := false
	vmCtx := &mock.SystemEIStub{
		ProcessBuiltInFunctionCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, gasLimit uint64) error {
			require.Equal(t, core.SystemAccountAddress, destination)
			require.Equal(t, expectedSender, sender)
			require.Equal(t, expectedInput, input)
			require.Equal(t, big.NewInt(0), value)
			require.Zero(t, gasLimit)

			wasProcessBuiltInCalled = true
			return nil
		},
	}

	sovVM, _ := NewSystemVMEEI(vmCtx)
	err := sovVM.SendGlobalSettingToAll(expectedSender, expectedInput)
	require.Nil(t, err)
	require.True(t, wasProcessBuiltInCalled)
}
