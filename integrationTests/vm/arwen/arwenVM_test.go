package arwen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, err := getBytecode("misc/fib_arwen.wasm")
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
		scCodeString+"@"+hex.EncodeToString(factory.ArwenVirtualMachine),
	)

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, senderNonce, senderAddressBytes, senderBalance)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99999699)
	fmt.Printf("%s \n", hex.EncodeToString(expectedBalance.Bytes()))

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)
}

func TestSCMoveBalanceBeforeSCDeploy(t *testing.T) {
	_ = logger.SetLogLevel("*:INFO,process/smartcontract:DEBUG")

	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(100000000)
	gasPrice := uint64(0)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, err := getBytecode("misc/fib_arwen.wasm")
	assert.Nil(t, err)
	scCodeString := hex.EncodeToString(scCode)

	txProc, accnts, blockChainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, ownerNonce, ownerAddressBytes, ownerBalance)
	scAddressBytes, _ := blockChainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.ArwenVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddressBytes))

	tx := vm.CreateTx(t,
		ownerAddressBytes,
		scAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"")

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	ownerNonce++
	tx = vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.ArwenVirtualMachine),
	)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := ownerBalance.Uint64() - 2*transferOnCalls.Uint64()
	vm.TestAccount(
		t,
		accnts,
		ownerAddressBytes,
		ownerNonce+1,
		big.NewInt(0).SetUint64(expectedBalance))

	expectedBalance = 2 * transferOnCalls.Uint64()
	vm.TestAccount(
		t,
		accnts,
		scAddressBytes,
		0,
		big.NewInt(0).SetUint64(expectedBalance))
}

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "misc/fib_arwen.wasm", b.N, 32, nil)
}

func Benchmark_VmDeployWithCPUCalculateAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "misc/cpucalculate_arwen.wasm", b.N, 8000, nil)
}

func Benchmark_VmDeployWithStringConcatAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "misc/stringconcat_arwen.wasm", b.N, 10000, nil)
}

func runWASMVMBenchmark(
	tb testing.TB,
	fileSC string,
	numRun int,
	testingValue uint64,
	gasSchedule map[string]map[string]uint64,
) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xffffffffffffffff)
	transferOnCalls := big.NewInt(1)

	scCode, err := getBytecode(fileSC)
	assert.Nil(tb, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     transferOnCalls,
		RcvAddr:   vm.CreateEmptyAddress().Bytes(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine),
		Signature: nil,
		Challenge: nil,
	}

	txProc, accnts, blockchainHook := vm.CreateTxProcessorArwenVMWithGasSchedule(tb, ownerNonce, ownerAddressBytes, ownerBalance, gasSchedule)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(tb, err)

	_, err = accnts.Commit()
	assert.Nil(tb, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(10000000000))

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     big.NewInt(0).SetUint64(testingValue),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  0,
		GasLimit:  gasLimit,
		Data:      "_main",
		Signature: nil,
		Challenge: nil,
	}

	for i := 0; i < numRun; i++ {
		tx.Nonce = aliceNonce

		_ = txProc.ProcessTransaction(tx)

		aliceNonce++
	}
}

func TestGasModel(t *testing.T) {
	_ = logger.SetLogLevel("*:INFO,process/smartcontract:DEBUG")

	gasSchedule, _ := core.LoadGasScheduleConfig("./gasSchedule.toml")

	totalOp := uint64(0)
	for _, opCodeClass := range gasSchedule {
		for _, opCode := range opCodeClass {
			totalOp += opCode
		}
	}
	fmt.Println("gasSchedule: " + big.NewInt(int64(totalOp)).String())
	fmt.Println("FIBONNACI 32 ")
	runWASMVMBenchmark(t, "misc/fib_arwen.wasm", 1, 32, gasSchedule)
	fmt.Println("CPUCALCULATE 8000 ")
	runWASMVMBenchmark(t, "misc/cpucalculate_arwen.wasm", 1, 8000, gasSchedule)
	fmt.Println("STRINGCONCAT 1000 ")
	runWASMVMBenchmark(t, "misc/stringconcat_arwen.wasm", 1, 10000, gasSchedule)
	fmt.Println("ERC20 ")
	deployWithTransferAndExecuteERC20(t, 2, gasSchedule)
	fmt.Println("ERC20 BIGINT")
	deployAndExecuteERC20WithBigInt(t, 2, gasSchedule)
}

func TestMultipleTimesERC20InBatches(t *testing.T) {
	for i := 0; i < 10; i++ {
		deployWithTransferAndExecuteERC20(t, 1000, nil)
	}
}

func deployWithTransferAndExecuteERC20(t *testing.T, numRun int, gasSchedule map[string]map[string]uint64) {
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)
	transferOnCalls := big.NewInt(5)

	scCode, err := getBytecode("erc20/wrc20_arwen_01.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)
	txProc, accnts, blockchainHook := vm.CreateTxProcessorArwenVMWithGasSchedule(t, ownerNonce, ownerAddressBytes, ownerBalance, gasSchedule)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.ArwenVirtualMachine),
	)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	initAlice := big.NewInt(100000)
	tx = vm.CreateTopUpTx(aliceNonce, initAlice, scAddress, alice)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	aliceNonce++

	start := time.Now()

	for i := 0; i < numRun; i++ {
		tx = vm.CreateTransferTx(aliceNonce, transferOnCalls, scAddress, alice, bob)

		err = txProc.ProcessTransaction(tx)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Nil(t, err)

		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", numRun, elapsedTime.String())

	_, err = accnts.Commit()
	assert.Nil(t, err)

	finalAlice := big.NewInt(0).Sub(initAlice, big.NewInt(int64(numRun)*transferOnCalls.Int64()))
	assert.Equal(t, finalAlice.Uint64(), vm.GetIntValueFromSC(gasSchedule, accnts, scAddress, "do_balance", alice).Uint64())
	finalBob := big.NewInt(int64(numRun) * transferOnCalls.Int64())
	assert.Equal(t, finalBob.Uint64(), vm.GetIntValueFromSC(gasSchedule, accnts, scAddress, "do_balance", bob).Uint64())
}

func TestWASMNamespacing(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xffffffffffffffff)
	transferOnCalls := big.NewInt(1)

	// This SmartContract had its imports modified after compilation, replacing
	// the namespace 'env' to 'ethereum'. If WASM namespacing is done correctly
	// by Arwen, then this SC should have no problem to call imported functions
	// (as if it were run by Ethereuem).
	scCode, err := getBytecode("misc/fib_ewasmified.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     transferOnCalls,
		RcvAddr:   vm.CreateEmptyAddress().Bytes(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine),
		Signature: nil,
		Challenge: nil,
	}

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, ownerNonce, ownerAddressBytes, ownerBalance)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	aliceInitialBalance := uint64(3000)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(0).SetUint64(aliceInitialBalance))

	testingValue := uint64(15)

	gasLimit = uint64(2000)

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     big.NewInt(0).SetUint64(testingValue),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      "main",
		Signature: nil,
		Challenge: nil,
	}

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)
}

func TestWASMMetering(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xffffffffffffffff)
	transferOnCalls := big.NewInt(1)

	scCode, err := getBytecode("misc/cpucalculate_arwen.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     transferOnCalls,
		RcvAddr:   vm.CreateEmptyAddress().Bytes(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine),
		Signature: nil,
		Challenge: nil,
	}

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithVMs(t, ownerNonce, ownerAddressBytes, ownerBalance)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	aliceInitialBalance := uint64(3000)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(0).SetUint64(aliceInitialBalance))

	testingValue := uint64(15)

	gasLimit = uint64(2000)

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     big.NewInt(0).SetUint64(testingValue),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      "_main",
		Signature: nil,
		Challenge: nil,
	}

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	expectedBalance := big.NewInt(2090)
	expectedNonce := uint64(1)

	actualBalanceBigInt := vm.TestAccount(
		t,
		accnts,
		alice,
		expectedNonce,
		expectedBalance)

	actualBalance := actualBalanceBigInt.Uint64()

	consumedGasValue := aliceInitialBalance - actualBalance - testingValue

	assert.Equal(t, 895, int(consumedGasValue))
}

func TestMultipleTimesERC20BigIntInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	for i := 0; i < 10; i++ {
		deployAndExecuteERC20WithBigInt(t, 1000, nil)
	}
}

func deployAndExecuteERC20WithBigInt(t *testing.T, numRun int, gasSchedule map[string]map[string]uint64) {
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)
	transferOnCalls := big.NewInt(5)

	scCode, err := getBytecode("erc20/wrc20_arwen_02.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)
	txProc, accnts, blockchainHook := vm.CreateTxProcessorArwenVMWithGasSchedule(t, ownerNonce, ownerAddressBytes, ownerBalance, gasSchedule)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.ArwenVirtualMachine)+"@"+hex.EncodeToString(ownerBalance.Bytes()),
	)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, initAlice, scAddress, ownerAddressBytes, alice)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	start := time.Now()

	for i := 0; i < numRun; i++ {
		tx = vm.CreateTransferTokenTx(aliceNonce, transferOnCalls, scAddress, alice, bob)

		err = txProc.ProcessTransaction(tx)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Nil(t, err)

		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", numRun, elapsedTime.String())

	_, err = accnts.Commit()
	assert.Nil(t, err)

	finalAlice := big.NewInt(0).Sub(initAlice, big.NewInt(int64(numRun)*transferOnCalls.Int64()))
	assert.Equal(t, finalAlice.Uint64(), vm.GetIntValueFromSC(gasSchedule, accnts, scAddress, "balanceOf", alice).Uint64())
	finalBob := big.NewInt(int64(numRun) * transferOnCalls.Int64())
	assert.Equal(t, finalBob.Uint64(), vm.GetIntValueFromSC(gasSchedule, accnts, scAddress, "balanceOf", bob).Uint64())
}

func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = rand.Read(r)
	return r
}

func createTestAddresses(numAddresses uint64) [][]byte {
	testAccounts := make([][]byte, numAddresses)

	for i := uint64(0); i < numAddresses; i++ {
		acc := generateRandomByteArray(32)
		testAccounts[i] = append(testAccounts[i], acc...)
	}

	return testAccounts
}

func TestJurnalizingAndTimeToProcessChange(t *testing.T) {
	// Only a test to benchmark jurnalizing and getting data from trie
	t.Skip()

	numRun := 1000
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)
	transferOnCalls := big.NewInt(5)

	scCode, err := getBytecode("erc20/wrc20_arwen_02.wasm")
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)
	txProc, accnts, blockchainHook := vm.CreateTxProcessorArwenVMWithGasSchedule(t, ownerNonce, ownerAddressBytes, ownerBalance, nil)
	scAddress, _ := blockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString+"@"+hex.EncodeToString(factory.ArwenVirtualMachine)+"@"+hex.EncodeToString(ownerBalance.Bytes()),
	)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	testAddresses := createTestAddresses(20000)
	fmt.Println("done")

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, initAlice, scAddress, ownerAddressBytes, alice)

	err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	for j := 0; j < 20; j++ {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			tx = vm.CreateTransferTokenTx(aliceNonce, transferOnCalls, scAddress, alice, testAddresses[j*1000+i])

			err = txProc.ProcessTransaction(tx)
			if err != nil {
				assert.Nil(t, err)
			}
			assert.Nil(t, err)

			aliceNonce++
		}

		elapsedTime := time.Since(start)
		fmt.Printf("time elapsed to process 1000 ERC20 transfers %s \n", elapsedTime.String())

		_, err = accnts.Commit()
		assert.Nil(t, err)
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	start := time.Now()

	for i := 0; i < numRun; i++ {
		tx = vm.CreateTransferTokenTx(aliceNonce, transferOnCalls, scAddress, alice, testAddresses[i])

		err = txProc.ProcessTransaction(tx)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Nil(t, err)

		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", numRun, elapsedTime.String())

	_, err = accnts.Commit()
	assert.Nil(t, err)
}

func getBytecode(relativePath string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(".", "testdata", relativePath))
}
