package vmRunContract

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/iele"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestRunWithTransferAndGasShouldRunSCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101@%s@0000@%X",
		hex.EncodeToString(factory.IELEVirtualMachine), initialValueForInternalVariable)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	err := iele.DeployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		scCode,
		testContext.TxProcessor,
		testContext.Accounts,
	)
	assert.NoError(t, err)

	destinationAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)
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

	_, err = testContext.TxProcessor.ProcessTransaction(txRun)
	assert.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(0).SetUint64(99999791)
	vm.TestAccount(
		t,
		testContext.Accounts,
		senderAddressBytes,
		senderNonce+2,
		//2*gasLimit because we do 2 operations: deploy and call
		expectedBalance)

	expectedValueForVariable := big.NewInt(0).Add(big.NewInt(int64(initialValueForInternalVariable)), big.NewInt(int64(addValue)))
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		testContext.Accounts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": expectedValueForVariable})
}

func TestRunWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101@%s@0000@%X",
		hex.EncodeToString(factory.IELEVirtualMachine), initialValueForInternalVariable)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	//deploy will transfer 0 and will succeed
	err := iele.DeployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		scCode,
		testContext.TxProcessor,
		testContext.Accounts,
	)
	assert.NoError(t, err)

	destinationAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)
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

	_, err = testContext.TxProcessor.ProcessTransaction(txRun)
	assert.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(0).SetUint64(99999851)
	//following operations happened: deploy and call, deploy succeed, call failed, transfer has been reverted, gas consumed
	vm.TestAccount(
		t,
		testContext.Accounts,
		senderAddressBytes,
		senderNonce+2,
		expectedBalance)

	//value did not change, remained initial so the transfer did not happened
	expectedValueForVariable := big.NewInt(0).SetUint64(initialValueForInternalVariable)
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		testContext.Accounts,
		big.NewInt(0),
		scCode,
		map[string]*big.Int{"a": expectedValueForVariable})
}
