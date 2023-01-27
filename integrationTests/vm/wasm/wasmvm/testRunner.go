package wasmvm

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ResultInfo will hold the result information after running the tests
type ResultInfo struct {
	FunctionName      string
	GasUsed           uint64
	ExecutionTimeSpan time.Duration
}

// RunTest runs a test with the provided parameters
func RunTest(
	fileSC string,
	testingValue uint64,
	function string,
	arguments [][]byte,
	numRun int,
	gasSchedule map[string]map[string]uint64,
	txGasLimit uint64,
) (ResultInfo, error) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xfffffffffffffff)
	if txGasLimit == 0 {
		txGasLimit = uint64(0xfffffffffffffff)
	}
	scCode := wasm.GetSCCode(fileSC)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     big.NewInt(0),
		RcvAddr:   vm.CreateEmptyAddress(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte(wasm.CreateDeployTxData(scCode)),
		Signature: nil,
	}

	testContext, err := vm.CreateTxProcessorWasmVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		config.EnableEpochs{},
	)
	if err != nil {
		return ResultInfo{}, err
	}

	defer testContext.Close()

	scAddress, err := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)
	if err != nil {
		return ResultInfo{}, err
	}

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return ResultInfo{}, err
	}
	if returnCode != vmcommon.Ok {
		return ResultInfo{}, fmt.Errorf("return code is not vmcommon.Ok but %s", returnCode)
	}

	_, err = testContext.Accounts.Commit()
	if err != nil {
		return ResultInfo{}, err
	}

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, err = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(0).Mul(ownerBalance, ownerBalance))
	if err != nil {
		return ResultInfo{}, err
	}

	txData := function
	for _, arg := range arguments {
		txData += "@" + hex.EncodeToString(arg)
	}
	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     new(big.Int).Set(big.NewInt(0).SetUint64(testingValue)),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  1,
		GasLimit:  txGasLimit,
		Data:      []byte(txData),
		Signature: nil,
	}

	startTime := time.Now()
	for i := 0; i < numRun; i++ {
		tx.Nonce = aliceNonce
		_, _ = testContext.TxProcessor.ProcessTransaction(tx)
		aliceNonce++
	}

	return ResultInfo{
		FunctionName:      function,
		GasUsed:           uint64(numRun) * (tx.GasLimit - testContext.GetGasRemaining()),
		ExecutionTimeSpan: time.Since(startTime),
	}, nil
}

// DeployAndExecuteERC20WithBigInt will stress test the erc20 contract
func DeployAndExecuteERC20WithBigInt(
	numRun int,
	numTransferInBatch int,
	gasSchedule map[string]map[string]uint64,
	fileName string,
	functionName string,
) ([]time.Duration, error) {
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(1000000000000000)
	gasPrice := uint64(1)
	transferOnCalls := big.NewInt(5)

	scCode := wasm.GetSCCode(fileName)

	testContext, err := vm.CreateTxProcessorWasmVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		config.EnableEpochs{
			MaxBlockchainHookCountersEnableEpoch: integrationTests.UnreachableEpoch,
		},
	)
	if err != nil {
		return nil, err
	}
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)

	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		300_000_000,
		wasm.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return nil, err
	}
	if returnCode != vmcommon.Ok {
		return nil, fmt.Errorf(returnCode.String())
	}
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(0).Mul(ownerBalance, ownerBalance))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(testContext.Accounts, bob, 0, big.NewInt(0).Mul(ownerBalance, ownerBalance))

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, functionName, initAlice, scAddress, ownerAddressBytes, alice)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return nil, err
	}
	if returnCode != vmcommon.Ok {
		return nil, fmt.Errorf("return code on init is not Ok, but %s", returnCode.String())
	}

	benchmarks, err := runERC20TransactionsWithBenchmarks(
		testContext,
		numRun,
		numTransferInBatch,
		alice,
		bob,
		aliceNonce,
		scAddress,
		functionName,
		transferOnCalls,
	)
	if err != nil {
		return nil, err
	}

	finalAlice := big.NewInt(0).Sub(initAlice, big.NewInt(int64(numRun*numTransferInBatch)*transferOnCalls.Int64()))
	valueFromSc := vm.GetIntValueFromSC(gasSchedule, testContext.Accounts, scAddress, "balanceOf", alice).Uint64()
	if finalAlice.Uint64() != valueFromSc {
		return nil, fmt.Errorf("alice balance mismatch: computed %d, got %d", finalAlice.Uint64(), valueFromSc)
	}

	finalBob := big.NewInt(int64(numRun*numTransferInBatch) * transferOnCalls.Int64())
	valueFromSc = vm.GetIntValueFromSC(gasSchedule, testContext.Accounts, scAddress, "balanceOf", bob).Uint64()
	if finalBob.Uint64() != valueFromSc {
		return nil, fmt.Errorf("bob balance mismatch: computed %d, got %d", finalBob.Uint64(), valueFromSc)
	}

	return benchmarks, nil
}

// SetupERC20Test prepares an ERC20 contract and the accounts to transfer tokens between
func SetupERC20Test(
	testContext *vm.VMTestContext,
	contractCodeFile string,
) error {
	var err error

	largeNumber := big.NewInt(1000000000000000)

	testContext.ContractOwner.Address = []byte("12345678901234567890123456789011")
	testContext.ContractOwner.Nonce = 11
	testContext.ContractOwner.Balance = largeNumber
	testContext.CreateAccount(&testContext.ContractOwner)

	testContext.Contract.Address, err = testContext.BlockchainHook.NewAddress(
		testContext.ContractOwner.Address,
		testContext.ContractOwner.Nonce,
		factory.WasmVirtualMachine,
	)
	if err != nil {
		return err
	}

	scCode := wasm.GetSCCode(contractCodeFile)
	initialSupply := "00" + hex.EncodeToString(largeNumber.Bytes())
	gasPrice := uint64(1)
	tx := vm.CreateDeployTx(
		testContext.ContractOwner.Address,
		testContext.ContractOwner.Nonce,
		big.NewInt(0),
		gasPrice,
		300_000_000,
		wasm.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return err
	}
	if returnCode != vmcommon.Ok {
		return fmt.Errorf(returnCode.String())
	}

	testContext.ContractOwner.Nonce++

	initialAccountBalance := big.NewInt(0).Mul(largeNumber, largeNumber)
	testContext.Alice.Address = []byte("12345678901234567890123456789111")
	testContext.Alice.Nonce = 0
	testContext.Alice.Balance = big.NewInt(0).Set(initialAccountBalance)
	testContext.CreateAccount(&testContext.Alice)

	testContext.Bob.Address = []byte("12345678901234567890123456789222")
	testContext.Bob.Nonce = 0
	testContext.Bob.Balance = big.NewInt(0).Set(initialAccountBalance)
	testContext.CreateAccount(&testContext.Bob)

	testContext.Alice.TokenBalance = big.NewInt(1000000000)
	testContext.Bob.TokenBalance = big.NewInt(0)

	tx = testContext.CreateTransferTokenTx(
		&testContext.ContractOwner,
		&testContext.Alice,
		testContext.Alice.TokenBalance,
		"transferToken",
	)

	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return err
	}
	if returnCode != vmcommon.Ok {
		return fmt.Errorf("return code on init is not Ok, but %s", returnCode.String())
	}

	return nil
}

func runERC20TransactionsWithBenchmarks(
	testContext *vm.VMTestContext,
	numRun int,
	numTransferInBatch int,
	alice []byte,
	bob []byte,
	aliceNonce uint64,
	scAddress []byte,
	functionName string,
	transferOnCalls *big.Int,
) ([]time.Duration, error) {

	benchmarks := make([]time.Duration, 0, numRun)
	for batch := 0; batch < numRun; batch++ {
		start := time.Now()

		for i := 0; i < numTransferInBatch; i++ {
			tx := vm.CreateTransferTokenTx(aliceNonce, functionName, transferOnCalls, scAddress, alice, bob)

			returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
			if err != nil {
				return nil, err
			}
			if returnCode != vmcommon.Ok {
				return nil, fmt.Errorf("return code on transfer erc20 token is not Ok, but %s", returnCode.String())
			}
			aliceNonce++
		}

		benchmarks = append(benchmarks, time.Since(start))

		_, err := testContext.Accounts.Commit()
		if err != nil {
			return nil, err
		}

		testContext.CreateBlockStarted()
	}

	testContext.Alice.Nonce = aliceNonce

	return benchmarks, nil
}

// RunERC20TransactionSet performs a predetermined set of ERC20 token transfers
func RunERC20TransactionSet(testContext *vm.VMTestContext) error {
	_, err := RunERC20TransactionsWithBenchmarksInVMTestContext(
		testContext,
		1,
		100,
		"transferToken",
		big.NewInt(5),
	)

	return err
}

// MakeHeaderHandlerStub prepares a HeaderHandlerStub with the provided epoch
func MakeHeaderHandlerStub(epoch uint32) data.HeaderHandler {
	return &testscommon.HeaderHandlerStub{
		EpochField: epoch,
	}
}

// RunERC20TransactionsWithBenchmarksInVMTestContext executes a configurable set of ERC20 token transfers
func RunERC20TransactionsWithBenchmarksInVMTestContext(
	testContext *vm.VMTestContext,
	numRun int,
	numTransferInBatch int,
	functionName string,
	transferOnCalls *big.Int,
) ([]time.Duration, error) {
	benchmarks, err := runERC20TransactionsWithBenchmarks(
		testContext,
		numRun,
		numTransferInBatch,
		testContext.Alice.Address,
		testContext.Bob.Address,
		testContext.Alice.Nonce,
		testContext.Contract.Address,
		functionName,
		transferOnCalls,
	)
	if err != nil {
		return nil, err
	}

	err = ValidateERC20TransactionsInVMTestContext(
		testContext,
		numRun,
		numTransferInBatch,
		transferOnCalls,
	)
	if err != nil {
		return nil, err
	}

	return benchmarks, nil
}

// ValidateERC20TransactionsInVMTestContext verifies whether the ERC20 transfers were executed correctly
func ValidateERC20TransactionsInVMTestContext(
	testContext *vm.VMTestContext,
	numRun int,
	numTransferInBatch int,
	transferOnCalls *big.Int,
) error {
	initAlice := testContext.Alice.TokenBalance
	finalAlice := big.NewInt(0).Sub(initAlice, big.NewInt(int64(numRun*numTransferInBatch)*transferOnCalls.Int64()))
	valueFromScAlice := testContext.GetIntValueFromSCWithTransientVM("balanceOf", testContext.Alice.Address)
	if finalAlice.Uint64() != valueFromScAlice.Uint64() {
		return fmt.Errorf("alice balance mismatch: computed %d, got %d", finalAlice.Uint64(), valueFromScAlice.Uint64())
	}

	initBob := testContext.Bob.TokenBalance
	finalBob := big.NewInt(0).Add(initBob, big.NewInt(int64(numRun*numTransferInBatch)*transferOnCalls.Int64()))
	valueFromScBob := testContext.GetIntValueFromSCWithTransientVM("balanceOf", testContext.Bob.Address)
	if finalBob.Uint64() != valueFromScBob.Uint64() {
		return fmt.Errorf("bob balance mismatch: computed %d, got %d", finalBob.Uint64(), valueFromScBob.Uint64())
	}

	testContext.Alice.TokenBalance = valueFromScAlice
	testContext.Bob.TokenBalance = valueFromScBob

	return nil
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
