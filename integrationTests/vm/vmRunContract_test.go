package vm

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

//TODO add integration and unit tests with generating and broadcasting transaction with empty recv address

func TestRunSCWithoutTransferShouldRunSCCode(t *testing.T) {
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

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	addValue := uint64(128)
	//contract call tx
	txRun := createTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"Add",
		addValue,
	)

	err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+2,
		computeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))

	expectedValueForVariable := big.NewInt(0).Add(big.NewInt(int64(initialValueForInternalVariable)), big.NewInt(int64(addValue)))
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": expectedValueForVariable})
}

func TestRunSCWithTransferShouldRunSCCode(t *testing.T) {
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

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)
	//deploy will transfer 0
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	addValue := uint64(128)
	//contract call tx
	txRun := createTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"Add",
		addValue,
	)

	err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+2,
		computeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))

	expectedValueForVariable := big.NewInt(0).Add(big.NewInt(int64(initialValueForInternalVariable)), big.NewInt(int64(addValue)))
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": expectedValueForVariable})
}

func TestRunWithTransferAndGasShouldRunSCCode(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)
	//deploy will transfer 0
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		scCode,
		initialValueForInternalVariable,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	addValue := uint64(128)
	//contract call tx
	txRun := createTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"Add",
		addValue,
	)

	err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+2,
		//2*gasLimit because we do 2 operations: deploy and call
		computeExpectedBalance(senderBalance, transferOnCalls, 2*gasLimit, gasPrice))

	expectedValueForVariable := big.NewInt(0).Add(big.NewInt(int64(initialValueForInternalVariable)), big.NewInt(int64(addValue)))
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": expectedValueForVariable})
}

func TestRunWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas - 1
	transferOnCalls := big.NewInt(50)

	scCode := "mocked code, not taken into account"
	initialValueForInternalVariable := uint64(45)

	txProc, accnts := createPreparedTxProcessorAndAccounts(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)
	//deploy will transfer 0 and will succeed
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		vmOpGas,
		scCode,
		initialValueForInternalVariable,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	addValue := uint64(128)
	//contract call tx
	txRun := createTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"Add",
		addValue,
	)

	err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	testAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+2,
		//following operations happened: deploy and call, deploy succeed, call failed, transfer has been reverted, gas consumed
		computeExpectedBalance(senderBalance, big.NewInt(0), vmOpGas+gasLimit, gasPrice))

	//value did not change, remained initial
	expectedValueForVariable := big.NewInt(0).SetUint64(initialValueForInternalVariable)
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		//transfer did not happened
		big.NewInt(0),
		scCode,
		map[string]*big.Int{"a": expectedValueForVariable})
}

func deployContract(
	t *testing.T,
	senderAddressBytes []byte,
	senderNonce uint64,
	transferOnCalls *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCode string,
	initialValueForInternalVariable uint64,
	round uint32,
	txProc process.TransactionProcessor,
	accnts state.AccountsAdapter,
) {

	//contract creation tx
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

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)
}
