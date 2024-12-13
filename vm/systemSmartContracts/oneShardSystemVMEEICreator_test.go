package systemSmartContracts

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

func TestSovereignVMContextCreator_CreateVmContext(t *testing.T) {
	sovCtxCreator := NewOneShardSystemVMEEICreator()
	require.False(t, sovCtxCreator.IsInterfaceNil())
	require.Implements(t, new(VMContextCreatorHandler), sovCtxCreator)

	t.Run("should work", func(t *testing.T) {
		args := createArgsVMContext()
		sovCtx, err := sovCtxCreator.CreateVmContext(args)
		require.Nil(t, err)
		require.Equal(t, "*systemSmartContracts.oneShardSystemVMEEI", fmt.Sprintf("%T", sovCtx))
	})

	t.Run("invalid args, should return error", func(t *testing.T) {
		args := createArgsVMContext()
		args.BlockChainHook = nil
		sovCtx, err := sovCtxCreator.CreateVmContext(args)
		require.Equal(t, vm.ErrNilBlockchainHook, err)
		require.Nil(t, sovCtx)
	})
}
