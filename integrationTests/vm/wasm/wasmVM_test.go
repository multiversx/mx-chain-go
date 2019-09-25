package arwen

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

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

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, senderNonce, senderAddressBytes, senderBalance)

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

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "./fibonacci_ewasmified.wasm", b.N, 32)
}

func Benchmark_VmDeployWithCPUCalculateAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "./cpucalculate_ewasmified.wasm", b.N, 8000)
}

func Benchmark_VmDeployWithStringConcatAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "./stringconcat_ewasmified.wasm", b.N, 10000)
}

func runWASMVMBenchmark(tb testing.TB, fileSC string, numRun int, testingValue uint64) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(1)

	scCode, err := ioutil.ReadFile(fileSC)
	assert.Nil(tb, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     transferOnCalls,
		RcvAddr:   vm.CreateEmptyAddress().Bytes(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      scCodeString + "@" + hex.EncodeToString(factory.HeraWAVMVirtualMachine),
		Signature: nil,
		Challenge: nil,
	}

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(tb, ownerNonce, ownerAddressBytes, ownerBalance)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.HeraWAVMVirtualMachine)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(tb, err)

	_, err = accnts.Commit()
	assert.Nil(tb, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(10000000000))

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

	for i := 0; i < numRun; i++ {
		tx.Nonce = aliceNonce

		_ = txProc.ProcessTransaction(tx, round)

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

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, ownerNonce, ownerAddressBytes, ownerBalance)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.HeraWABTVirtualMachine)

	fmt.Println(hex.EncodeToString(scAddress))
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.HeraWABTVirtualMachine),
	)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	initAlice := big.NewInt(100000)
	tx = vm.CreateTopUpTx(aliceNonce, initAlice, scAddress, alice)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	aliceNonce++

	start := time.Now()
	nrTxs := 10000

	for i := 0; i < nrTxs; i++ {
		tx = vm.CreateTransferTx(aliceNonce, transferOnCalls, scAddress, alice, bob)

		err = txProc.ProcessTransaction(tx, round)
		assert.Nil(t, err)

		aliceNonce++
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", nrTxs, elapsedTime.String())

	finalAlice := big.NewInt(0).Sub(initAlice, big.NewInt(int64(nrTxs)*transferOnCalls.Int64()))
	assert.Equal(t, finalAlice.Uint64(), vm.GetIntValueFromSC(accnts, scAddress, "balance", alice).Uint64())
	finalBob := big.NewInt(int64(nrTxs) * transferOnCalls.Int64())
	assert.Equal(t, finalBob.Uint64(), vm.GetIntValueFromSC(accnts, scAddress, "balance", bob).Uint64())
}
