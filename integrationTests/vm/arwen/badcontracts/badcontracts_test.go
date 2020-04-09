package badcontracts

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

func Test_Bad_C_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-misc/bad.wasm", "")
	require.Nil(t, err)

	_ = context.ExecuteSC(&context.Owner, "memoryFault")
	_ = context.ExecuteSC(&context.Owner, "divideByZero")

	_ = context.ExecuteSC(&context.Owner, "badGetOwner1")
	_ = context.ExecuteSC(&context.Owner, "badBigIntStorageStore1")

	_ = context.ExecuteSC(&context.Owner, "badWriteLog1")
	_ = context.ExecuteSC(&context.Owner, "badWriteLog2")
	_ = context.ExecuteSC(&context.Owner, "badWriteLog3")
	_ = context.ExecuteSC(&context.Owner, "badWriteLog4")

	_ = context.ExecuteSC(&context.Owner, "badGetBlockHash1")
	_ = context.ExecuteSC(&context.Owner, "badGetBlockHash2")
	_ = context.ExecuteSC(&context.Owner, "badGetBlockHash3")

	_ = context.ExecuteSC(&context.Owner, "badRecursive")
}

func Test_Empty_C_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-empty/empty.wasm", "")
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Owner, "thisDoesNotExist")
	require.NotNil(t, err)
}

func Test_Corrupt_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad_corrupt.wasm", "")
	require.NotNil(t, err)
	err = context.ExecuteSC(&context.Owner, "thisDoesNotExist")
	require.NotNil(t, err)
}

func Test_NoMemoryDeclaration_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-nomemory/nomemory.wasm", "")
	require.NotNil(t, err)
	err = context.ExecuteSC(&context.Owner, "memoryFault")
	require.NotNil(t, err)
}

func Test_BadFunctionNames_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/bad-functionNames/badFunctionNames.wasm", "")
	require.NotNil(t, err)
}
