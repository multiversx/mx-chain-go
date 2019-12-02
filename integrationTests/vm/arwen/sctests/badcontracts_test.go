package sctests

import (
	"testing"
)

func Test_Bad_C_NoPanic(t *testing.T) {
	context := setupTestContext(t)

	context.deploySC("./badC/bad.wasm", "")

	context.executeSC(&context.Owner, "memoryFault")
	context.executeSC(&context.Owner, "divideByZero")

	context.executeSC(&context.Owner, "badGetOwner1")
	context.executeSC(&context.Owner, "badBigIntStorageStore1")

	context.executeSC(&context.Owner, "badWriteLog1")
	context.executeSC(&context.Owner, "badWriteLog2")
	context.executeSC(&context.Owner, "badWriteLog3")
	context.executeSC(&context.Owner, "badWriteLog4")

	// TODO: Uncomment upon implementing recovery (global throw / catch)
	context.executeSC(&context.Owner, "badGetBlockHash1")
	context.executeSC(&context.Owner, "badGetBlockHash2")
	context.executeSC(&context.Owner, "badGetBlockHash3")
}
