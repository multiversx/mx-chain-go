package mockVM

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101@%X",
		initialValueForInternalVariable)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99922044)
	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)
	destinationAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		string(scCode),
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVMDeployWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	//less than requirement
	gasLimit := uint64(100)
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101@%X",
		initialValueForInternalVariable)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99999900)
	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		//the transfer should get back to the sender as the tx failed
		expectedBalance)
	destinationAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	assert.False(t, vm.AccountExists(accnts, destinationAddressBytes))
}
