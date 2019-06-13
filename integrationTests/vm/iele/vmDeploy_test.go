package mockVM

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/vm"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithoutTransferShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := vm.CreateDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(0)

	scCode, _ := hex.DecodeString("0000003B6302690003616464690004676574416700000001616101550468000100016161015406010A6161015506F6000068000200006161005401F6000101")
	initialValueForInternalVariable := uint64(45)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		string(scCode),
		initialValueForInternalVariable,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes := vm.ComputeSCDestinationAddressBytes(senderNonce, senderAddressBytes)
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		string(scCode),
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}
