package wasm

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	vmLib "github.com/ElrondNetwork/elrond-go/core/libLocator"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, err := ioutil.ReadFile("./store.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.HeraWABTVirtualMachine),
	)

	config := vmLib.WASMLibLocation() + ",engine=wavm"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, senderNonce, senderAddressBytes, senderBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99999805)
	fmt.Printf("%s \n", hex.EncodeToString(expectedBalance.Bytes()))

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)
}

func TestVmDeployWithFibbonacciAndExecute(t *testing.T) {
	runWASMVMBenchmark(t, "./fibonacci_ewasmified.wasm", 100, 32)
}

func TestVmDeployWithCPUCalculateAndExecute(t *testing.T) {
	runWASMVMBenchmark(t, "./cpucalculate_ewasmified.wasm", 100, 8000)
}

func TestVmDeployWithStringConcatAndExecute(t *testing.T) {
	runWASMVMBenchmark(t, "./stringconcat_ewasmified.wasm", 100, 10000)
}

func runWASMVMBenchmark(t *testing.T, fileSC string, numRun int, testingValue uint64) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(1)

	scCode, err := ioutil.ReadFile(fileSC)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.HeraWABTVirtualMachine),
	)

	config := vmLib.WASMLibLocation() + ",engine=wavm"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, ownerNonce, ownerAddressBytes, ownerBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	scAddress, _ := hex.DecodeString("000000000000000002001a2983b179a480a60c4308da48f13b4480dbb4d33132")

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(10000000000))

	for i := 0; i < numRun; i++ {
		tx = &transaction.Transaction{
			Nonce:     aliceNonce,
			Value:     big.NewInt(0).SetUint64(testingValue),
			RcvAddr:   scAddress,
			SndAddr:   alice,
			GasPrice:  0,
			GasLimit:  5000,
			Data:      "benchmark",
			Signature: nil,
			Challenge: nil,
		}

		startTime := time.Now()
		err = txProc.ProcessTransaction(tx, round)
		elapsedTime := time.Since(startTime)
		fmt.Printf("time elapsed full process %s \n", elapsedTime.String())
		assert.Nil(t, err)

		_, err = accnts.Commit()
		assert.Nil(t, err)

		aliceNonce++
	}
}

func TestVmDeployWithTransferAndExecuteERC20(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(5)

	scCode, err := ioutil.ReadFile("./wrc20_ewasmified.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.HeraWABTVirtualMachine),
	)

	config := vmLib.WASMLibLocation() + ",engine=wabt"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, ownerNonce, ownerAddressBytes, ownerBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	scAddress, _ := hex.DecodeString("000000000000000002001a2983b179a480a60c4308da48f13b4480dbb4d33132")

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     big.NewInt(100000),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  0,
		GasLimit:  5000,
		Data:      "topUp@1",
		Signature: nil,
		Challenge: nil,
	}
	start := time.Now()
	err = txProc.ProcessTransaction(tx, round)
	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process topup %s \n", elapsedTime.String())
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	aliceNonce++

	start = time.Now()
	for i := 0; i < 10000; i++ {
		tx = &transaction.Transaction{
			Nonce:     aliceNonce,
			Value:     big.NewInt(5),
			RcvAddr:   scAddress,
			SndAddr:   alice,
			GasPrice:  0,
			GasLimit:  5000,
			Data:      "transfer@" + string(bob) + "@5",
			Signature: nil,
			Challenge: nil,
		}

		err = txProc.ProcessTransaction(tx, round)
		assert.Nil(t, err)

		aliceNonce++
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	elapsedTime = time.Since(start)
	fmt.Printf("time elapsed to process 10000 ERC20 transfers %s \n", elapsedTime.String())
}
