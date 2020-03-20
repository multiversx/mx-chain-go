package badcontracts

import (
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"testing"
)

func Test_Bad_C_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)

	context.DeploySC("../testdata/bad/bad.wasm", "")

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

	context.DeploySC("../testdata/bad/empty.wasm", "")
	context.ExecuteSC(&context.Owner, "thisDoesNotExist")
}

func Test_Corrupt_NoPanic(t *testing.T) {
	context := arwen.SetupTestContext(t)

	context.DeploySC("../testdata/bad/corrupt.wasm", "")
	context.ExecuteSC(&context.Owner, "thisDoesNotExist")
}
