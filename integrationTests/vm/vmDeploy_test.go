package vm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithoutTransferShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(0)
	senderAddressBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(0)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := createTx(
		t,
		senderAddressBytes,
		createEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		computeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVmDeployWithTransferShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(0)
	senderAddressBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := createTx(
		t,
		senderAddressBytes,
		createEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		computeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	//equal with requirement
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := createTx(
		t,
		senderAddressBytes,
		createEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		computeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})

}

func TestVMDeployWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	//slightly less than requirement
	gasLimit := vmOpGas - 1
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	tx := createTx(
		t,
		senderAddressBytes,
		createEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
	)

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		//the transfer should get back to the sender as the tx failed
		computeExpectedBalance(senderBalance, big.NewInt(0), gasLimit, gasPrice))
	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)

	assert.False(t, accountExists(accnts, destinationAddressBytes))
}
