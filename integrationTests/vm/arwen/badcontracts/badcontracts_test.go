package badcontracts

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
)

func Test_Bad_C_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.DeploySC("../testdata/bad-misc/bad.wasm", "")

	context.ExecuteSC(&context.Owner, "memoryFault")
	context.ExecuteSC(&context.Owner, "divideByZero")

	context.ExecuteSC(&context.Owner, "badGetOwner1")
	context.ExecuteSC(&context.Owner, "badBigIntStorageStore1")

	context.ExecuteSC(&context.Owner, "badWriteLog1")
	context.ExecuteSC(&context.Owner, "badWriteLog2")
	context.ExecuteSC(&context.Owner, "badWriteLog3")
	context.ExecuteSC(&context.Owner, "badWriteLog4")

	context.ExecuteSC(&context.Owner, "badGetBlockHash1")
	context.ExecuteSC(&context.Owner, "badGetBlockHash2")
	context.ExecuteSC(&context.Owner, "badGetBlockHash3")

	context.ExecuteSC(&context.Owner, "badRecursive")
}

func Test_Empty_C_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.DeploySC("../testdata/bad-empty/empty.wasm", "")
	context.ExecuteSC(&context.Owner, "thisDoesNotExist")
}

func Test_Corrupt_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.DeploySC("../testdata/bad_corrupt.wasm", "")
	context.ExecuteSC(&context.Owner, "thisDoesNotExist")
}

func Test_NoMemoryDeclaration_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.DeploySC("../testdata/bad-nomemory/nomemory.wasm", "")
	context.ExecuteSC(&context.Owner, "memoryFault")
}

func Test_BadFunctionNames_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.DeploySC("../testdata/bad-functionNames/badFunctionNames.wasm", "")
}
