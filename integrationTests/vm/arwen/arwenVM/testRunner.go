package arwenVM

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process/factory"
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
) (ResultInfo, error) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xfffffffffffffff)

	scCode := arwen.GetSCCode(fileSC)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     big.NewInt(0),
		RcvAddr:   vm.CreateEmptyAddress(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte(arwen.CreateDeployTxData(scCode)),
		Signature: nil,
	}

	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		false,
		vm.ArgEnableEpoch{},
	)
	if err != nil {
		return ResultInfo{}, err
	}

	defer testContext.Close()

	scAddress, err := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)
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
		GasLimit:  gasLimit,
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
		GasUsed:           gasLimit - testContext.GetGasRemaining(),
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
	outOfProcess bool,
) ([]time.Duration, error) {
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(1000000000000000)
	gasPrice := uint64(1)
	transferOnCalls := big.NewInt(5)

	scCode := arwen.GetSCCode(fileName)

	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		outOfProcess,
		vm.ArgEnableEpoch{},
	)
	if err != nil {
		return nil, err
	}
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		300_000_000,
		arwen.CreateDeployTxData(scCode)+"@"+initialSupply,
	)

	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return nil, err
	}
	if testContext.GetLatestError() != nil {
		return nil, testContext.GetLatestError()
	}
	ownerNonce++

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(0).Mul(ownerBalance, ownerBalance))

	bob := []byte("12345678901234567890123456789222")
	_, _ = vm.CreateAccount(testContext.Accounts, bob, 0, big.NewInt(0).Mul(ownerBalance, ownerBalance))

	initAlice := big.NewInt(100000)
	tx = vm.CreateTransferTokenTx(ownerNonce, functionName, initAlice, scAddress, ownerAddressBytes, alice)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return nil, err
	}
	if returnCode != vmcommon.Ok {
		return nil, fmt.Errorf("return code on init is not Ok, but %s", returnCode.String())
	}

	benchmarks := make([]time.Duration, 0, numRun)
	for batch := 0; batch < numRun; batch++ {
		start := time.Now()

		for i := 0; i < numTransferInBatch; i++ {
			tx = vm.CreateTransferTokenTx(aliceNonce, functionName, transferOnCalls, scAddress, alice, bob)

			returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
			if err != nil {
				return nil, err
			}
			if returnCode != vmcommon.Ok {
				return nil, fmt.Errorf("return code on transfer erc20 token is not Ok, but %s", returnCode.String())
			}
			aliceNonce++
		}

		benchmarks = append(benchmarks, time.Since(start))

		_, err = testContext.Accounts.Commit()
		if err != nil {
			return nil, err
		}

		testContext.CreateBlockStarted()
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
