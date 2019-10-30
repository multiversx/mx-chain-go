package arwen

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-vm-common"
)

var addrConv, _ = addressConverters.NewPlainAddressConverter(32, "0x")
var maxGasValue = big.NewInt(math.MaxInt64)
var ownerAddressBytes = []byte{
	1, 2, 3, 4, 5, 6, 7, 8,
	9, 0, 1, 2, 3, 4, 5, 6,
	7, 8, 9, 0, 1, 2, 3, 4,
	5, 6, 7, 8, 9, 0, 1, 2,
}

var callerAddressBytes = []byte{
	02, 03, 05, 07, 11, 13, 17, 19,
	02, 03, 05, 07, 11, 13, 17, 19,
	02, 03, 05, 07, 11, 13, 17, 19,
	02, 03, 05, 07, 11, 13, 17, 19,
}

func Test_BasicSCMethod(t *testing.T) {
	wasmVM, scAddress := deploy(t)

	arguments := defaultArgs()
	header := defaultCallHeader()
	vmInput := defaultVMInput(header, arguments)
	vmInput.CallerAddr = callerAddressBytes

	var expectedBytes []byte

	// Call the method which sets ReturnData to nil
	callInput := makeCallInput(scAddress, "SCMethod_FinishNil", vmInput)
	vmOutput, _ := wasmVM.RunSmartContractCall(callInput)
	assertReturnData(t, vmOutput, nil)

	// Call the method which sets ReturnData to 42
	callInput = makeCallInput(scAddress, "SCMethod_Finish42", vmInput)
	vmOutput, _ = wasmVM.RunSmartContractCall(callInput)
	assertReturnData(t, vmOutput, []byte{42})

	// Call the method which sets ReturnData to the owner address
	// TODO getOwner currently returns the SC address, without padding it
	// TODO to 32 bytes.
	callInput = makeCallInput(scAddress, "SCMethod_FinishOwnerAddress", vmInput)
	vmOutput, _ = wasmVM.RunSmartContractCall(callInput)
	// assertReturnData(t, vmOutput, ownerAddressBytes)

	callInput = makeCallInput(scAddress, "SCMethod_FinishCallerAddress", vmInput)
	vmOutput, _ = wasmVM.RunSmartContractCall(callInput)
	assertReturnData(t, vmOutput, callerAddressBytes)

	vmInput.CallValue = big.NewInt(0x5DB96713)
	callInput = makeCallInput(scAddress, "SCMethod_FinishCallValue", vmInput)
	vmOutput, _ = wasmVM.RunSmartContractCall(callInput)
	assertReturnData(t, vmOutput, vmInput.CallValue.Bytes())

	callInput = makeCallInput(scAddress, "SCMethod_FinishFunctionName", vmInput)
	vmOutput, _ = wasmVM.RunSmartContractCall(callInput)
	assertReturnData(t, vmOutput, []byte("SCMethod_FinishFunctionName..-"))

	header.Timestamp.SetInt64(0x000000005DB96702)
	callInput = makeCallInput(scAddress, "SCMethod_FinishBlockTimestamp", vmInput)
	vmOutput, _ = wasmVM.RunSmartContractCall(callInput)
	expectedBytes = padBytesLeft(big.NewInt(0x5DB96702).Bytes(), 8, 0)
	assertReturnData(t, vmOutput, expectedBytes)
}

func Test_MiscSCMethods(t *testing.T) {
	wasmVM, scAddress := deploy(t)

	arguments := defaultArgs()
	header := defaultCallHeader()
	vmInput := defaultVMInput(header, arguments)
	vmInput.CallerAddr = callerAddressBytes
	vmInput.CallValue = big.NewInt(64)

	// Call the method which increments the elements of a small array
	callInput := makeCallInput(scAddress, "SCMethod_IncrementSmallArray", vmInput)
	vmOutput, _ := wasmVM.RunSmartContractCall(callInput)
	assertReturnData(t, vmOutput, []byte{1, 1, 1, 1, 1, 1, 1, 0})
}

func padBytesLeft(source []byte, totalLen int, padByte byte) []byte {
	result := make([]byte, 0)
	paddingSize := totalLen - len(source)
	for i := 0; i < paddingSize; i++ {
		result = append(result, padByte)
	}
	for _, b := range source {
		result = append(result, b)
	}
	return result
}

func assertReturnData(t *testing.T, vmOutput *vmcommon.VMOutput, expectedBytes []byte) {
	assert.Equal(t, vmcommon.Ok, vmOutput.ReturnCode, vmOutput.ReturnCode)
	if len(vmOutput.ReturnData) == 0 {
		assert.True(t, expectedBytes == nil)
		return
	}
	assert.Equal(t, 1, len(vmOutput.ReturnData))
	returnedBytes := vmOutput.ReturnData[0].Bytes()

	assert.Equal(t, expectedBytes, returnedBytes)
}

func defaultArgs() []*big.Int {
	return make([]*big.Int, 0)
}

func defaultCallHeader() *vmcommon.SCCallHeader {
	return &vmcommon.SCCallHeader{
		GasLimit:    big.NewInt(0),
		Timestamp:   big.NewInt(0),
		Beneficiary: big.NewInt(0),
		Number:      big.NewInt(0),
	}
}

func defaultVMInput(header *vmcommon.SCCallHeader, arguments []*big.Int) vmcommon.VMInput {
	return vmcommon.VMInput{
		CallerAddr:  nil,
		CallValue:   big.NewInt(0),
		GasPrice:    big.NewInt(0),
		GasProvided: maxGasValue,
		Arguments:   arguments,
		Header:      header,
	}
}

func makeCallInput(scAddress []byte, function string, vmInput vmcommon.VMInput) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		RecipientAddr: scAddress,
		Function:      function,
		VMInput:       vmInput,
	}
}

func deploy(t *testing.T) (vmcommon.VMExecutionHandler, []byte) {
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(5)

	scCode, err := ioutil.ReadFile("./eei_test.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, ownerNonce, ownerAddressBytes, ownerBalance)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.ArwenVirtualMachine),
	)
	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	vmFactory, _ := shard.NewVMContainerFactory(accnts, addrConv)
	vmContainer, _ := vmFactory.Create()
	wasmVM, _ := vmContainer.Get(factory.ArwenVirtualMachine)

	return wasmVM, scAddress
}
