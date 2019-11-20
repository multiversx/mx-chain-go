package mockVM

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestRunWithTransferAndGasShouldRunSCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101@%s@%X",
		hex.EncodeToString(factory.IELEVirtualMachine), initialValueForInternalVariable)

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, senderNonce, senderAddressBytes, senderBalance)

	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		scCode,
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes, _ := blockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)
	addValue := uint64(128)
	data := fmt.Sprintf("add@00%X", addValue)
	//contract call tx
	txRun := vm.CreateTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		data,
	)

	err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(0).SetUint64(99999791)
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
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101@%s@%X",
		hex.EncodeToString(factory.IELEVirtualMachine), initialValueForInternalVariable)

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, senderNonce, senderAddressBytes, senderBalance)
	//deploy will transfer 0 and will succeed
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes, _ := blockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)
	addValue := uint64(128)
	data := fmt.Sprintf("add@00%X", addValue)
	//contract call tx that will fail with out of gas
	gasLimitFail := uint64(10)
	txRun := vm.CreateTx(
		t,
		senderAddressBytes,
		destinationAddressBytes,
		senderNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimitFail,
		data,
	)

	err := txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(0).SetUint64(99999851)
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
	tb testing.TB,
	senderAddressBytes []byte,
	senderNonce uint64,
	transferOnCalls *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCode string,
	round uint64,
	txProc process.TransactionProcessor,
	accnts state.AccountsAdapter,
) {

	//contract creation tx
	tx := vm.CreateTx(
		tb,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(tb, err)

	_, err = accnts.Commit()
	assert.Nil(tb, err)
}
