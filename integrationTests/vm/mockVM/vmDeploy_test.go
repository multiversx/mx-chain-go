package mockVM

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithoutTransferShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(0)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(0)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	_, err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes, _ := hex.DecodeString("b473e38c40c60c8b22eb80fffff0717a98386bb129a7750ffe8091c77b85990e")

	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVmDeployWithTransferShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(0)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	_, err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes, _ := hex.DecodeString("b473e38c40c60c8b22eb80fffff0717a98386bb129a7750ffe8091c77b85990e")
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	//equal with requirement
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	_, err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))

	destinationAddressBytes, _ := hex.DecodeString("b473e38c40c60c8b22eb80fffff0717a98386bb129a7750ffe8091c77b85990e")
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})

}

func TestVMDeployWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	//slightly less than requirement
	gasLimit := vmOpGas - 1
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	_, err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		//the transfer should get back to the sender as the tx failed
		vm.ComputeExpectedBalance(senderBalance, big.NewInt(0), gasLimit, gasPrice))
	destinationAddressBytes, _ := hex.DecodeString("b473e38c40c60c8b22eb80fffff0717a98386bb129a7750ffe8091c77b85990e")

	assert.False(t, vm.AccountExists(accnts, destinationAddressBytes))
}
