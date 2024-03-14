package badcontracts

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func Test_Bad_C_NoPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-misc/output/bad.wasm", "")
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "memoryFault")
	require.Equal(t, fmt.Errorf("execution failed"), err)
	err = context.ExecuteSC(&context.Owner, "divideByZero")
	require.Equal(t, fmt.Errorf("execution failed"), err)

	err = context.ExecuteSC(&context.Owner, "badGetOwner1")
	require.Equal(t, fmt.Errorf("bad bounds (upper)"), err)
	err = context.ExecuteSC(&context.Owner, "badBigIntStorageStore1")
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "badWriteLog1")
	require.Equal(t, fmt.Errorf("negative length"), err)
	err = context.ExecuteSC(&context.Owner, "badWriteLog2")
	require.Equal(t, fmt.Errorf("negative length"), err)
	err = context.ExecuteSC(&context.Owner, "badWriteLog3")
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Owner, "badWriteLog4")
	require.Equal(t, fmt.Errorf("mem load: bad bounds"), err)

	err = context.ExecuteSC(&context.Owner, "badGetBlockHash1")
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Owner, "badGetBlockHash2")
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Owner, "badGetBlockHash3")
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "badRecursive")
	require.Equal(t, fmt.Errorf("execution failed"), err)
}

func Test_Empty_C_NoPanic(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-empty/empty.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)
	err = context.ExecuteSC(&context.Owner, "thisDoesNotExist")
	require.Equal(t, fmt.Errorf("invalid contract code (not found)"), err)
}

func Test_Corrupt_NoPanic(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad_corrupt.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)
	err = context.ExecuteSC(&context.Owner, "thisDoesNotExist")
	require.Equal(t, fmt.Errorf("invalid contract code (not found)"), err)
}

func Test_NoMemoryDeclaration_NoPanic(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-nomemory/nomemory.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)
	err = context.ExecuteSC(&context.Owner, "memoryFault")
	require.Equal(t, fmt.Errorf("invalid contract code (not found)"), err)
}

func Test_BadFunctionNames_NoPanic(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-functionNames/output/badFunctionNames.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)
}

func Test_BadReservedFunctions(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-reservedFunctions/function-ClaimDeveloperRewards/output/bad.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)

	err = context.DeploySC("../testdata/bad-reservedFunctions/function-ChangeOwnerAddress/output/bad.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)

	err = context.DeploySC("../testdata/bad-reservedFunctions/function-asyncCall/output/bad.wasm", "")
	require.Equal(t, fmt.Errorf("invalid contract code"), err)
}
