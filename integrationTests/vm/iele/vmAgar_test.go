package mockVM

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "agarioV2.hex"

func TestDeployAgarioContract(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
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

	destinationAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		big.NewInt(0),
		string(scCode),
		make(map[string]*big.Int))
}

func TestAgarioContractTopUpShouldWork(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
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
	scAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	userAddress := []byte("10000000000000000000000000000000")
	userNonce := uint64(10)
	userBalance := big.NewInt(100000000)
	_ = vm.CreateAccount(accnts, userAddress, userNonce, userBalance)
	_, _ = accnts.Commit()

	//balanceOf should return 0 for userAddress
	assert.Equal(t, big.NewInt(0), getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))

	transfer := big.NewInt(123456)
	data := "topUp"
	//contract call tx
	txRun := vm.CreateTx(
		t,
		userAddress,
		scAddressBytes,
		userNonce,
		transfer,
		gasPrice,
		gasLimit,
		data,
	)

	err = txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	assert.Equal(t, transfer, getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))
}

func TestAgarioContractTopUpAnfWithdrawShouldWork(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint32(444)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
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
	scAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	userAddress := []byte("10000000000000000000000000000000")
	userNonce := uint64(10)
	userBalance := big.NewInt(100000000)
	_ = vm.CreateAccount(accnts, userAddress, userNonce, userBalance)
	_, _ = accnts.Commit()

	//balanceOf should return 0 for userAddress
	assert.Equal(t, big.NewInt(0), getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))

	transfer := big.NewInt(123456)
	data := "topUp"
	//contract call tx
	txRun := vm.CreateTx(
		t,
		userAddress,
		scAddressBytes,
		userNonce,
		transfer,
		gasPrice,
		gasLimit,
		data,
	)

	err = txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	assert.Equal(t, transfer, getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))

	//withdraw
	withdraw := big.NewInt(49999)
	data = fmt.Sprintf("withdraw@%X", withdraw)
	//contract call tx
	txRun = vm.CreateTx(
		t,
		userAddress,
		scAddressBytes,
		userNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit*30,
		data,
	)

	err = txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	newValue := big.NewInt(0).Set(transfer)
	newValue.Sub(newValue, withdraw)
	assert.Equal(t, newValue, getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))
}

func getIntValueFromSC(accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) *big.Int {
	ieleVM, _ := vm.CreateVMAndBlockchainHook(accnts)
	scgd, _ := smartContract.NewSCDataGetter(ieleVM)

	returnedVals, _ := scgd.Get(scAddressBytes, funcName, args...)
	return big.NewInt(0).SetBytes(returnedVals)
}
