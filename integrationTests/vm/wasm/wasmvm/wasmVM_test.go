package wasmvm

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/multiversx/mx-chain-vm-common-go/txDataBuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	syncDisabled "github.com/multiversx/mx-chain-go/process/sync/disabled"
	processTransaction "github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

var log = logger.GetOrCreate("wasmVMtest")

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(1962)
	transferOnCalls := big.NewInt(50)

	scCode := wasm.GetSCCode("../testdata/misc/fib_wasm/output/fib_wasm.wasm")

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode),
	)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(99997990)

	vm.TestAccount(
		t,
		testContext.Accounts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)
}

func TestVmSCDeployFactory(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000)
	transferOnCalls := big.NewInt(0)

	scCode := wasm.GetSCCode("../testdata/misc/deploy-two-contracts.wasm")

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode),
	)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.WasmVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddressBytes))

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	for i := 0; i < 5; i++ {
		senderNonce++
		tx = vm.CreateTx(senderAddressBytes, scAddressBytes, senderNonce, big.NewInt(0), gasPrice, gasLimit, "deployContract")

		returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Equal(t, vmcommon.Ok, returnCode)
	}

	senderNonce++
	tx = vm.CreateTx(senderAddressBytes, scAddressBytes, senderNonce, big.NewInt(0), gasPrice, gasLimit, "deployTwoContracts")

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
}

func TestSCMoveBalanceBeforeSCDeployV1(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode := wasm.GetSCCode("../testdata/misc/fib_wasm/output/fib_wasm.wasm")

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
			SCProcessorV2EnableEpoch:       integrationTests.UnreachableEpoch,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.WasmVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddressBytes))

	tx := vm.CreateTx(
		ownerAddressBytes,
		scAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"")

	_, err = testContext.TxProcessor.ProcessTransaction(tx)

	require.Equal(t, process.ErrFailedTransaction, err)
	require.Equal(t, process.ErrAccountNotPayable, testContext.GetCompositeTestError())
	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+1,
		big.NewInt(0).Sub(ownerBalance, big.NewInt(1)))

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	tx = vm.CreateTx(
		ownerAddressBytes,
		vm.CreateEmptyAddress(),
		ownerNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+2,
		big.NewInt(99997989))

	vm.TestAccount(
		t,
		testContext.Accounts,
		scAddressBytes,
		0,
		transferOnCalls)
}

func TestSCMoveBalanceBeforeSCDeploy(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode := wasm.GetSCCode("../testdata/misc/fib_wasm/output/fib_wasm.wasm")

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.WasmVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddressBytes))

	tx := vm.CreateTx(
		ownerAddressBytes,
		scAddressBytes,
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		"")

	_, err = testContext.TxProcessor.ProcessTransaction(tx)

	require.Equal(t, process.ErrFailedTransaction, err)
	require.Equal(t, process.ErrAccountNotPayable, testContext.GetCompositeTestError())
	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+1,
		big.NewInt(0).Sub(ownerBalance, big.NewInt(1)))

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	tx = vm.CreateTx(
		ownerAddressBytes,
		vm.CreateEmptyAddress(),
		ownerNonce+1,
		transferOnCalls,
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+2,
		big.NewInt(99899949))

	vm.TestAccount(
		t,
		testContext.Accounts,
		scAddressBytes,
		0,
		transferOnCalls)
}

func TestWASMMetering(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xfffffffffffffff)
	transferOnCalls := big.NewInt(1)

	scCode := wasm.GetSCCode("../testdata/misc/cpucalculate_wasm/output/cpucalculate.wasm")

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     new(big.Int).Set(transferOnCalls),
		RcvAddr:   vm.CreateEmptyAddress(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte(wasm.CreateDeployTxData(scCode)),
		Signature: nil,
	}

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	aliceInitialBalance := uint64(3000000)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(0).SetUint64(aliceInitialBalance))

	testingValue := uint64(15)

	gasLimit = uint64(500000)

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     new(big.Int).Set(big.NewInt(0).SetUint64(testingValue)),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte("cpuCalculate"),
		Signature: nil,
	}

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	expectedBalance := big.NewInt(2499985)
	expectedNonce := uint64(1)

	actualBalanceBigInt := vm.TestAccount(
		t,
		testContext.Accounts,
		alice,
		expectedNonce,
		expectedBalance)

	actualBalance := actualBalanceBigInt.Uint64()

	consumedGasValue := aliceInitialBalance - actualBalance - testingValue

	require.Equal(t, 500000, int(consumedGasValue))
}

func TestMultipleTimesERC20BigIntInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	durations, err := DeployAndExecuteERC20WithBigInt(3, 1000, gasSchedule, "../testdata/erc20-c-03/wrc20_wasm.wasm", "transferToken")
	require.Nil(t, err)
	displayBenchmarksResults(durations)

	durations, err = DeployAndExecuteERC20WithBigInt(3, 1000, nil, "../testdata/erc20-c-03/wrc20_wasm.wasm", "transferToken")
	require.Nil(t, err)
	displayBenchmarksResults(durations)
}

func TestMultipleTimesERC20RustBigIntInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	durations, err := DeployAndExecuteERC20WithBigInt(3, 1000, gasSchedule, "../testdata/erc20-c-03/rust-simple-erc20.wasm", "transfer")
	require.Nil(t, err)
	displayBenchmarksResults(durations)

	durations, err = DeployAndExecuteERC20WithBigInt(3, 1000, nil, "../testdata/erc20-c-03/rust-simple-erc20.wasm", "transfer")
	require.Nil(t, err)
	displayBenchmarksResults(durations)
}

func displayBenchmarksResults(durations []time.Duration) {
	if len(durations) == 0 {
		return
	}

	minTime := time.Hour
	maxTime := time.Duration(0)
	sumTime := time.Duration(0)
	for _, d := range durations {
		sumTime += d
		if minTime > d {
			minTime = d
		}
		if maxTime < d {
			maxTime = d
		}
	}

	log.Info("execution complete",
		"total time", sumTime,
		"average time", sumTime/time.Duration(len(durations)),
		"total erc20 batches", len(durations),
		"min time", minTime,
		"max time", maxTime,
	)
}

func TestDeployERC20WithNotEnoughGasShouldReturnOutOfGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(1000000000000000)
	gasPrice := uint64(1)

	scCode := wasm.GetSCCode("../testdata/erc20-c-03/wrc20_wasm.wasm")

	testContext, err := vm.CreateTxProcessorWasmVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		config.EnableEpochs{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		2_800_000,
		wasm.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.UserError)
}

func TestJournalizingAndTimeToProcessChange(t *testing.T) {
	t.Skip("Only a test to benchmark jurnalizing and getting data from trie")

	numRun := 1000
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)
	transferOnCalls := big.NewInt(5)

	scCode := wasm.GetSCCode("../testdata/erc20-c-03/wrc20_wasm.wasm")

	testContext, err := vm.CreateTxProcessorWasmVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		nil,
		config.EnableEpochs{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode)+"@00"+hex.EncodeToString(ownerBalance.Bytes()),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(testContext.Accounts, bob, 0, big.NewInt(1000000))

	testAddresses := createTestAddresses(2000000)
	fmt.Println("done")

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, "transferToken", initAlice, scAddress, ownerAddressBytes, alice)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	for j := 0; j < 2000; j++ {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			tx = vm.CreateTransferTokenTx(aliceNonce, "transferToken", transferOnCalls, scAddress, alice, testAddresses[j*1000+i])

			returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
			require.Nil(t, err)
			require.Equal(t, returnCode, vmcommon.Ok)
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
		tx = vm.CreateTransferTokenTx(aliceNonce, "transferToken", transferOnCalls, scAddress, alice, testAddresses[i])

		returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Equal(t, returnCode, vmcommon.Ok)

		aliceNonce++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process %d ERC20 transfers %s \n", numRun, elapsedTime.String())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)
}

func TestExecuteTransactionAndTimeToProcessChange(t *testing.T) {
	t.Skip("Only a test to benchmark transaction processing")

	testMarshalizer := &marshal.JsonMarshalizer{}
	testHasher := sha256.NewSha256()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	pubkeyConv, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	accnts := integrationtests.CreateInMemoryShardAccountsDB()
	esdtTransferParser, _ := parsers.NewESDTTransferParser(testMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     pubkeyConv,
		ShardCoordinator:    shardCoordinator,
		BuiltInFunctions:    builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: enableEpochsHandler,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	feeHandler := &economicsmocks.EconomicsHandlerMock{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
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
		Accounts:            accnts,
		Hasher:              testHasher,
		PubkeyConv:          pubkeyConv,
		Marshalizer:         testMarshalizer,
		SignMarshalizer:     testMarshalizer,
		ShardCoordinator:    shardCoordinator,
		ScProcessor:         &testscommon.SCProcessorMock{},
		TxFeeHandler:        &testscommon.UnsignedTxHandlerStub{},
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        &economicsmocks.EconomicsHandlerMock{},
		ReceiptForwarder:    &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:      &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:          smartContract.NewArgumentParser(),
		ScrForwarder:        &mock.IntermediateTransactionHandlerMock{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxLogsProcessor:     &mock.TxLogsProcessorStub{},
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
	tx := vm.CreateTransaction(ownerNonce, initAlice, ownerAddressBytes, alice, 1, gasLimit, nil)

	_, err := txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	for j := 0; j < 20; j++ {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			tx = vm.CreateTransaction(aliceNonce, transferOnCalls, alice, testAddresses[j*1000+i], 1, gasLimit, nil)

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
		tx = vm.CreateTransaction(aliceNonce, transferOnCalls, alice, testAddresses[i], 1, gasLimit, nil)

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

	scCode := wasm.GetSCCode("../testdata/erc20-c-03/wrc20_wasm.wasm")

	testContext, err := vm.CreateTxProcessorWasmVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		nil,
		config.EnableEpochs{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)

	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	numAccounts := 100
	testAddresses := createTestAddresses(uint64(numAccounts))
	// ERD Minting
	for _, testAddress := range testAddresses {
		_, _ = vm.CreateAccount(testContext.Accounts, testAddress, 0, big.NewInt(0).Mul(big.NewInt(math.MaxUint64/2), big.NewInt(math.MaxUint64/2)))
	}

	accumulateAddress := createTestAddresses(1)[0]

	// ERC20 Minting
	erc20value := big.NewInt(100)
	for _, testAddress := range testAddresses {
		tx = vm.CreateTransferTokenTx(ownerNonce, "transferToken", erc20value, scAddress, ownerAddressBytes, testAddress)
		ownerNonce++

		returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Equal(t, returnCode, vmcommon.Ok)
	}

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	receiverAddresses := createTestAddresses(uint64(numAccounts))

	transferNonce := uint64(0)
	// Transfer among each person revert and retry
	for i := 0; i < 51; i++ {
		rootHash, _ := testContext.Accounts.RootHash()

		for index, testAddress := range testAddresses {
			tx = vm.CreateTransferTokenTx(transferNonce, "transferToken", erc20value, scAddress, testAddress, receiverAddresses[index])

			snapShot := testContext.Accounts.JournalLen()
			_, _ = testContext.TxProcessor.ProcessTransaction(tx)

			if index%5 == 0 {
				errRevert := testContext.Accounts.RevertToSnapshot(snapShot)
				if errRevert != nil {
					log.Warn("revert to snapshot", "error", errRevert.Error())
				}
			}
		}

		tx = vm.CreateTransferTokenTx(ownerNonce, "transferToken", erc20value, scAddress, ownerAddressBytes, accumulateAddress)
		require.NotNil(t, tx)

		newRootHash, errNewRh := testContext.Accounts.Commit()
		require.Nil(t, errNewRh)

		for index, testAddress := range receiverAddresses {
			if index%5 == 0 {
				continue
			}

			tx = vm.CreateTransferTokenTx(transferNonce, "transferToken", erc20value, scAddress, testAddress, testAddresses[index])

			snapShot := testContext.Accounts.JournalLen()
			_, _ = testContext.TxProcessor.ProcessTransaction(tx)

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
			testContext.Accounts.PruneTrie(extraNewRootHash, state.NewRoot, state.NewPruningHandler(state.EnableDataRemoval))
			_ = testContext.Accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(rootHash))
			continue
		}

		ownerNonce++
		transferNonce++
		testContext.Accounts.PruneTrie(rootHash, state.OldRoot, state.NewPruningHandler(state.EnableDataRemoval))
		testContext.Accounts.PruneTrie(newRootHash, state.OldRoot, state.NewPruningHandler(state.EnableDataRemoval))
	}
}

func TestCommunityContract_InShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	zero := big.NewInt(0)
	transferEGLD := big.NewInt(42)

	net := integrationTests.NewTestNetworkSized(t, 1, 1, 1)
	net.Start()
	net.Step()

	net.CreateWallets(1)
	net.MintWalletsUint64(100000000000)
	owner := net.Wallets[0]

	codePath := "../testdata/community"
	funderCode := codePath + "/funder.wasm"
	parentCode := codePath + "/parent_old.wasm"

	funderAddress := net.DeployPayableSC(owner, funderCode)
	funderSC := net.GetAccountHandler(funderAddress)
	require.Equal(t, owner.Address, funderSC.GetOwnerAddress())

	parentAddress := net.DeploySCWithInitArgs(
		owner,
		parentCode,
		true,
		funderAddress,
	)

	parentSC := net.GetAccountHandler(parentAddress)
	require.Equal(t, owner.Address, parentSC.GetOwnerAddress())

	txData := txDataBuilder.NewBuilder().Func("register").ToBytes()
	tx := net.CreateTx(owner, parentAddress, transferEGLD, txData)
	tx.GasLimit = 1_000_000

	_ = net.SignAndSendTx(owner, tx)

	net.Steps(2)
	funderSC = net.GetAccountHandler(funderAddress)
	require.Equal(t, transferEGLD, funderSC.GetBalance())
	require.Equal(t, zero, parentSC.GetBalance())
}

func TestCommunityContract_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	zero := big.NewInt(0)
	transferEGLD := big.NewInt(42)

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	net.Step()

	net.CreateWallets(2)
	net.MintWalletsUint64(100000000000)
	ownerOfFunder := net.Wallets[0]

	codePath := "../testdata/community"
	funderCode := codePath + "/funder.wasm"
	parentCode := codePath + "/parent_old.wasm"

	funderAddress := net.DeployPayableSC(ownerOfFunder, funderCode)
	funderSC := net.GetAccountHandler(funderAddress)
	require.Equal(t, ownerOfFunder.Address, funderSC.GetOwnerAddress())

	ownerOfParent := net.Wallets[1]
	parentAddress := net.DeploySCWithInitArgs(
		ownerOfParent,
		parentCode,
		true,
		funderAddress,
	)

	parentSC := net.GetAccountHandler(parentAddress)
	require.Equal(t, ownerOfParent.Address, parentSC.GetOwnerAddress())

	txData := txDataBuilder.NewBuilder().Func("register").ToBytes()
	tx := net.CreateTx(ownerOfParent, parentAddress, transferEGLD, txData)
	tx.GasLimit = 1_000_000

	_ = net.SignAndSendTx(ownerOfParent, tx)

	net.Steps(8)
	funderSC = net.GetAccountHandler(funderAddress)
	require.Equal(t, transferEGLD, funderSC.GetBalance())

	parentSC = net.GetAccountHandler(parentAddress)
	require.Equal(t, zero, parentSC.GetBalance())
}

func TestCommunityContract_CrossShard_TxProcessor(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Scenario:
	// 1. Deploy FunderSC on shard 0, owned by funderOwner
	// 2. Deploy ParentSC on shard 1, owned by parentOwner; deployment needs address of FunderSC
	// 3. parentOwner sends tx to ParentSC with method call 'register' and 42 EGLD (in-shard call, shard 1)
	// 4. ParentSC emits a cross-shard asyncCall to FunderSC with method 'acceptFunds' and 42 EGLD
	// 5. assert FunderSC has 42 EGLD
	zero := big.NewInt(0)
	transferEGLD := big.NewInt(42)

	testContextFunderSC, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContextFunderSC.Close()

	testContextParentSC, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContextParentSC.Close()

	funderOwner := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextFunderSC.ShardCoordinator.ComputeId(funderOwner))

	parentOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextParentSC.ShardCoordinator.ComputeId(parentOwner))

	egldBalance := big.NewInt(1000000000000)

	_, _ = vm.CreateAccount(testContextFunderSC.Accounts, funderOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextParentSC.Accounts, parentOwner, 0, egldBalance)

	gasPrice := uint64(10)
	deployGasLimit := uint64(5000000)

	codePath := "../testdata/community"
	funderCode := codePath + "/funder.wasm"
	parentCode := codePath + "/parent_old.wasm"

	logger.ToggleLoggerName(true)

	// Deploy Funder SC in shard 0
	funderOwnerAccount, _ := testContextFunderSC.Accounts.LoadAccount(funderOwner)
	funderAddress := utils.DoDeploySecond(t,
		testContextFunderSC,
		funderCode,
		funderOwnerAccount,
		gasPrice,
		deployGasLimit,
		nil,
		zero)

	// Deploy Parent SC in shard 1
	parentOwnerAccount, _ := testContextParentSC.Accounts.LoadAccount(parentOwner)
	args := [][]byte{[]byte(hex.EncodeToString(funderAddress))}
	parentAddress := utils.DoDeploySecond(t,
		testContextParentSC,
		parentCode,
		parentOwnerAccount,
		gasPrice,
		deployGasLimit,
		args,
		zero)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextFunderSC)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextParentSC)

	// Prepare tx from ParentSC owner to ParentSC (same shard, 1)
	gasLimit := uint64(5000000)
	tx := vm.CreateTransaction(1,
		transferEGLD,
		parentOwner,
		parentAddress,
		gasPrice,
		gasLimit,
		[]byte("register"))

	// execute on the sender shard, which emits an async call
	// from ParentSC (shard 1) to FunderSC (shard 0)
	retCode, err := testContextParentSC.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	intermediateTxs := testContextParentSC.GetIntermediateTransactions(t)
	require.NotEmpty(t, intermediateTxs)

	_, _ = testContextParentSC.Accounts.Commit()
	_, _ = testContextFunderSC.Accounts.Commit()

	utils.TestAccount(t, testContextParentSC.Accounts, parentAddress, 0, big.NewInt(0))

	// execute async call on the FunderSC shard (shard 0)
	scr := intermediateTxs[0]
	require.Equal(t, transferEGLD, scr.GetValue())
	require.Equal(t, parentAddress, scr.GetSndAddr())
	require.Equal(t, funderAddress, scr.GetRcvAddr())
	require.Equal(t, []byte("acceptFunds@168ec815aaa4dfec4de9062e611c8ccc99500101d63962305f8af2d726cc3d04@c7a233a7a0c3889270e967c77aea29871c31740769940739109605ffc4102ddc@01a5c7"), scr.GetData())
	utils.ProcessSCRResult(t, testContextFunderSC, scr, vmcommon.Ok, nil)

	intermediateTxs = testContextFunderSC.GetIntermediateTransactions(t)
	require.NotEmpty(t, intermediateTxs)

	_, _ = testContextParentSC.Accounts.Commit()
	_, _ = testContextFunderSC.Accounts.Commit()

	// return to the ParentSC shard (shard 1)
	scr = intermediateTxs[0]
	utils.ProcessSCRResult(t, testContextParentSC, scr, vmcommon.Ok, nil)

	intermediateTxs = testContextParentSC.GetIntermediateTransactions(t)
	require.NotEmpty(t, intermediateTxs)

	utils.TestAccount(t, testContextParentSC.Accounts, parentAddress, 0, zero)
	utils.TestAccount(t, testContextFunderSC.Accounts, funderAddress, 0, transferEGLD)
}

func TestDeployDNSV2SetDeleteUserNames(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	senderAddressBytes, _ := vm.TestAddressPubkeyConverter.Decode(vm.DNSV2DeployerAddress)
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000)

	scCode := wasm.GetSCCode("../testdata/manage-user-contract.wasm")

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode),
	)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(90000000)

	vm.TestAccount(
		t,
		testContext.Accounts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)

	dnsV2Address, _ := vm.TestAddressPubkeyConverter.Decode(vm.DNSV2Address)
	senderNonce++
	tx = vm.CreateTx(
		senderAddressBytes,
		dnsV2Address,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		"saveName@"+hex.EncodeToString(senderAddressBytes)+"@"+hex.EncodeToString([]byte("userName1")),
	)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	vm.TestAccountUsername(t, testContext.Accounts, senderAddressBytes, []byte("userName1"))

	senderNonce++
	tx = vm.CreateTx(
		senderAddressBytes,
		dnsV2Address,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		"removeName@"+hex.EncodeToString(senderAddressBytes),
	)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	vm.TestAccountUsername(t, testContext.Accounts, senderAddressBytes, nil)

	senderNonce++
	tx = vm.CreateTx(
		senderAddressBytes,
		dnsV2Address,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		"saveName@"+hex.EncodeToString(senderAddressBytes)+"@"+hex.EncodeToString([]byte("userName2")),
	)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	vm.TestAccountUsername(t, testContext.Accounts, senderAddressBytes, []byte("userName2"))

	senderNonce++
	tx = vm.CreateTx(
		senderAddressBytes,
		dnsV2Address,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		"saveName@"+hex.EncodeToString(senderAddressBytes)+"@"+hex.EncodeToString([]byte("userName3")),
	)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	vm.TestAccountUsername(t, testContext.Accounts, senderAddressBytes, []byte("userName3"))
}

type removalVMFixture struct {
	context *vm.VMTestContext
	owner   []byte
	sender  []byte
	address []byte
	token   []byte
}

func newRemovalVMFixture(t *testing.T) *removalVMFixture {
	t.Helper()
	context, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.NoError(t, err)
	t.Cleanup(context.Close)
	fixture := &removalVMFixture{
		context: context,
		owner:   []byte("12345678901234567890123456789011"),
		sender:  []byte("12345678901234567890123456789012"),
		token:   []byte("TKN-abcdef"),
	}
	balance := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	_, err = vm.CreateAccount(context.Accounts, fixture.owner, 0, balance)
	require.NoError(t, err)
	utils.CreateAccountWithESDTBalance(t, context.Accounts, fixture.sender, balance,
		fixture.token, 0, big.NewInt(1000000), uint32(core.Fungible))
	code := wasm.GetSCCode("../testdata/counter/output/counter.wasm")
	transaction := vm.CreateTransaction(0, big.NewInt(0), fixture.owner,
		vm.CreateEmptyAddress(), 10, 100000, []byte(wasm.CreateDeployTxData(code)))
	result, err := context.TxProcessor.ProcessTransaction(transaction)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, result)
	_, err = context.Accounts.Commit()
	require.NoError(t, err)
	fixture.address, err = context.BlockchainHook.NewAddress(fixture.owner, 0, factory.WasmVirtualMachine)
	require.NoError(t, err)
	fixture.requireCounter(t, 1)
	return fixture
}

func (fixture *removalVMFixture) call(t *testing.T, nonce uint64, function string) {
	t.Helper()
	transaction := vm.CreateTransaction(nonce, big.NewInt(0), fixture.owner,
		fixture.address, 10, 100000, []byte(function))
	result, err := fixture.context.TxProcessor.ProcessTransaction(transaction)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, result)
}

func (fixture *removalVMFixture) requireCounter(t *testing.T, expected byte) {
	t.Helper()
	account, err := fixture.context.Accounts.GetExistingAccount(fixture.address)
	require.NoError(t, err)
	key := append([]byte("mycounter"), make([]byte, 23)...)
	value, _, err := account.(state.UserAccountHandler).RetrieveValue(key)
	require.NoError(t, err)
	require.Equal(t, []byte{expected}, value)
}

func TestDeleteContractMiniblockRevertRestoresStorage(t *testing.T) {
	for _, failedRecreation := range []bool{false, true} {
		name := "delete-only"
		if failedRecreation {
			name = "failed-esdt-call-after-delete"
		}
		t.Run(name, func(t *testing.T) {
			fixture := newRemovalVMFixture(t)
			// Earlier work in the block makes this a nonzero miniblock snapshot.
			transaction := vm.CreateTransaction(1, big.NewInt(1), fixture.owner,
				[]byte("12345678901234567890123456789014"), 10, 50000, nil)
			result, err := fixture.context.TxProcessor.ProcessTransaction(transaction)
			require.NoError(t, err)
			require.Equal(t, vmcommon.Ok, result)
			snapshot := fixture.context.Accounts.JournalLen()
			require.Positive(t, snapshot)
			fixture.call(t, 2, "deleteContract")
			if failedRecreation {
				// The built-in credits storage, but the function fails because
				// the deleted contract has no code. Its transaction is reverted.
				transaction = utils.CreateESDTTransferTx(0, fixture.sender, fixture.address,
					fixture.token, big.NewInt(100), 10, 200000)
				transaction.Data = append(transaction.Data, []byte("@"+hex.EncodeToString([]byte("increment")))...)
				result, err = fixture.context.TxProcessor.ProcessTransaction(transaction)
				require.NoError(t, err)
				require.Equal(t, vmcommon.UserError, result)
			}
			// Use the miniblock snapshot to exercise the legacy coordinator's rollback boundary.
			require.NoError(t, fixture.context.Accounts.RevertToSnapshot(snapshot))
			fixture.requireCounter(t, 1) // Assert before C-01's commit cleanup.
			_, err = fixture.context.Accounts.Commit()
			require.NoError(t, err)
			fixture.call(t, 2, "increment")
			fixture.requireCounter(t, 2)
		})
	}
}

func TestDeleteContractMiniblockRevertPersistsEarlierStorage(t *testing.T) {
	fixture := newRemovalVMFixture(t)
	fixture.call(t, 1, "increment")
	fixture.requireCounter(t, 2)
	snapshot := fixture.context.Accounts.JournalLen()
	require.Positive(t, snapshot)
	// The delete needs an on-disk root while rollback needs the earlier dirty trie.
	fixture.call(t, 2, "decrement")
	fixture.call(t, 3, "deleteContract")
	require.NoError(t, fixture.context.Accounts.RevertToSnapshot(snapshot))
	fixture.requireCounter(t, 2)
	root, err := fixture.context.Accounts.Commit()
	require.NoError(t, err)
	require.NoError(t, fixture.context.Accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(root)))
	fixture.requireCounter(t, 2)
	fixture.call(t, 2, "increment")
	fixture.requireCounter(t, 3)
}

func TestLazyDataTrieLoad_WASMExecutionAndQuery(t *testing.T) {
	for _, processor := range []int{1, 2} {
		for _, version := range []int{0, 1} {
			t.Run(fmt.Sprintf("processor-%d/trie-%d", processor, version), func(t *testing.T) {
				epochs := config.EnableEpochs{}
				if processor == 1 {
					epochs.SCProcessorV2EnableEpoch = math.MaxUint32
				}
				if version == 0 {
					epochs.AutoBalanceDataTriesEnableEpoch = math.MaxUint32
				}
				ctx, err := vm.CreatePreparedTxProcessorWithVMs(epochs, 1)
				require.NoError(t, err)
				t.Cleanup(ctx.Close)
				owner := []byte("12345678901234567890123456789011")
				_, err = vm.CreateAccount(ctx.Accounts, owner, 0, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
				require.NoError(t, err)
				code := wasm.GetSCCode("../testdata/counter/output/counter.wasm")
				tx := vm.CreateTransaction(0, big.NewInt(0), owner, vm.CreateEmptyAddress(), 10, 100000, []byte(wasm.CreateDeployTxData(code)))
				result, err := ctx.TxProcessor.ProcessTransaction(tx)
				require.NoError(t, err)
				require.Equal(t, vmcommon.Ok, result)
				root, err := ctx.Accounts.Commit()
				require.NoError(t, err)
				address, err := ctx.BlockchainHook.NewAddress(owner, 0, factory.WasmVirtualMachine)
				require.NoError(t, err)
				acc, err := ctx.Accounts.GetExistingAccount(address)
				require.NoError(t, err)
				dataRoot := acc.(state.UserAccountHandler).GetRootHash()
				tr, err := ctx.Accounts.GetTrie(root)
				require.NoError(t, err)
				storage := tr.GetStorageManager()
				encodedRoot, err := storage.Get(dataRoot)
				require.NoError(t, err)
				vmExec, err := ctx.VMContainer.Get(factory.WasmVirtualMachine)
				require.NoError(t, err)
				input := &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallerAddr: owner, CallValue: big.NewInt(0), GasProvided: 1000000}, RecipientAddr: address, Function: "get"}
				output, err := vmExec.RunSmartContractCall(input)
				require.NoError(t, err)
				require.Equal(t, vmcommon.Ok, output.ReturnCode)
				require.Equal(t, [][]byte{{1}}, output.ReturnData)
				require.NoError(t, scrCommon.GetDataTrieLoadError(vmExec, nil))
				require.NoError(t, storage.Remove(dataRoot))
				require.NoError(t, ctx.Accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(root)))
				// The real VM turns a hook error into ExecutionFailed with a nil Go error.
				output, err = vmExec.RunSmartContractCall(input)
				require.NoError(t, err)
				require.Equal(t, vmcommon.ExecutionFailed, output.ReturnCode)
				loadErr := scrCommon.GetDataTrieLoadError(vmExec, err)
				require.ErrorIs(t, loadErr, state.ErrDataTrieLoadFailed)
				require.Equal(t, dataRoot, core.UnwrapGetNodeFromDBErr(loadErr).GetKey())
				// A read-only method isolates the VM fault from any later SaveAccount error.
				tx = vm.CreateTransaction(1, big.NewInt(0), owner, address, 10, 100000, []byte("get"))
				_, err = ctx.TxProcessor.ProcessTransaction(tx)
				require.ErrorIs(t, err, state.ErrDataTrieLoadFailed)
				require.Equal(t, dataRoot, core.UnwrapGetNodeFromDBErr(err).GetKey())
				require.NoError(t, ctx.Accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(root)))
				queryService, err := smartContract.NewSCQueryService(smartContract.ArgsNewSCQueryService{
					VmContainer: ctx.VMContainer, EconomicsFee: &economicsmocks.EconomicsHandlerMock{},
					BlockChainHook: ctx.BlockchainHook.(process.BlockChainHookWithAccountsAdapter),
					MainBlockChain: &testscommon.ChainHandlerStub{}, APIBlockChain: &testscommon.ChainHandlerStub{},
					WasmVMChangeLocker: &sync.RWMutex{}, Bootstrapper: syncDisabled.NewDisabledBootstrapper(),
					AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(), HistoryRepository: &dblookupext.HistoryRepositoryStub{},
					ShardCoordinator: testscommon.NewMultiShardsCoordinatorMock(1), StorageService: &storageStubs.ChainStorerStub{},
					Marshaller: integrationtests.TestMarshalizer, Hasher: integrationtests.TestHasher, Uint64ByteSliceConverter: uint64ByteSlice.NewBigEndianConverter(),
				})
				require.NoError(t, err)
				query := &process.SCQuery{ScAddress: address, FuncName: "get", CallValue: big.NewInt(0)}
				output, _, err = queryService.ExecuteQuery(query)
				require.ErrorIs(t, err, state.ErrDataTrieLoadFailed)
				require.Nil(t, output)
				// Repair the node, then retry on the same VM to detect stale runtime failures.
				require.NoError(t, storage.Put(dataRoot, encodedRoot))
				output, _, err = queryService.ExecuteQuery(query)
				require.NoError(t, err)
				require.Equal(t, vmcommon.Ok, output.ReturnCode)
				require.Equal(t, [][]byte{{1}}, output.ReturnData)
				result, err = ctx.TxProcessor.ProcessTransaction(tx)
				require.NoError(t, err)
				require.Equal(t, vmcommon.Ok, result)
			})
		}
	}
}
