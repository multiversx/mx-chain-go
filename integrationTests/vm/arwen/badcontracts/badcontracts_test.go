package badcontracts

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

func Test_Bad_C_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-misc/bad.wasm", "")
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "memoryFault")
	require.Equal(t, fmt.Errorf("execution failed"), err)
	err = context.ExecuteSC(&context.Owner, "divideByZero")
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "badGetOwner1")
	require.Equal(t, fmt.Errorf("execution failed"), err)
	err = context.ExecuteSC(&context.Owner, "badBigIntStorageStore1")
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "badWriteLog1")
	require.Equal(t, fmt.Errorf("execution failed"), err)
	err = context.ExecuteSC(&context.Owner, "badWriteLog2")
	require.Equal(t, fmt.Errorf("execution failed"), err)
	err = context.ExecuteSC(&context.Owner, "badWriteLog3")
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Owner, "badWriteLog4")
	require.Equal(t, fmt.Errorf("execution failed"), err)

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
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-empty/empty.wasm", "")
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Owner, "thisDoesNotExist")
	require.Equal(t, fmt.Errorf("function not found"), err)
}

func Test_Corrupt_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad_corrupt.wasm", "")
	require.Equal(t, fmt.Errorf("contract invalid"), err)
	err = context.ExecuteSC(&context.Owner, "thisDoesNotExist")
	require.Equal(t, fmt.Errorf("contract not found"), err)
}

func Test_NoMemoryDeclaration_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-nomemory/nomemory.wasm", "")
	require.Equal(t, fmt.Errorf("contract invalid"), err)
	err = context.ExecuteSC(&context.Owner, "memoryFault")
	require.Equal(t, fmt.Errorf("contract not found"), err)
}

func Test_BadFunctionNames_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-functionNames/badFunctionNames.wasm", "")
	require.Equal(t, fmt.Errorf("contract invalid"), err)
}

func Test_BadReservedFunctions(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-reservedFunctions/function-ClaimDeveloperRewards.wasm", "")
	require.Equal(t, fmt.Errorf("contract invalid"), err)

	err = context.DeploySC("../testdata/bad-reservedFunctions/function-ChangeOwnerAddress.wasm", "")
	require.Equal(t, fmt.Errorf("contract invalid"), err)

	err = context.DeploySC("../testdata/bad-reservedFunctions/function-asyncCall.wasm", "")
	require.Equal(t, fmt.Errorf("contract invalid"), err)

	err = context.DeploySC("../testdata/bad-reservedFunctions/function-foobar.wasm", "")
	require.Nil(t, err)
}
