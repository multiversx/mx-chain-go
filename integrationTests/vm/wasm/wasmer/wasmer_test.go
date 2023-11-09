//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package wasmer

import (
	"fmt"
	"math/big"
	"testing"

	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process/factory"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

var ownerAddressBytes = []byte("12345678901234567890123456789012")

func TestAllowNonFloatingPointSC(t *testing.T) {
	wasmvm, scAddress := deploy(t, "../testdata/floating_point/non_fp.wasm")
	defer closeVM(wasmvm)

	arguments := make([][]byte, 0)
	vmInput := defaultVMInput(arguments)
	vmInput.CallerAddr = ownerAddressBytes

	callInput := makeCallInput(scAddress, "doSomething", vmInput)
	vmOutput, err := wasmvm.RunSmartContractCall(callInput)
	require.Nil(t, err)

	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	fmt.Printf("VM Return Code: %s\n", vmOutput.ReturnCode)
}

func TestDisallowFloatingPointSC(t *testing.T) {
	wasmvm, scAddress := deploy(t, "../testdata/floating_point/fp.wasm")
	defer closeVM(wasmvm)

	arguments := make([][]byte, 0)
	vmInput := defaultVMInput(arguments)
	vmInput.CallerAddr = ownerAddressBytes

	callInput := makeCallInput(scAddress, "doSomething", vmInput)
	vmOutput, err := wasmvm.RunSmartContractCall(callInput)
	require.Nil(t, err)

	require.Equal(t, vmcommon.ContractNotFound, vmOutput.ReturnCode)
	fmt.Printf("VM Return Code: %s\n", vmOutput.ReturnCode)
}

func TestSCAbortExecution_DontAbort(t *testing.T) {
	wasmvm, scAddress := deploy(t, "../testdata/misc/test_abort/test_abort.wasm")
	defer closeVM(wasmvm)

	// Run testFunc with argument 0, which will not abort execution, leading to a
	// call to int64finish(100).
	arguments := make([][]byte, 0)
	arguments = append(arguments, []byte{0x00})

	vmInput := defaultVMInput(arguments)
	vmInput.CallerAddr = ownerAddressBytes

	callInput := makeCallInput(scAddress, "testFunc", vmInput)
	vmOutput, err := wasmvm.RunSmartContractCall(callInput)
	require.Nil(t, err)

	expectedBytes := []byte{100}
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	assertReturnData(t, vmOutput, vmcommon.Ok, expectedBytes)
}

func TestSCAbortExecution_Abort(t *testing.T) {
	wasmvm, scAddress := deploy(t, "../testdata/misc/test_abort/test_abort.wasm")
	defer closeVM(wasmvm)

	arguments := make([][]byte, 0)
	arguments = append(arguments, []byte{0x01})

	vmInput := defaultVMInput(arguments)
	vmInput.CallerAddr = ownerAddressBytes

	callInput := makeCallInput(scAddress, "testFunc", vmInput)
	vmOutput, err := wasmvm.RunSmartContractCall(callInput)
	require.Nil(t, err)

	require.Equal(t, 0, len(vmOutput.ReturnData))
	assertReturnData(t, vmOutput, vmcommon.UserError, nil)
	require.Equal(t, "abort here", vmOutput.ReturnMessage)
}

func deploy(t *testing.T, wasmFilename string) (vmcommon.VMExecutionHandler, []byte) {
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xfffffffffffffff)

	scCode := wasm.GetSCCode(wasmFilename)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		config.EnableEpochs{},
	)
	require.Nil(t, err)

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode),
	)
	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)

	wasmVM, _ := testContext.VMContainer.Get(factory.WasmVirtualMachine)
	return wasmVM, scAddressBytes
}

func assertReturnData(
	t *testing.T,
	vmOutput *vmcommon.VMOutput,
	expectedReturnCode vmcommon.ReturnCode,
	expectedBytes []byte,
) {
	require.Equal(t, expectedReturnCode, vmOutput.ReturnCode, vmOutput.ReturnCode)
	if len(vmOutput.ReturnData) == 0 {
		require.True(t, expectedBytes == nil)
		return
	}
	require.Equal(t, 1, len(vmOutput.ReturnData))
	returnedBytes := vmOutput.ReturnData[0]

	require.Equal(t, expectedBytes, returnedBytes)
}

func makeCallInput(scAddress []byte, function string, vmInput vmcommon.VMInput) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		RecipientAddr: scAddress,
		Function:      function,
		VMInput:       vmInput,
	}
}

func defaultVMInput(arguments [][]byte) vmcommon.VMInput {
	return vmcommon.VMInput{
		CallerAddr:  nil,
		CallValue:   big.NewInt(0),
		GasPrice:    uint64(0),
		GasProvided: uint64(0xfffffffffffffff),
		Arguments:   arguments,
		CallType:    vmData.DirectCall,
	}
}

func closeVM(wasmvm vmcommon.VMExecutionHandler) {
	if asCloser, ok := wasmvm.(interface{ Close() error }); ok {
		_ = asCloser.Close()
	}
}
