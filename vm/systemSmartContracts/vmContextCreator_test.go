package systemSmartContracts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVMContextCreator_CreateVmContext(t *testing.T) {
	ctxCreator := NewVMContextCreator()
	require.False(t, ctxCreator.IsInterfaceNil())
	require.Implements(t, new(VMContextCreatorHandler), ctxCreator)

	args := createArgsVMContext()
	sovCtx, err := ctxCreator.CreateVmContext(args)
	require.Nil(t, err)
	require.Equal(t, "*systemSmartContracts.vmContext", fmt.Sprintf("%T", sovCtx))
}
