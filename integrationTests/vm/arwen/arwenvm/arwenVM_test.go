// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package arwenvm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
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

	scCode := arwen.GetSCCode("../testdata/misc/fib_arwen/output/fib_arwen.wasm")

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode),
	)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		vm.ArgEnableEpoch{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(99999101)

	vm.TestAccount(
		t,
		testContext.Accounts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)
}

func TestVmSCDeployFactory(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000)
	transferOnCalls := big.NewInt(0)

	scCode := arwen.GetSCCode("../testdata/misc/deploy-two-contracts.wasm")

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode),
	)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		vm.ArgEnableEpoch{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.ArwenVirtualMachine)
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

func TestSCMoveBalanceBeforeSCDeploy(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode := arwen.GetSCCode("../testdata/misc/fib_arwen/output/fib_arwen.wasm")

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		vm.ArgEnableEpoch{
			PenalizedTooMuchGasEnableEpoch: 100,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce+1, factory.ArwenVirtualMachine)
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
	require.Equal(t, process.ErrAccountNotPayable, testContext.GetLatestError())
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
		arwen.CreateDeployTxData(scCode),
	)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	vm.TestAccount(
		t,
		testContext.Accounts,
		ownerAddressBytes,
		ownerNonce+2,
		big.NewInt(99999100))

	vm.TestAccount(
		t,
		testContext.Accounts,
		scAddressBytes,
		0,
		transferOnCalls)
}

func TestWASMMetering(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xfffffffffffffff)
	transferOnCalls := big.NewInt(1)

	scCode := arwen.GetSCCode("../testdata/misc/cpucalculate_arwen/output/cpucalculate.wasm")

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

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		vm.ArgEnableEpoch{
			PenalizedTooMuchGasEnableEpoch: 100,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

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

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	expectedBalance := big.NewInt(2998080)
	expectedNonce := uint64(1)

	actualBalanceBigInt := vm.TestAccount(
		t,
		testContext.Accounts,
		alice,
		expectedNonce,
		expectedBalance)

	actualBalance := actualBalanceBigInt.Uint64()

	consumedGasValue := aliceInitialBalance - actualBalance - testingValue

	require.Equal(t, 1905, int(consumedGasValue))
}

func TestMultipleTimesERC20BigIntInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	durations, err := DeployAndExecuteERC20WithBigInt(3, 1000, gasSchedule, "../testdata/erc20-c-03/wrc20_arwen.wasm", "transferToken")
	require.Nil(t, err)
	displayBenchmarksResults(durations)

	durations, err = DeployAndExecuteERC20WithBigInt(3, 1000, nil, "../testdata/erc20-c-03/wrc20_arwen.wasm", "transferToken")
	require.Nil(t, err)
	displayBenchmarksResults(durations)
}

func TestMultipleTimesERC20RustBigIntInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
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
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(1000000000000000)
	gasPrice := uint64(1)

	scCode := arwen.GetSCCode("../testdata/erc20-c-03/wrc20_arwen.wasm")

	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		vm.ArgEnableEpoch{},
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
		arwen.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, "out of gas", testContext.GetLatestError().Error())
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

	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		nil,
		vm.ArgEnableEpoch{},
	)
	require.Nil(t, err)
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

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
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
	tx = vm.CreateTransferTokenTx(ownerNonce, "transferToken", initAlice, scAddress, ownerAddressBytes, alice)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	for j := 0; j < 2000; j++ {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			tx = vm.CreateTransferTokenTx(aliceNonce, "transferToken", transferOnCalls, scAddress, alice, testAddresses[j*1000+i])

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
		tx = vm.CreateTransferTokenTx(aliceNonce, "transferToken", transferOnCalls, scAddress, alice, testAddresses[i])

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
	testHasher := sha256.NewSha256()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	pubkeyConv, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	accnts := vm.CreateInMemoryShardAccountsDB()
	esdtTransferParser, _ := parsers.NewESDTTransferParser(testMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubkeyConv,
		ShardCoordinator:   shardCoordinator,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		EpochNotifier:      forking.NewGenericEpochNotifier(),
		ESDTTransferParser: esdtTransferParser,
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
		ScProcessor:      &testscommon.SCProcessorMock{},
		TxFeeHandler:     &testscommon.UnsignedTxHandlerStub{},
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

	scCode := arwen.GetSCCode("../testdata/erc20-c-03/wrc20_arwen.wasm")

	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		nil,
		vm.ArgEnableEpoch{},
	)
	require.Nil(t, err)
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

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
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
		tx = vm.CreateTransferTokenTx(ownerNonce, "transferToken", erc20value, scAddress, ownerAddressBytes, testAddress)
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
			tx = vm.CreateTransferTokenTx(transferNonce, "transferToken", erc20value, scAddress, testAddress, receiverAddresses[index])

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
			testContext.Accounts.PruneTrie(extraNewRootHash, temporary.NewRoot)
			_ = testContext.Accounts.RecreateTrie(rootHash)
			continue
		}

		ownerNonce++
		transferNonce++
		testContext.Accounts.PruneTrie(rootHash, temporary.OldRoot)
		testContext.Accounts.PruneTrie(newRootHash, temporary.OldRoot)
	}
}
