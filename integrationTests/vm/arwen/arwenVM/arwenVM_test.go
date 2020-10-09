package arwenVM

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("arwenVMtest")

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(1000)
	transferOnCalls := big.NewInt(50)

	scCode := arwen.GetSCCode("../testdata/misc/fib_arwen.wasm")

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode),
	)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(99999699)
	fmt.Printf("%s \n", hex.EncodeToString(expectedBalance.Bytes()))

	vm.TestAccount(
		t,
		testContext.Accounts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)
}

func TestSCMoveBalanceBeforeSCDeploy(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(100000000)
	gasPrice := uint64(0)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode := arwen.GetSCCode("../testdata/misc/fib_arwen.wasm")

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(ownerNonce, ownerAddressBytes, ownerBalance)
	defer testContext.Close()

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.ArwenVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddressBytes))

	tx := vm.CreateTx(t,
		ownerAddressBytes,
		scAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"")

	_, err := testContext.TxProcessor.ProcessTransaction(tx)

	require.Equal(t, process.ErrFailedTransaction, err)
	require.Equal(t, process.ErrAccountNotPayable, testContext.GetLatestError())
	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+1,
		big.NewInt(0).Set(ownerBalance))

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	tx = vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress(),
		ownerNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode),
	)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := ownerBalance.Uint64() - transferOnCalls.Uint64()
	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+2,
		big.NewInt(0).SetUint64(expectedBalance))

	expectedBalance = transferOnCalls.Uint64()
	vm.TestAccount(
		t,
		testContext.Accounts,
		scAddressBytes,
		0,
		big.NewInt(0).SetUint64(expectedBalance))
}

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/fib_arwen.wasm", b.N, 32, nil)
}

func Benchmark_VmDeployWithCPUCalculateAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cpucalculate_arwen.wasm", b.N, 8000, nil)
}

func Benchmark_VmDeployWithStringConcatAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/stringconcat_arwen.wasm", b.N, 10000, nil)
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

	scCode := arwen.GetSCCode(fileSC)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     new(big.Int).Set(transferOnCalls),
		RcvAddr:   vm.CreateEmptyAddress(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte(arwen.CreateDeployTxData(scCode)),
		Signature: nil,
	}

	testContext := vm.CreateTxProcessorArwenVMWithGasSchedule(ownerNonce, ownerAddressBytes, ownerBalance, gasSchedule)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(tb, err)
	require.Nil(tb, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(10000000000))

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     new(big.Int).Set(big.NewInt(0).SetUint64(testingValue)),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  0,
		GasLimit:  gasLimit,
		Data:      []byte("_main"),
		Signature: nil,
	}

	for i := 0; i < numRun; i++ {
		tx.Nonce = aliceNonce

		_, _ = testContext.TxProcessor.ProcessTransaction(tx)

		aliceNonce++
	}
}

func TestGasModel(t *testing.T) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../gasSchedule.toml")

	totalOp := uint64(0)
	for _, opCodeClass := range gasSchedule {
		for _, opCode := range opCodeClass {
			totalOp += opCode
		}
	}
	fmt.Println("gasSchedule: " + big.NewInt(int64(totalOp)).String())
	fmt.Println("FIBONNACI 32 ")
	runWASMVMBenchmark(t, "../testdata/misc/fib_arwen.wasm", 1, 32, gasSchedule)
	fmt.Println("CPUCALCULATE 8000 ")
	runWASMVMBenchmark(t, "../testdata/misc/cpucalculate_arwen.wasm", 1, 8000, gasSchedule)
	fmt.Println("STRINGCONCAT 1000 ")
	runWASMVMBenchmark(t, "../testdata/misc/stringconcat_arwen.wasm", 1, 10000, gasSchedule)
	fmt.Println("ERC20 BIGINT")
	deployAndExecuteERC20WithBigInt(t, 2, gasSchedule)
}

func TestWASMMetering(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xffffffffffffffff)
	transferOnCalls := big.NewInt(1)

	scCode := arwen.GetSCCode("../testdata/misc/cpucalculate_arwen.wasm")

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     new(big.Int).Set(transferOnCalls),
		RcvAddr:   vm.CreateEmptyAddress(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte(arwen.CreateDeployTxData(scCode)),
		Signature: nil,
	}

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(ownerNonce, ownerAddressBytes, ownerBalance)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	aliceInitialBalance := uint64(3000)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(0).SetUint64(aliceInitialBalance))

	testingValue := uint64(15)

	gasLimit = uint64(500)

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     new(big.Int).Set(big.NewInt(0).SetUint64(testingValue)),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte("_main"),
		Signature: nil,
	}

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	expectedBalance := big.NewInt(2615)
	expectedNonce := uint64(1)

	actualBalanceBigInt := vm.TestAccount(
		t,
		testContext.Accounts,
		alice,
		expectedNonce,
		expectedBalance)

	actualBalance := actualBalanceBigInt.Uint64()

	consumedGasValue := aliceInitialBalance - actualBalance - testingValue

	require.Equal(t, 370, int(consumedGasValue))
}

func TestMultipleTimesERC20BigIntInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	for i := 0; i < 3; i++ {
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

	scCode := arwen.GetSCCode("../testdata/erc20-c-03/wrc20_arwen.wasm")

	testContext := vm.CreateTxProcessorArwenVMWithGasSchedule(ownerNonce, ownerAddressBytes, ownerBalance, gasSchedule)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(testContext.Accounts, bob, 0, big.NewInt(1000000))

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, initAlice, scAddress, ownerAddressBytes, alice)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	start := time.Now()

	for i := 0; i < numRun; i++ {
		tx = vm.CreateTransferTokenTx(aliceNonce, transferOnCalls, scAddress, alice, bob)

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())
		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", numRun, elapsedTime.String())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	finalAlice := big.NewInt(0).Sub(initAlice, big.NewInt(int64(numRun)*transferOnCalls.Int64()))
	require.Equal(t, finalAlice.Uint64(), vm.GetIntValueFromSC(gasSchedule, testContext.Accounts, scAddress, "balanceOf", alice).Uint64())
	finalBob := big.NewInt(int64(numRun) * transferOnCalls.Int64())
	require.Equal(t, finalBob.Uint64(), vm.GetIntValueFromSC(gasSchedule, testContext.Accounts, scAddress, "balanceOf", bob).Uint64())
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

func TestJournalizingAndTimeToProcessChange(t *testing.T) {
	// Only a test to benchmark jurnalizing and getting data from trie
	t.Skip()

	numRun := 1000
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)
	transferOnCalls := big.NewInt(5)

	scCode := arwen.GetSCCode("../testdata/erc20-c-03/wrc20_arwen.wasm")

	testContext := vm.CreateTxProcessorArwenVMWithGasSchedule(ownerNonce, ownerAddressBytes, ownerBalance, nil)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode)+"@00"+hex.EncodeToString(ownerBalance.Bytes()),
	)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(testContext.Accounts, bob, 0, big.NewInt(1000000))

	testAddresses := createTestAddresses(2000000)
	fmt.Println("done")

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, initAlice, scAddress, ownerAddressBytes, alice)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	for j := 0; j < 2000; j++ {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			tx = vm.CreateTransferTokenTx(aliceNonce, transferOnCalls, scAddress, alice, testAddresses[j*1000+i])

			_, err = testContext.TxProcessor.ProcessTransaction(tx)
			require.Nil(t, err)
			require.Nil(t, testContext.GetLatestError())
			aliceNonce++
		}

		elapsedTime := time.Since(start)
		fmt.Printf("time elapsed to process 1000 ERC20 transfers %s \n", elapsedTime.String())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)
	}

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	start := time.Now()
	for i := 0; i < numRun; i++ {
		tx = vm.CreateTransferTokenTx(aliceNonce, transferOnCalls, scAddress, alice, testAddresses[i])

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", numRun, elapsedTime.String())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)
}

func TestExecuteTransactionAndTimeToProcessChange(t *testing.T) {
	// Only a test to benchmark transaction processing
	t.Skip()

	testMarshalizer := &marshal.JsonMarshalizer{}
	testHasher := sha256.Sha256{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	pubkeyConv, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	accnts := vm.CreateInMemoryShardAccountsDB()
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pubkeyConv,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: make(map[string]struct{}),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	feeHandler := &mock.FeeHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(10)
		},
	}
	numRun := 20000
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	transferOnCalls := big.NewInt(5)

	_, _ = vm.CreateAccount(accnts, ownerAddressBytes, ownerNonce, ownerBalance)
	argsNewTxProcessor := processTransaction.ArgsNewTxProcessor{
		Accounts:         accnts,
		Hasher:           testHasher,
		PubkeyConv:       pubkeyConv,
		Marshalizer:      testMarshalizer,
		SignMarshalizer:  testMarshalizer,
		ShardCoordinator: shardCoordinator,
		ScProcessor:      &mock.SCProcessorMock{},
		TxFeeHandler:     &mock.UnsignedTxHandlerMock{},
		TxTypeHandler:    txTypeHandler,
		EconomicsFee:     &mock.FeeHandlerStub{},
		ReceiptForwarder: &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:       smartContract.NewArgumentParser(),
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		EpochNotifier:    forking.NewGenericEpochNotifier(),
	}
	txProc, _ := processTransaction.NewTxProcessor(argsNewTxProcessor)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	testAddresses := createTestAddresses(uint64(numRun))
	fmt.Println("done")

	gasLimit := feeHandler.ComputeMoveBalanceFeeCalled(&transaction.Transaction{}).Uint64()
	initAlice := big.NewInt(100000)
	tx := vm.CreateMoveBalanceTx(ownerNonce, initAlice, ownerAddressBytes, alice, gasLimit)

	_, err := txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	for j := 0; j < 20; j++ {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			tx = vm.CreateMoveBalanceTx(aliceNonce, transferOnCalls, alice, testAddresses[j*1000+i], gasLimit)

			_, err = txProc.ProcessTransaction(tx)
			if err != nil {
				assert.Nil(t, err)
			}
			assert.Nil(t, err)

			aliceNonce++
		}

		elapsedTime := time.Since(start)
		fmt.Printf("time elapsed to process 1000 move balances %s \n", elapsedTime.String())

		_, err = accnts.Commit()
		assert.Nil(t, err)
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	start := time.Now()

	for i := 0; i < numRun; i++ {
		tx = vm.CreateMoveBalanceTx(aliceNonce, transferOnCalls, alice, testAddresses[i], gasLimit)

		_, err = txProc.ProcessTransaction(tx)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Nil(t, err)

		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d move balances %s \n", numRun, elapsedTime.String())

	_, err = accnts.Commit()
	assert.Nil(t, err)
}

func TestAndCatchTrieError(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)

	scCode := arwen.GetSCCode("../testdata/erc20-c-03/wrc20_arwen.wasm")

	testContext := vm.CreateTxProcessorArwenVMWithGasSchedule(ownerNonce, ownerAddressBytes, ownerBalance, nil)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())
	ownerNonce++

	numAccounts := 100
	testAddresses := createTestAddresses(uint64(numAccounts))
	// ERD Minting
	for _, testAddress := range testAddresses {
		_, _ = vm.CreateAccount(testContext.Accounts, testAddress, 0, big.NewInt(1000000))
	}

	accumulateAddress := createTestAddresses(1)[0]

	// ERC20 Minting
	erc20value := big.NewInt(100)
	for _, testAddress := range testAddresses {
		tx = vm.CreateTransferTokenTx(ownerNonce, erc20value, scAddress, ownerAddressBytes, testAddress)
		ownerNonce++

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())
	}

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	receiverAddresses := createTestAddresses(uint64(numAccounts))

	transferNonce := uint64(0)
	// Transfer among each person revert and retry
	for i := 0; i < 51; i++ {
		rootHash, _ := testContext.Accounts.RootHash()

		for index, testAddress := range testAddresses {
			tx = vm.CreateTransferTokenTx(transferNonce, erc20value, scAddress, testAddress, receiverAddresses[index])

			snapShot := testContext.Accounts.JournalLen()
			_, _ = testContext.TxProcessor.ProcessTransaction(tx)
			require.Nil(t, testContext.GetLatestError())

			if index%5 == 0 {
				errRevert := testContext.Accounts.RevertToSnapshot(snapShot)
				if errRevert != nil {
					log.Warn("revert to snapshot", "error", errRevert.Error())
				}
			}
		}

		tx = vm.CreateTransferTokenTx(ownerNonce, erc20value, scAddress, ownerAddressBytes, accumulateAddress)
		require.NotNil(t, tx)

		newRootHash, errNewRh := testContext.Accounts.Commit()
		require.Nil(t, errNewRh)

		for index, testAddress := range receiverAddresses {
			if index%5 == 0 {
				continue
			}

			tx = vm.CreateTransferTokenTx(transferNonce, erc20value, scAddress, testAddress, testAddresses[index])

			snapShot := testContext.Accounts.JournalLen()
			_, _ = testContext.TxProcessor.ProcessTransaction(tx)
			require.Nil(t, testContext.GetLatestError())

			if index%5 == 0 {
				errRevert := testContext.Accounts.RevertToSnapshot(snapShot)
				if errRevert != nil {
					log.Warn("revert to snapshot", "error", errRevert.Error())
				}
			}
		}

		extraNewRootHash, _ := testContext.Accounts.Commit()
		require.Nil(t, err)
		log.Info("finished a set - commit and recreate trie", "index", i)
		if i%10 == 5 {
			testContext.Accounts.PruneTrie(extraNewRootHash, data.NewRoot)
			_ = testContext.Accounts.RecreateTrie(rootHash)
			continue
		}

		ownerNonce++
		transferNonce++
		testContext.Accounts.PruneTrie(rootHash, data.OldRoot)
		testContext.Accounts.PruneTrie(newRootHash, data.OldRoot)
	}
}
