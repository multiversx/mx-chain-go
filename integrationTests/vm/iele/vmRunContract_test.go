package mockVM

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func TestRunWithTransferAndGasShouldRunSCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, _ := hex.DecodeString("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101")
	initialValueForInternalVariable := uint64(45)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)

	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		initialValueForInternalVariable,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes, _ := hex.DecodeString("195d84b4aec942d3534d2ad210b548f26776b8859b1fabdf8298d9ce0d973132")
	addValue := uint64(128)
	//contract call tx
	txRun := vm.CreateTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"add",
		addValue,
	)

	crossShardScrs, err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(crossShardScrs))

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(0).SetUint64(99979388)
	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+2,
		//2*gasLimit because we do 2 operations: deploy and call
		expectedBalance)

	expectedValueForVariable := big.NewInt(0).Add(big.NewInt(int64(initialValueForInternalVariable)), big.NewInt(int64(addValue)))
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		string(scCode),
		map[string]*big.Int{"a": expectedValueForVariable})
}

func TestRunWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, _ := hex.DecodeString("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101")
	initialValueForInternalVariable := uint64(45)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
	//deploy will transfer 0 and will succeed
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		initialValueForInternalVariable,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes, _ := hex.DecodeString("195d84b4aec942d3534d2ad210b548f26776b8859b1fabdf8298d9ce0d973132")
	addValue := uint64(128)
	//contract call tx that will feil with out of gas
	gasLimitFail := uint64(100)
	txRun := vm.CreateTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimitFail,
		"add",
		addValue,
	)

	crossShardScrs, err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(crossShardScrs))

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(0).SetUint64(99981547)
	//following operations happened: deploy and call, deploy succeed, call failed, transfer has been reverted, gas consumed
	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+2,
		expectedBalance)

	//value did not change, remained initial so the transfer did not happened
	expectedValueForVariable := big.NewInt(0).SetUint64(initialValueForInternalVariable)
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		big.NewInt(0),
		string(scCode),
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

	crossShardScrs, err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(crossShardScrs))

	_, err = accnts.Commit()
	assert.Nil(t, err)
}
